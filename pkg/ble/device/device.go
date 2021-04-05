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
	"github.com/currantlabs/ble"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kubeedge/mappers-go/pkg/ble/configmap"
	"github.com/kubeedge/mappers-go/pkg/ble/driver"
	"github.com/kubeedge/mappers-go/pkg/ble/globals"
	"github.com/kubeedge/mappers-go/pkg/common"
	mappercommon "github.com/kubeedge/mappers-go/pkg/common"
	"k8s.io/klog/v2"
	"regexp"
	"sync"
	"time"
)

var devices map[string]*globals.BluetoothDev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitory is readonly, if not then set it.
func setVisitor(visitorConfig *configmap.BluetoothVisitorConfig, twin *common.Twin, bleClient *driver.BluetoothClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(1).Info("Visit readonly characteristicUUID: ", visitorConfig.CharacteristicUUID)
		return
	}

	klog.V(2).Infof("Convert type: %s, value: %s ", twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	value, err := common.Convert(twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	if err != nil {
		klog.Error(err)
		return
	}

	valueString, _ := value.(string)
	err = bleClient.Set(ble.MustParse(visitorConfig.CharacteristicUUID), []byte(valueString))
	if err != nil {
		klog.Error(err, visitorConfig)
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

	var dev *globals.BluetoothDev
	var ok bool
	if dev, ok = devices[id]; !ok {
		klog.Error("Device not exist")
		return
	}

	// Get twin map key as the propertyName
	var delta common.DeviceTwinDelta
	if err := json.Unmarshal(message.Payload(), &delta); err != nil {
		klog.Error("Unmarshal message failed: ", err)
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
		var visitorConfig configmap.BluetoothVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Error("Unmarshal visitor config failed")
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.BluetoothClient)
	}
}

// initModbus initialize modbus client
func initBluetooth(protocolConfig configmap.BluetoothProtocolConfig, name string) (client driver.BluetoothClient, err error) {
	if protocolConfig.MacAddress != "" {
		config := driver.BluetoothConfig{
			Addr: protocolConfig.MacAddress,
		}
		client, err = driver.NewClient(config)
	}
	return
}

// initTwin initialize the timer to get twin value.
func initTwin(dev *globals.BluetoothDev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig configmap.BluetoothVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Error(err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.BluetoothClient)

		twinData := TwinData{
			BluetoothClient: dev.BluetoothClient,
			Name:                   dev.Instance.Twins[i].PropertyName,
			Type:                   dev.Instance.Twins[i].Desired.Metadatas.Type,
			BluetoothVisitorConfig: visitorConfig,
			Topic:                  fmt.Sprintf(mappercommon.TopicTwinUpdate, dev.Instance.ID)}
		collectCycle := time.Duration(dev.Instance.Twins[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		timer := mappercommon.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
		wg.Add(1)
		go func() {
			defer wg.Done()
			timer.Start()
		}()
	}
}

// initData initialize the timer to get data.
func initData(dev *globals.BluetoothDev) {
	for i := 0; i < len(dev.Instance.Datas.Properties); i++ {
		var visitorConfig configmap.BluetoothVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Error(err)
			continue
		}

		twinData := TwinData{
			BluetoothClient: dev.BluetoothClient,
			Name:                   dev.Instance.Datas.Properties[i].PropertyName,
			Type:                   dev.Instance.Datas.Properties[i].Metadatas.Type,
			BluetoothVisitorConfig: visitorConfig,
			Topic:                  fmt.Sprintf(mappercommon.TopicDataUpdate, dev.Instance.ID)}
		collectCycle := time.Duration(dev.Instance.Datas.Properties[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		timer := mappercommon.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
		wg.Add(1)
		go func() {
			defer wg.Done()
			timer.Start()
		}()
	}
}

// initSubscribeMqtt subscribe Mqtt topics.
func initSubscribeMqtt(instanceID string) error {
	topic := fmt.Sprintf(mappercommon.TopicTwinUpdateDelta, instanceID)
	klog.V(1).Info("Subscribe topic: ", topic)
	return globals.MqttClient.Subscribe(topic, onMessage)
}

// initGetStatus start timer to get device status and send to eventbus.
func initGetStatus(dev *globals.BluetoothDev) {
	getStatus := GetStatus{Client: dev.BluetoothClient,
		topic: fmt.Sprintf(mappercommon.TopicStateUpdate, dev.Instance.ID)}
	timer := mappercommon.Timer{Function: getStatus.Run, Duration: 1 * time.Second, Times: 0}
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer.Start()
	}()
}

// start start the device.
func start(dev *globals.BluetoothDev) {
	var protocolConfig configmap.BluetoothProtocolConfig
	if err := json.Unmarshal([]byte(dev.Instance.PProtocol.ProtocolConfigs), &protocolConfig); err != nil {
		klog.Error(err)
		return
	}

	client, err := initBluetooth(protocolConfig, protocolConfig.MacAddress)
	if err != nil {
		klog.Error(err)
		return
	}
	dev.BluetoothClient = &client

	initTwin(dev)
	initData(dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Error(err)
		return
	}

	initGetStatus(dev)
}

// DevInit initialize the device datas.
func DevInit(configmapPath string) error {
	devices = make(map[string]*globals.BluetoothDev)
	models = make(map[string]mappercommon.DeviceModel)
	protocols = make(map[string]mappercommon.Protocol)
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
