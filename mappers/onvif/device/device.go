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
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mappers/common"
	"github.com/kubeedge/mappers-go/mappers/onvif/configmap"
	"github.com/kubeedge/mappers-go/mappers/onvif/driver"
	"github.com/kubeedge/mappers-go/mappers/onvif/globals"
)

var devices map[string]*globals.OnvifDev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitory property is readonly, if not then set it.
func setVisitor(visitorConfig *configmap.OnvifVisitorConfig, twin *common.Twin, client *driver.OnvifClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		return
	}

	err := client.Set(visitorConfig.ConfigData, twin.Desired.Value)
	if err != nil {
		klog.Errorf("Set error: %v, %v", err, visitorConfig)
		return
	}
}

// getDeviceID extract the device ID from Mqtt topic.
func getDeviceID(topic string) (id string) {
	re := regexp.MustCompile(`hw/events/device/(.+)/twin/update/delta`)
	return re.FindStringSubmatch(topic)[1]
}

// initOnvif initialize Onvif client.
func initOnvif(name string, protocolConfig configmap.OnvifProtocolCommonConfig) (client *driver.OnvifClient, err error) {
	OnvifConfig := driver.OnvifConfig{Name: name}

	var ok bool
	if OnvifConfig.URL, ok = protocolConfig.ConfigData["url"]; !ok {
		return nil, errors.New("Protocol config has not url")
	}
	if OnvifConfig.User, ok = protocolConfig.ConfigData["userName"]; !ok {
		klog.V(2).Info("protocol config has not userName")
	}
	if OnvifConfig.Passwordfile, ok = protocolConfig.ConfigData["password"]; !ok {
		klog.V(2).Info("protocl config has not passwordfile")
	}
	if OnvifConfig.Certfile, ok = protocolConfig.ConfigData["cert"]; !ok {
		return nil, errors.New("Protocol config has not certfile")
	}
	if OnvifConfig.RemoteCertfile, ok = protocolConfig.ConfigData["remoteCert"]; !ok {
		klog.V(2).Info("protocol config has not remoteCertfile")
	}
	if OnvifConfig.Keyfile, ok = protocolConfig.ConfigData["key"]; !ok {
		klog.V(2).Info("protocl config has not keyfile")
	}

	client, _ = driver.NewClient(OnvifConfig)
	return client, nil
}

// initTwin initialize the timer to get twin value.
func initTwin(dev *globals.OnvifDev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig configmap.OnvifVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.OnvifClient)
	}
}

// initSubscribeMqtt subscribe Mqtt topics.
func initSubscribeMqtt(instanceID string) error {
	topic := fmt.Sprintf(common.TopicTwinUpdateDelta, instanceID)
	klog.V(1).Info("Subscribe topic: ", topic)
	err = globals.MqttClient.Subscribe(topic, OnEventBus)
	if err != nil {
		return err
	}
	return globals.MqttClient.Subscribe(TopicOnvifGetResource, On3rdParty)
}

// initGetStatus start timer to get device status and send to eventbus.
func initGetStatus(dev *globals.OnvifDev) {
	getStatus := GetStatus{Client: dev.OnvifClient,
		topic: fmt.Sprintf(common.TopicStateUpdate, dev.Instance.ID)}
	timer := common.Timer{Function: getStatus.Run, Duration: 1 * time.Second, Times: 0}
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer.Start()
	}()
}

// start start the device.
func start(dev *globals.OnvifDev) {
	var protocolConfig configmap.OnvifProtocolConfig
	if err := json.Unmarshal([]byte(dev.Instance.PProtocol.ProtocolConfigs), &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfig error: %v", err)
		return
	}

	client, err := initOnvif(dev.Instance.Name, protocolCommConfig)
	if err != nil {
		klog.Errorf("Init error: %v", err)
		return
	}
	dev.OnvifClient = client

	initTwin(dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Errorf("Init subscribe mqtt error: %v", err)
		return
	}

	initGetStatus(dev)
}

// DevInit initialize the device datas.
func DevInit(configmapPath string) error {
	devices = make(map[string]*globals.OnvifDev)
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
