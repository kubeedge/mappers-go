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
	"strings"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/pkg/common"
	modbuscommon "github.com/kubeedge/mappers-go/pkg/modbus/common"
	"github.com/kubeedge/mappers-go/pkg/modbus/configmap"
	"github.com/kubeedge/mappers-go/pkg/modbus/driver"
	"github.com/kubeedge/mappers-go/pkg/modbus/globals"
)

// TwinData is the timer structure for getting twin/data.
type TwinData struct {
	Client        *driver.ModbusClient
	Name          string
	Type          string
	VisitorConfig *configmap.ModbusVisitorConfig
	Results       []byte
	Topic         string
}

// Run timer function.
func (td *TwinData) Run() {
	var err error
	td.Results, err = td.Client.Get(td.VisitorConfig.Register,
		td.VisitorConfig.Offset, uint16(td.VisitorConfig.Limit))
	if err != nil {
		klog.Error("Get register failed: ", err)
		return
	}
	// transfer data according to the dpl configuration
	sData, err := modbuscommon.TransferData(td.VisitorConfig.IsRegisterSwap,
		td.VisitorConfig.IsSwap, td.Type, td.VisitorConfig.Scale, td.Results)
	if err != nil {
		klog.Error("Transfer Data failed: ", err)
		return
	}
	// construct payload
	var payload []byte
	if strings.Contains(td.Topic, "$hw") {
		if payload, err = common.CreateMessageTwinUpdate(td.Name, td.Type, sData); err != nil {
			klog.Error("Create message twin update failed")
			return
		}
	} else {
		if payload, err = common.CreateMessageData(td.Name, td.Type, sData); err != nil {
			klog.Error("Create message data failed")
			return
		}
	}
	if err = globals.MqttClient.Publish(td.Topic, payload); err != nil {
		klog.Errorf("Publish topic %v failed, err: %v", td.Topic, err)
		return
	}

	klog.V(2).Infof("Update value: %s, topic: %s", sData, td.Topic)
}
