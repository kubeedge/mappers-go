/*
Copyright 2020 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/modbus"
	"github.com/kubeedge/mappers-go/pkg/global"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

var deviceMuxs map[string]context.CancelFunc
var devices map[string]*modbus.ModbusDev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitor property is readonly, if not then set it.
func setVisitor(visitorConfig *modbus.ModbusVisitorConfig, twin *common.Twin, client *modbus.ModbusClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(1).Info("Visit readonly register: ", visitorConfig.Offset)
		return
	}

	klog.V(2).Infof("Convert type: %s, value: %s ", twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	value, err := common.Convert(twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	if err != nil {
		klog.Errorf("Convert error: %v", err)
		return
	}

	valueInt, _ := value.(int64)
	_, err = client.Set(visitorConfig.Register, visitorConfig.Offset, uint16(valueInt))
	if err != nil {
		klog.Errorf("Set visitor error: %v %v", err, visitorConfig)
		return
	}
}

// getDeviceID extract the device ID from Mqtt topic.
func getDeviceID(topic string) (id string) {
	re := regexp.MustCompile(`hw/events/device/(.+)/twin/update/delta`)
	return re.FindStringSubmatch(topic)[1]
}

// onMessage callback function of Mqtt subscribe message.
func onMessage(client mqtt.Client, message mqtt.Message) {
	klog.V(2).Info("Receive message", message.Topic())
	// Get device ID and get device instance
	id := getDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Device id: ", id)

	var dev *modbus.ModbusDev
	var ok bool
	if dev, ok = devices[id]; !ok {
		klog.Error("Device not exist")
		return
	}

	// Get twin map key as the propertyName
	var delta common.DeviceTwinDelta
	if err := json.Unmarshal(message.Payload(), &delta); err != nil {
		klog.Errorf("Unmarshal message failed: %v", err)
		return
	}
	for twinName, twinValue := range delta.Delta {
		i := 0
		for i = 0; i < len(dev.Instance.Twins); i++ {
			if twinName == dev.Instance.Twins[i].PropertyName {
				break
			}
		}
		if i == len(dev.Instance.Twins) {
			klog.Error("Twin not found: ", twinName)
			continue
		}
		// Desired value is not changed.
		if dev.Instance.Twins[i].Desired.Value == twinValue {
			continue
		}
		dev.Instance.Twins[i].Desired.Value = twinValue
		var visitorConfig modbus.ModbusVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal visitor config failed: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.ModbusClient)
	}
}

// isRS485Enabled is RS485 feature enabled for RTU.
func isRS485Enabled(customizedValue modbus.CustomizedValue) bool {
	isEnabled := false

	if len(customizedValue) != 0 {
		if value, ok := customizedValue["serialType"]; ok {
			if value == "RS485" {
				isEnabled = true
			}
		}
	}
	return isEnabled
}

// initModbus initialize modbus client
func initModbus(protocolConfig modbus.ModbusProtocolCommonConfig, slaveID int16) (client *modbus.ModbusClient, err error) {
	if protocolConfig.COM.SerialPort != "" {
		modbusRTU := modbus.ModbusRTU{SlaveID: byte(slaveID),
			SerialName:   protocolConfig.COM.SerialPort,
			BaudRate:     int(protocolConfig.COM.BaudRate),
			DataBits:     int(protocolConfig.COM.DataBits),
			StopBits:     int(protocolConfig.COM.StopBits),
			Parity:       protocolConfig.COM.Parity,
			RS485Enabled: isRS485Enabled(protocolConfig.CustomizedValues),
			Timeout:      5 * time.Second}
		client, _ = modbus.NewClient(modbusRTU)
	} else if protocolConfig.TCP.IP != "" {
		modbusTCP := modbus.ModbusTCP{
			SlaveID:  byte(slaveID),
			DeviceIP: protocolConfig.TCP.IP,
			TCPPort:  strconv.FormatInt(protocolConfig.TCP.Port, 10),
			Timeout:  5 * time.Second}
		client, _ = modbus.NewClient(modbusTCP)
	} else {
		return nil, errors.New("No protocol found")
	}
	return client, nil
}

// initTwin initialize the timer to get twin value.
func initTwin(ctx context.Context, dev *modbus.ModbusDev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig modbus.ModbusVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.ModbusClient)

		twinData := TwinData{Client: dev.ModbusClient,
			Name:          dev.Instance.Twins[i].PropertyName,
			Type:          dev.Instance.Twins[i].Desired.Metadatas.Type,
			VisitorConfig: &visitorConfig,
			Topic:         fmt.Sprintf(common.TopicTwinUpdate, dev.Instance.ID)}
		collectCycle := time.Duration(dev.Instance.Twins[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		ticker := time.NewTicker(collectCycle)
		select {
		case <-ticker.C:
			twinData.Run()
		case <-ctx.Done():
			return
		}
	}
}

// initData initialize the timer to get data.
func initData(ctx context.Context, dev *modbus.ModbusDev) {
	for i := 0; i < len(dev.Instance.Datas.Properties); i++ {
		var visitorConfig modbus.ModbusVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Datas.Properties[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		twinData := TwinData{Client: dev.ModbusClient,
			Name:          dev.Instance.Datas.Properties[i].PropertyName,
			Type:          dev.Instance.Datas.Properties[i].Metadatas.Type,
			VisitorConfig: &visitorConfig,
			Topic:         fmt.Sprintf(common.TopicDataUpdate, dev.Instance.ID)}
		collectCycle := time.Duration(dev.Instance.Datas.Properties[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		ticker := time.NewTicker(collectCycle)
		select {
		case <-ticker.C:
			twinData.Run()
		case <-ctx.Done():
			return
		}
	}
}

// initSubscribeMqtt subscribe Mqtt topics.
func initSubscribeMqtt(instanceID string) error {
	topic := fmt.Sprintf(common.TopicTwinUpdateDelta, instanceID)
	klog.V(1).Info("Subscribe topic: ", topic)
	return global.MqttClient.Subscribe(topic, onMessage)
}

// initGetStatus start timer to get device status and send to eventbus.
func initGetStatus(ctx context.Context, dev *modbus.ModbusDev) {
	getStatus := GetStatus{Client: dev.ModbusClient,
		topic: fmt.Sprintf(common.TopicStateUpdate, dev.Instance.ID)}
	ticker := time.NewTicker(time.Second)
	select {
	case <-ticker.C:
		getStatus.Run()
	case <-ctx.Done():
		return
	}
}

// start the device.
func start(ctx context.Context, dev *modbus.ModbusDev) {
	var protocolCommConfig modbus.ModbusProtocolCommonConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolCommonConfig, &protocolCommConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolCommonConfig error: %v", err)
		return
	}

	var protocolConfig modbus.ModbusProtocolConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolConfigs, &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfigs error: %v", err)
		return
	}

	client, err := initModbus(protocolCommConfig, protocolConfig.SlaveID)
	if err != nil {
		klog.Errorf("Init error: %v", err)
		return
	}
	dev.ModbusClient = client

	go initTwin(ctx, dev)
	go initData(ctx, dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Errorf("Init subscribe mqtt error: %v", err)
		return
	}

	go initGetStatus(ctx, dev)
	wg.Add(1)
	select {
	case <-ctx.Done():
		wg.Done()
	}
}

// DevInit initialize the device data.
func DevInit(cfg *config.Config) error {
	devices = make(map[string]*modbus.ModbusDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)
	deviceMuxs = make(map[string]context.CancelFunc)
	devs := make(map[string]*common.DeviceInstance)

	if cfg.MetaServer.Enable {
		if err := parse.ParseByUsingMetaServer(cfg, devs, models, protocols); err != nil {
			return err
		}
	} else {
		if err := parse.Parse(cfg.Configmap, devs, models, protocols); err != nil {
			return err
		}
	}

	for key, deviceInstance := range devs {
		cur := new(modbus.ModbusDev)
		cur.Instance = *deviceInstance
		devices[key] = cur
	}
	return nil
}

// DevStart start all devices.
func DevStart() {
	for id, dev := range devices {
		klog.V(4).Info("Dev: ", id, dev)
		ctx, cancel := context.WithCancel(context.Background())
		deviceMuxs[id] = cancel
		go start(ctx, dev)
	}
	wg.Wait()
}

func UpdateDev(model *common.DeviceModel, device *common.DeviceInstance, protocol *common.Protocol) {
	devices[device.ID] = new(modbus.ModbusDev)
	devices[device.ID].Instance = *device
	models[device.Model] = *model
	protocols[device.ProtocolName] = *protocol

	// stop old
	if err := stopDev(device.ID); err != nil {
		klog.Error(err)
	}
	// start new
	ctx, cancelFunc := context.WithCancel(context.Background())
	deviceMuxs[device.ID] = cancelFunc
	go start(ctx, devices[device.ID])
}

func stopDev(id string) error {
	cancelFunc, ok := deviceMuxs[id]
	if !ok {
		return fmt.Errorf("can not find device %s from device muxs", id)
	}
	cancelFunc()
	return nil
}
