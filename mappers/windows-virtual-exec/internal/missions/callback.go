/*
Copyright 2024 The KubeEdge Authors.
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

package missions

import (
	"encoding/json"
	"fmt"
	"path"

	mq "github.com/eclipse/paho.mqtt.golang"
	klog "k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/core/mqtt"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/dto"
)

func InitCallback(nodeName string) {
	mqttClient := mqtt.GetClient()
	err := mqttClient.Subscribe(fmt.Sprintf(mqtt.TopicRecNodeDeviceUpdate, nodeName), onMembershipUpdateMessage)
	if err != nil {
		klog.Error("Subscribe error: ", err)
	} else {
		klog.Info("Subscribe topic: ", fmt.Sprintf(mqtt.TopicRecNodeDeviceUpdate, nodeName))
	}
	err = mqttClient.Subscribe(fmt.Sprintf(mqtt.TopicRecModeDeviceListResponse, nodeName), onMembershipListMessage)
	if err != nil {
		klog.Error("Subscribe error: ", err)
	} else {
		klog.Info("Subscribe topic: ", fmt.Sprintf(mqtt.TopicRecModeDeviceListResponse, nodeName))
	}
	err = mqttClient.Subscribe(fmt.Sprintf(mqtt.TopicRevTwinUpdateDelta, "+"), onTwinDelta)
	if err != nil {
		klog.Error("Subscribe error: ", err)
	} else {
		klog.Info("Subscribe topic: ", fmt.Sprintf(mqtt.TopicRevTwinUpdateDelta, "+"))
	}
	err = mqttClient.Subscribe(fmt.Sprintf(mqtt.TopicRecTwinInfoResponse, "+"), onTwinInfo)
	if err != nil {
		klog.Error("Subscribe error: ", err)
	} else {
		klog.Info("Subscribe topic: ", fmt.Sprintf(mqtt.TopicRecTwinInfoResponse, "+"))
	}
}

func onMembershipUpdateMessage(_ mq.Client, message mq.Message) {
	klog.V(2).Info("Receive message from topic: ", message.Topic())
	nodeID := mqtt.GetNodeID(message.Topic())
	if nodeID == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Node id: ", nodeID)
	var req dto.DeviceListUpdate
	if err := json.Unmarshal(message.Payload(), &req); err != nil {
		klog.Error("Unmarshal error: ", err)
		return
	}

	klog.Info("Receive device list update: ", "nodeId: ", nodeID, " update: ", len(req.AddedDevices), " delete: ", len(req.RemovedDevices))

	for _, device := range req.RemovedDevices {
		RemoveMission(device.ID)
	}

	for _, device := range req.RemovedDevices {
		if _, ok := cache.Load(device.ID); ok {
			klog.Info("Device already exists: ", device.ID)
			continue
		}
		klog.Info("Waiting twin update to create device: ", device.ID)
	}
}

func onMembershipListMessage(_ mq.Client, message mq.Message) {
	klog.V(2).Info("Receive message from topic: ", message.Topic())
	nodeID := mqtt.GetNodeID(message.Topic())
	if nodeID == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Node id: ", nodeID)
	var req dto.DeviceList
	if err := json.Unmarshal(message.Payload(), &req); err != nil {
		klog.Error("Unmarshal error: ", err)
		return
	}

	klog.Info("Receive device list: ", "nodeID: ", nodeID, " count: ", len(req.Devices))
	for _, device := range req.Devices {
		err := mqtt.GetClient().Publish(fmt.Sprintf(mqtt.TopicPubTwinInfoRequest, device.ID), mqtt.CreateEmptyMessage())
		if err != nil {
			klog.Error("Publish error: ", err)
			return
		}
	}
}

func onTwinDelta(_ mq.Client, message mq.Message) {
	klog.V(2).Info("Receive message from topic: ", message.Topic())
	id := mqtt.GetDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Mission id: ", id)
	var req dto.MissionDelta
	if err := json.Unmarshal(message.Payload(), &req); err != nil {
		klog.Error("Unmarshal error: ", err)
		return
	}

	// check params
	if req.Twin.ExecCommand == nil || req.Twin.ExecFileName == nil || req.Twin.ExecFileContent == nil {
		klog.Error("Twin format error")
		return
	}

	if req.Twin.ExecCommand.Expected == nil || req.Twin.ExecCommand.Expected.Value == nil {
		klog.Error("Twin ExecCommand format error")
		return
	}

	if req.Twin.ExecFileName.Expected == nil || req.Twin.ExecFileName.Expected.Value == nil {
		klog.Error("Twin ExecFileName format error")
		return
	}

	if req.Twin.ExecFileContent.Expected == nil || req.Twin.ExecFileContent.Expected.Value == nil {
		klog.Error("Twin ExecFileContent format error")
		return
	}

	err := mqtt.GetClient().Publish(fmt.Sprintf(mqtt.TopicPubTwinInfoRequest, id), mqtt.CreateEmptyMessage())
	if err != nil {
		klog.Error("Publish error: ", err)
		return
	}
}

// onTwinInfo indeed trigger a new mission
func onTwinInfo(_ mq.Client, message mq.Message) {
	klog.V(2).Info("Receive message from topic: ", message.Topic())
	id := mqtt.GetDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Mission id: ", id)
	var req dto.MissionDelta
	if err := json.Unmarshal(message.Payload(), &req); err != nil {
		klog.Error("Unmarshal error: ", err)
		return
	}
	if req.Twin.ExecCommand == nil || req.Twin.ExecFileName == nil || req.Twin.ExecFileContent == nil || req.Twin.Output == nil || req.Twin.Status == nil {
		klog.Error("Twin format error")
		return
	}

	if req.Twin.ExecCommand.Expected == nil || req.Twin.ExecCommand.Expected.Value == nil {
		klog.Error("Twin ExecCommand format error")
		return
	}

	if req.Twin.ExecFileName.Expected == nil || req.Twin.ExecFileName.Expected.Value == nil {
		klog.Error("Twin ExecFileName format error")
		return
	}

	if req.Twin.ExecFileContent.Expected == nil || req.Twin.ExecFileContent.Expected.Value == nil {
		klog.Error("Twin ExecFileContent format error")
		return
	}

	_, err := NewMission(MissionConfig{
		UniqueName:       id,
		Command:          *req.Twin.ExecCommand.Expected.Value,
		FileContent:      *req.Twin.ExecFileContent.Expected.Value,
		FileName:         *req.Twin.ExecFileName.Expected.Value,
		WorkingDirectory: path.Join("tmp", id),
	})
	if err != nil {
		klog.Error("NewMission error: ", err)
		return
	}
}
