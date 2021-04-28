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

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mappercommon "github.com/kubeedge/mappers-go/mappers/common"

	"github.com/kubeedge/mappers-go/mappers/opcua/configmap"
	"github.com/kubeedge/mappers-go/mappers/opcua/driver"
	"github.com/kubeedge/mappers-go/mappers/opcua/globals"
	"k8s.io/klog/v2"
)

var devices map[string]*globals.OPCUADev
var models map[string]mappercommon.DeviceModel
var protocols map[string]mappercommon.Protocol
var wg sync.WaitGroup

// setVisitor check if visitory property is readonly, if not then set it.
func setVisitor(visitorConfig *configmap.VisitorConfigOPCUA, twin *mappercommon.Twin, client *driver.OPCUAClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(1).Info("Visit readonly register: ", visitorConfig.NodeID)
		return
	}

	results, err := client.Set(visitorConfig.NodeID, twin.Desired.Value)
	if err != nil || results != "OK" {
		klog.Errorf("Set error: %v, %v", err, visitorConfig)
		return
	}
}

// onMessage callback function of Mqtt subscribe message.
func onMessage(client mqtt.Client, message mqtt.Message) {
	klog.V(2).Info("Receive message", message.Topic())
	// Get device ID and get device instance
	id := mappercommon.GetDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Device id: ", id)

	var dev *globals.OPCUADev
	var ok bool
	if dev, ok = devices[id]; !ok {
		klog.Error("Device not exist")
		return
	}

	// Get twin map key as the propertyName
	var delta mappercommon.DeviceTwinDelta
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
		var visitorConfig configmap.VisitorConfigOPCUA
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig failed: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.OPCUAClient)
	}
}

// getRemoteCertfile get configuration of remote certification file.
func getRemoteCertfile(customizedValue configmap.CustomizedValue) string {
	if len(customizedValue) != 0 {
		if value, ok := customizedValue["remoteCertificate"]; ok {
			klog.V(4).Info("getRemoteCertfile: ", value)
			return fmt.Sprintf("%v", value)
		}
	}
	klog.Error("getRemoteCertfile null")
	return ""
}

// initOPCUA initialize OPCUA client
func initOPCUA(protocolConfig configmap.ProtocolConfigOPCUA, protocolCommConfig configmap.ProtocolCommonConfigOPCUA) (client *driver.OPCUAClient, err error) {
	if protocolConfig.SecurityPolicy == "" {
		protocolConfig.SecurityPolicy = "None"
	}
	if protocolConfig.SecurityMode == "" {
		protocolConfig.SecurityMode = "None"
	}
	config := driver.OPCUAConfig{
		URL:            protocolConfig.URL,
		User:           protocolConfig.UserName,
		Passwordfile:   protocolConfig.Password,
		SecurityPolicy: protocolConfig.SecurityPolicy,
		SecurityMode:   protocolConfig.SecurityMode,
		Certfile:       protocolConfig.Certificate,
		RemoteCertfile: getRemoteCertfile(protocolCommConfig.CustomizedValues),
		Keyfile:        protocolConfig.PrivateKey,
	}

	return driver.NewClient(config)
}

// initTwin initialize the timer to get twin value.
func initTwin(dev *globals.OPCUADev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig configmap.VisitorConfigOPCUA
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.OPCUAClient)

		twinData := TwinData{Client: dev.OPCUAClient,
			Name:   dev.Instance.Twins[i].PropertyName,
			Type:   dev.Instance.Twins[i].Desired.Metadatas.Type,
			NodeID: visitorConfig.NodeID,
			Topic:  fmt.Sprintf(mappercommon.TopicTwinUpdate, dev.Instance.ID)}
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
func initData(dev *globals.OPCUADev) {
	for i := 0; i < len(dev.Instance.Datas.Properties); i++ {
		var visitorConfig configmap.VisitorConfigOPCUA
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}

		twinData := TwinData{Client: dev.OPCUAClient,
			Name:   dev.Instance.Datas.Properties[i].PropertyName,
			Type:   dev.Instance.Datas.Properties[i].Metadatas.Type,
			NodeID: visitorConfig.NodeID,
			Topic:  fmt.Sprintf(mappercommon.TopicDataUpdate, dev.Instance.ID)}
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
func initGetStatus(dev *globals.OPCUADev) {
	getStatus := GetStatus{Client: dev.OPCUAClient,
		topic: fmt.Sprintf(mappercommon.TopicStateUpdate, dev.Instance.ID)}
	timer := mappercommon.Timer{Function: getStatus.Run, Duration: 1 * time.Second, Times: 0}
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer.Start()
	}()
}

// start start the device.
func start(dev *globals.OPCUADev) {
	var protocolConfig configmap.ProtocolConfigOPCUA
	if err := json.Unmarshal([]byte(dev.Instance.PProtocol.ProtocolConfigs), &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfig error: %v", err)
		return
	}

	var protocolCommConfig configmap.ProtocolCommonConfigOPCUA
	if err := json.Unmarshal([]byte(dev.Instance.PProtocol.ProtocolCommonConfig), &protocolCommConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolCommonConfig error: %v", err)
		return
	}

	klog.V(4).Info("Protocolconfig: ", protocolConfig, protocolCommConfig)
	client, err := initOPCUA(protocolConfig, protocolCommConfig)
	if err != nil {
		klog.Errorf("Init error: %v", err)
		return
	}
	dev.OPCUAClient = client

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
	devices = make(map[string]*globals.OPCUADev)
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
