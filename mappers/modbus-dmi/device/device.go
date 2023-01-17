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

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/modbus"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

type DevPanel struct {
	deviceMuxs map[string]context.CancelFunc
	devices    map[string]*modbus.ModbusDev
	models     map[string]common.DeviceModel
	protocols  map[string]common.Protocol
	wg         sync.WaitGroup
}

var devPanel *DevPanel

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
		if err := json.Unmarshal(dev.Instance.Twins[i].PVisitor.VisitorConfig, &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.ModbusClient)

		twinData := TwinData{Client: dev.ModbusClient,
			Name:          dev.Instance.Twins[i].PropertyName,
			Type:          dev.Instance.Twins[i].Desired.Metadatas.Type,
			VisitorConfig: &visitorConfig,
			Topic:         fmt.Sprintf(common.TopicTwinUpdate, dev.Instance.ID),
			DeviceName:    dev.Instance.Name}
		collectCycle := time.Duration(dev.Instance.Twins[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		ticker := time.NewTicker(collectCycle)
		go func() {
			for {
				select {
				case <-ticker.C:
					twinData.Run()
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// start the device.
func (d *DevPanel) start(ctx context.Context, dev *modbus.ModbusDev) {
	var protocolCommConfig modbus.ModbusProtocolCommonConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolCommonConfig, &protocolCommConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolCommonConfig error: %v", err)
		d.wg.Done()
		return
	}

	var protocolConfig modbus.ModbusProtocolConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolConfigs, &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfigs error: %v", err)
		d.wg.Done()
		return
	}

	client, err := initModbus(protocolCommConfig, protocolConfig.SlaveID)
	if err != nil {
		klog.Errorf("Init dev %s error: %v", dev.Instance.Name, err)
		d.wg.Done()
		return
	}
	dev.ModbusClient = client

	go initTwin(ctx, dev)

	<-ctx.Done()
	d.wg.Done()
}

// DevInit initialize the device data.
func (d *DevPanel) DevInit(cfg *config.Config) error {
	devs := make(map[string]*common.DeviceInstance)

	switch cfg.DevInit.Mode {
	case common.DevInitModeConfigmap:
		if err := parse.Parse(cfg.DevInit.Configmap, devs, d.models, d.protocols); err != nil {
			return err
		}
	case common.DevInitModeRegister:
		if err := parse.ParseByUsingRegister(cfg, devs, d.models, d.protocols); err != nil {
			return err
		}
	}

	for key, deviceInstance := range devs {
		cur := new(modbus.ModbusDev)
		cur.Instance = *deviceInstance
		d.devices[key] = cur
	}
	return nil
}

func NewDevPanel() *DevPanel {
	if devPanel == nil {
		devPanel = &DevPanel{
			deviceMuxs: make(map[string]context.CancelFunc),
			devices:    make(map[string]*modbus.ModbusDev),
			models:     make(map[string]common.DeviceModel),
			protocols:  make(map[string]common.Protocol),
			wg:         sync.WaitGroup{},
		}
	}
	return devPanel
}

// DevStart start all devices.
func (d *DevPanel) DevStart() {
	for id, dev := range d.devices {
		klog.V(4).Info("Dev: ", id, dev)
		ctx, cancel := context.WithCancel(context.Background())
		d.deviceMuxs[id] = cancel
		d.wg.Add(1)
		go d.start(ctx, dev)
	}
	d.wg.Wait()
}

func (d *DevPanel) UpdateDevTwins(deviceID string, twins []common.Twin) error {
	device, ok := d.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}

	device.Instance.Twins = twins
	model := d.models[device.Instance.Model]
	protocol := d.protocols[device.Instance.ProtocolName]
	d.UpdateDev(&model, &device.Instance, &protocol)
	return nil
}

func (d *DevPanel) UpdateDev(model *common.DeviceModel, device *common.DeviceInstance, protocol *common.Protocol) {
	d.devices[device.ID] = new(modbus.ModbusDev)
	d.devices[device.ID].Instance = *device
	d.models[device.Model] = *model
	d.protocols[device.ProtocolName] = *protocol

	// stop old
	if err := d.stopDev(device.ID); err != nil {
		klog.Error(err)
	}
	// start new
	ctx, cancelFunc := context.WithCancel(context.Background())
	d.deviceMuxs[device.ID] = cancelFunc
	d.wg.Add(1)
	go d.start(ctx, d.devices[device.ID])
}

func (d *DevPanel) stopDev(id string) error {
	cancelFunc, ok := d.deviceMuxs[id]
	if !ok {
		return fmt.Errorf("can not find device %s from device muxs", id)
	}
	cancelFunc()
	return nil
}

func (d *DevPanel) DealDeviceTwinGet(deviceID string, twinName string) (interface{}, error) {
	srcDev, ok := d.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("not found device %s", deviceID)
	}

	res := make([]parse.TwinResultResponse, 0)
	for _, twin := range srcDev.Instance.Twins {
		if twinName != "" && twin.PropertyName != twinName {
			continue
		}
		payload, err := getTwinData(deviceID, twin, srcDev.ModbusClient)
		if err != nil {
			return nil, err
		}
		cur := parse.TwinResultResponse{
			PropertyName: twin.PropertyName,
			Payload:      payload,
		}
		res = append(res, cur)
	}
	return json.Marshal(res)
}

func getTwinData(deviceID string, twin common.Twin, client *modbus.ModbusClient) ([]byte, error) {
	var visitorConfig modbus.ModbusVisitorConfig
	if err := json.Unmarshal(twin.PVisitor.VisitorConfig, &visitorConfig); err != nil {
		return nil, fmt.Errorf("unmarshal VisitorConfig error: %v", err)
	}
	setVisitor(&visitorConfig, &twin, client)

	td := TwinData{
		Client:        client,
		Name:          twin.PropertyName,
		Type:          twin.Desired.Metadatas.Type,
		VisitorConfig: &visitorConfig,
		Topic:         fmt.Sprintf(common.TopicTwinUpdate, deviceID),
	}
	return td.GetPayload()
}

func (d *DevPanel) GetDevice(deviceID string) (interface{}, error) {
	found, ok := d.devices[deviceID]
	if !ok || found == nil {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	// get the latest reported twin value
	for i, twin := range found.Instance.Twins {
		payload, err := getTwinData(deviceID, twin, found.ModbusClient)
		if err != nil {
			return nil, err
		}
		found.Instance.Twins[i].Reported.Value = string(payload)
	}
	return found, nil
}

func (d *DevPanel) RemoveDevice(deviceID string) error {
	delete(d.devices, deviceID)
	return d.stopDev(deviceID)
}

func (d *DevPanel) GetModel(modelName string) (common.DeviceModel, error) {
	found, ok := d.models[modelName]
	if !ok {
		return common.DeviceModel{}, fmt.Errorf("deviceModel %s not found", modelName)
	}

	return found, nil
}

func (d *DevPanel) UpdateModel(model *common.DeviceModel) {
	d.models[model.Name] = *model
}

func (d *DevPanel) RemoveModel(modelName string) {
	delete(d.models, modelName)
}
