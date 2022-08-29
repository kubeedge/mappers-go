/*
Copyright 2021 The KubeEdge Authors.

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
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/currantlabs/ble"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
	bledriver "github.com/kubeedge/mappers-go/pkg/driver/ble"
	"github.com/kubeedge/mappers-go/pkg/global"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

var deviceMuxs map[string]context.CancelFunc
var devices map[string]*bledriver.BleDev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitor property is readonly, if not then set it.
func setVisitor(visitorConfig *bledriver.BleVisitorConfig, twin *common.Twin, bleClient *bledriver.BleClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(1).Info("Visit readonly characteristicUUID: ", visitorConfig.CharacteristicUUID)
		return
	}

	klog.V(2).Infof("Convert type: %s, value: %s ", twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	value, err := common.Convert(twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	if err != nil {
		klog.Errorf("Convert error: %v", err)
		return
	}

	valueString, _ := value.(string)
	err = bleClient.Set(ble.MustParse(visitorConfig.CharacteristicUUID), []byte(valueString))
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

	var dev *bledriver.BleDev
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
		var visitorConfig bledriver.BleVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.BleClient)
	}
}

// initBLE initialize ble client
func initBle(protocolConfig bledriver.BleProtocolConfig, name string) (client bledriver.BleClient, err error) {
	if protocolConfig.MacAddress != "" {
		config := bledriver.BleConfig{
			Addr: protocolConfig.MacAddress,
		}
		client, err = bledriver.NewClient(config)
	}
	return
}

// initTwin initialize the timer to get twin value.
func initTwin(ctx context.Context, dev *bledriver.BleDev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig bledriver.BleVisitorConfig
		if err := json.Unmarshal(dev.Instance.Twins[i].PVisitor.VisitorConfig, &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.BleClient)

		twinData := TwinData{
			BleClient:        dev.BleClient,
			Name:             dev.Instance.Twins[i].PropertyName,
			Type:             dev.Instance.Twins[i].Desired.Metadatas.Type,
			BleVisitorConfig: visitorConfig,
			Topic:            fmt.Sprintf(common.TopicTwinUpdate, dev.Instance.ID)}
		collectCycle := time.Duration(dev.Instance.Twins[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		uuid := ble.MustParse(twinData.BleVisitorConfig.CharacteristicUUID)
		if p, err := twinData.BleClient.Client.DiscoverProfile(true); err == nil {
			if twinData.FindedCharacteristic = p.Find(ble.NewCharacteristic(uuid)); twinData.FindedCharacteristic == nil {
				klog.Errorf("can't find uuid %s", uuid.String())
				continue
			}
			c := twinData.FindedCharacteristic.(*ble.Characteristic)
			// If this Characteristic supports notifications and there's a CCCD
			// Then subscribe to it, the notifications operation is different from reading operation, notifications will keep looping when connected
			// so we can't use timer.Start() for notifications
			if (c.Property&ble.CharNotify) != 0 && c.CCCD != nil {
				wg.Add(1)
				go func() {
					if err := twinData.BleClient.Client.Subscribe(c, false, twinData.notificationHandler()); err != nil {
						klog.Errorf("Subscribe error: %v", err)
						wg.Done()
					}
				}()
			} else if (c.Property & ble.CharRead) != 0 { // // read data actively
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
	}
}

// initData initialize the timer to get data.
func initData(ctx context.Context, dev *bledriver.BleDev) {
	for i := 0; i < len(dev.Instance.Datas.Properties); i++ {
		var visitorConfig bledriver.BleVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}

		twinData := TwinData{
			BleClient:        dev.BleClient,
			Name:             dev.Instance.Datas.Properties[i].PropertyName,
			Type:             dev.Instance.Datas.Properties[i].Metadatas.Type,
			BleVisitorConfig: visitorConfig,
			Topic:            fmt.Sprintf(common.TopicDataUpdate, dev.Instance.ID)}
		collectCycle := time.Duration(dev.Instance.Datas.Properties[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		uuid := ble.MustParse(twinData.BleVisitorConfig.CharacteristicUUID)
		if p, err := twinData.BleClient.Client.DiscoverProfile(true); err == nil {
			if u := p.Find(ble.NewCharacteristic(uuid)); u == nil {
				klog.Errorf("can't find uuid %s", uuid.String())
				continue
			}
			ticker := time.NewTicker(collectCycle)
			select {
			case <-ticker.C:
				twinData.Run()
			case <-ctx.Done():
				return
			}
			// If this Characteristic supports notifications and there's a CCCD
			// Then subscribe to it, the notifications operation is different from reading operation, notifications will keep looping when connected
			// so we can't use timer.Start() for notifications
			// TODO why add() twice but done() once?
			//wg.Add(1)
			go func() {
				c := twinData.FindedCharacteristic.(*ble.Characteristic)
				if (c.Property&ble.CharNotify) != 0 && c.CCCD != nil {
					if err := twinData.BleClient.Client.Subscribe(c, false, twinData.notificationHandler()); err != nil {
						klog.Errorf("Subscribe failed: %v", err)
					}
				}
			}()
			//wg.Add(1)
			//go func() {
			//	defer wg.Done()
			//	timer.Start()
			//}()
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
func initGetStatus(ctx context.Context, dev *bledriver.BleDev) {
	getStatus := GetStatus{Client: dev.BleClient,
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
func start(ctx context.Context, dev *bledriver.BleDev) {
	var protocolConfig bledriver.BleProtocolConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolConfigs, &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfig error: %v", err)
		wg.Done()
		return
	}

	client, err := initBle(protocolConfig, protocolConfig.MacAddress)
	if err != nil {
		klog.Errorf("Init error: %v", err)
		wg.Done()
		return
	}
	dev.BleClient = &client

	go initTwin(ctx, dev)
	go initData(ctx, dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Errorf("Init subscribe mqtt error: %v", err)
		wg.Done()
		return
	}

	go initGetStatus(ctx, dev)

	<-ctx.Done()
	wg.Done()
}

// DevInit initialize the device data.
func DevInit(cfg *config.Config) error {
	devices = make(map[string]*bledriver.BleDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)
	deviceMuxs = make(map[string]context.CancelFunc)
	devs := make(map[string]*common.DeviceInstance)

	switch cfg.DevInit.Mode {
	case common.DevInitModeRegister:
		if err := parse.ParseByUsingRegister(cfg, devs, models, protocols); err != nil {
			return err
		}
	case common.DevInitModeConfigmap:
		if err := parse.Parse(cfg.DevInit.Configmap, devs, models, protocols); err != nil {
			return err
		}
	case common.DevInitModeMetaServer:
		if err := parse.ParseByUsingMetaServer(cfg, devs, models, protocols); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported dev init mode %s", cfg.DevInit.Mode)
	}

	for key, deviceInstance := range devs {
		cur := new(bledriver.BleDev)
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
		wg.Add(1)
		go start(ctx, dev)
	}
	wg.Wait()
}

func UpdateDev(model *common.DeviceModel, device *common.DeviceInstance, protocol *common.Protocol) {
	devices[device.ID] = new(bledriver.BleDev)
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
	wg.Add(1)
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

func DealDeviceTwinGet(deviceID string, twinName string) (interface{}, error) {
	srcDev, ok := devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("not found device %s", deviceID)
	}

	res := make([]parse.TwinResultResponse, 0)
	for _, twin := range srcDev.Instance.Twins {
		if twinName != "" && twin.PropertyName != twinName {
			continue
		}
		payload, err := getTwinData(deviceID, twin, srcDev.BleClient)
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

func getTwinData(deviceID string, twin common.Twin, client *bledriver.BleClient) ([]byte, error) {
	var visitorConfig bledriver.BleVisitorConfig
	if err := json.Unmarshal(twin.PVisitor.VisitorConfig, &visitorConfig); err != nil {
		return nil, fmt.Errorf("unmarshal VisitorConfig error: %v", err)
	}
	setVisitor(&visitorConfig, &twin, client)

	td := TwinData{
		BleClient:        client,
		Name:             twin.PropertyName,
		Type:             twin.Desired.Metadatas.Type,
		BleVisitorConfig: visitorConfig,
		Topic:            fmt.Sprintf(common.TopicTwinUpdate, deviceID),
	}
	uuid := ble.MustParse(visitorConfig.CharacteristicUUID)
	p, err := client.Client.DiscoverProfile(true)
	if err != nil {
		return nil, err
	}
	if td.FindedCharacteristic = p.Find(ble.NewCharacteristic(uuid)); td.FindedCharacteristic == nil {
		return nil, fmt.Errorf("can't find uuid %s", uuid.String())
	}
	c := td.FindedCharacteristic.(*ble.Characteristic)
	// read data actively
	b, err := td.BleClient.Read(c)
	if err != nil {
		klog.Errorf("Failed to read characteristic: %s\n", err)
	}

	return []byte(fmt.Sprintf("%f", td.ConvertReadData(b))), nil
}
