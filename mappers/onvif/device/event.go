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

	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/onvif"
	"github.com/kubeedge/mappers-go/pkg/global"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"
)

// Topics for communication with 3-rd applications.
const (
	TopicOnvifGetResource    = "$hw/events/onvif/resource/get"
	TopicOnvifGetResrouceRet = "$hw/events/onvif/resource/getRet"
)

// OnEventBus callback function of Mqtt message from eventbus.
func OnEventBus(client mqtt.Client, message mqtt.Message) {
	klog.V(2).Info("Receive message", message.Topic())
	// Get device ID and get device instance
	id := getDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Device id: ", id)

	var dev *onvif.OnvifDev
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
		var visitorConfig onvif.OnvifVisitorConfig
		if err := json.Unmarshal(dev.Instance.Twins[i].PVisitor.VisitorConfig, &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.OnvifClient)
	}
}

// On3rdParty callback function of Mqtt message from 3rd-party applications.
func On3rdParty(client mqtt.Client, message mqtt.Message) {
	klog.V(2).Info("Receive message", message.Topic())
	if TopicOnvifGetResource == message.Topic() {
		r := onvif.GetOnvifResources()
		msg, err := json.Marshal(r)
		if err != nil {
			klog.Errorf("Marshal message error: %v", err)
			return
		}
		err = global.MqttClient.Publish(TopicOnvifGetResrouceRet, msg)
		if err != nil {
			klog.Errorf("Publish %s error: %v", TopicOnvifGetResrouceRet, err)
			return
		}
	}
}
