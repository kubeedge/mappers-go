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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/currantlabs/ble"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mappers/ble/configmap"
	"github.com/kubeedge/mappers-go/mappers/ble/driver"
	"github.com/kubeedge/mappers-go/mappers/ble/globals"
	"github.com/kubeedge/mappers-go/mappers/common"
)

var devices map[string]*globals.BleDev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitor property is readonly, if not then set it.
func setVisitor(visitorConfig *configmap.BleVisitorConfig, twin *common.Twin, bleClient *driver.BleClient) {
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

// onMessage callback function of Mqtt subscribe message.
func onMessage(client mqtt.Client, message mqtt.Message) {
	klog.V(2).Info("Receive message", message.Topic())
	// Get device ID and get device instance
	id := common.GetDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Device id: ", id)

	var dev *globals.BleDev
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
		var visitorConfig configmap.BleVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.BleClient)
	}
}

// initBLE initialize ble client
func initBle(protocolConfig configmap.BleProtocolConfig, name string) (client driver.BleClient, err error) {
	if protocolConfig.MacAddress != "" {
		config := driver.BleConfig{
			Addr: protocolConfig.MacAddress,
		}
		client, err = driver.NewClient(config)
	}
	return
}

// initTwin initialize the timer to get twin value.
func initTwin(dev *globals.BleDev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig configmap.BleVisitorConfig
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
				timer := common.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
				wg.Add(1)
				go func() {
					defer wg.Done()
					timer.Start()
				}()
			}
		}
	}
}

// initData initialize the timer to get data.
func initData(dev *globals.BleDev) {
	for i := 0; i < len(dev.Instance.Datas.Properties); i++ {
		var visitorConfig configmap.BleVisitorConfig
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
			timer := common.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
			// If this Characteristic supports notifications and there's a CCCD
			// Then subscribe to it, the notifications operation is different from reading operation, notifications will keep looping when connected
			// so we can't use timer.Start() for notifications
			wg.Add(1)
			go func() {
				c := twinData.FindedCharacteristic.(*ble.Characteristic)
				if (c.Property&ble.CharNotify) != 0 && c.CCCD != nil {
					if err := twinData.BleClient.Client.Subscribe(c, false, twinData.notificationHandler()); err != nil {
						klog.Errorf("Subscribe failed: %v", err)
					}
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				timer.Start()
			}()
		}
	}
}

// initSubscribeMqtt subscribe Mqtt topics.
func initSubscribeMqtt(instanceID string) error {
	topic := fmt.Sprintf(common.TopicTwinUpdateDelta, instanceID)
	klog.V(1).Info("Subscribe topic: ", topic)
	return globals.MqttClient.Subscribe(topic, onMessage)
}

// initGetStatus start timer to get device status and send to eventbus.
func initGetStatus(dev *globals.BleDev) {
	getStatus := GetStatus{Client: dev.BleClient,
		topic: fmt.Sprintf(common.TopicStateUpdate, dev.Instance.ID)}
	timer := common.Timer{Function: getStatus.Run, Duration: 1 * time.Second, Times: 0}
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer.Start()
	}()
}

// start start the device.
func start(dev *globals.BleDev) {
	var protocolConfig configmap.BleProtocolConfig
	if err := json.Unmarshal([]byte(dev.Instance.PProtocol.ProtocolConfigs), &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfig error: %v", err)
		return
	}

	client, err := initBle(protocolConfig, protocolConfig.MacAddress)
	if err != nil {
		klog.Errorf("Init error: %v", err)
		return
	}
	dev.BleClient = &client

	initTwin(dev)
	initData(dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Errorf("Init subscribe mqtt error: %v", err)
		return
	}

	initGetStatus(dev)
}

// DevInit initialize the device datas.
func DevInit(configmapPath string) error {
	devices = make(map[string]*globals.BleDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)
	return configmap.Parse(configmapPath, devices, models, protocols)
}

// DevStart start all devices.
func DevStart() {
	for id, dev := range devices {
		klog.V(4).Info("Dev: ", id, dev)
		start(dev)
	}
	wg.Wait()
}
