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
	"fmt"
	"strconv"
	"strings"

	"github.com/currantlabs/ble"
	"k8s.io/klog/v2"

	"github.com/kubeedge/kubeedge/pkg/apis/devices/v1alpha2"
	"github.com/kubeedge/mappers-go/mappers/ble/configmap"
	"github.com/kubeedge/mappers-go/mappers/ble/driver"
	"github.com/kubeedge/mappers-go/mappers/ble/globals"
	"github.com/kubeedge/mappers-go/mappers/common"
)

// TwinData is the timer structure for getting twin/data.
type TwinData struct {
	BleClient            *driver.BleClient
	Name                 string
	Type                 string
	BleVisitorConfig     configmap.BleVisitorConfig
	Result               string
	Topic                string
	FindedCharacteristic interface{}
}

// Run timer function.
func (td *TwinData) Run() {
	c := td.FindedCharacteristic.(*ble.Characteristic)
	// read data actively
	b, err := td.BleClient.Read(c)
	if err != nil {
		klog.Errorf("Failed to read characteristic: %s\n", err)
	}

	td.Result = fmt.Sprintf("%f", td.ConvertReadData(b))

	if err = td.handlerPublish(); err != nil {
		klog.Errorf("publish data to mqtt failed: %v", err)
	}
}

func (td *TwinData) notificationHandler() func(req []byte) {
	return func(req []byte) {
		td.Result = fmt.Sprintf("%f", td.ConvertReadData(req))
		if err := td.handlerPublish(); err != nil {
			klog.Errorf("publish data to mq failed: %v", err)
		}
	}
}

func (td *TwinData) handlerPublish() (err error) {
	// construct payload
	var payload []byte
	if strings.Contains(td.Topic, "$hw") {
		if payload, err = common.CreateMessageTwinUpdate(td.Name, td.Type, td.Result); err != nil {
			klog.Errorf("Create message twin update failed: %v", err)
			return
		}
	} else {
		if payload, err = common.CreateMessageData(td.Name, td.Type, td.Result); err != nil {
			klog.Errorf("Create message data failed: %v", err)
			return
		}
	}
	if err = globals.MqttClient.Publish(td.Topic, payload); err != nil {
		klog.Errorf("Publish topic %v failed, err: %v", td.Topic, err)
	}

	klog.V(2).Infof("Update value: %s, topic: %s", td.Result, td.Topic)
	return
}

// ConvertReadData is the function responsible to convert the data read from the device into meaningful data.
// If currently logic of converting data is not suitable for your device, you can change ConvertReadData function manually.
func (td *TwinData) ConvertReadData(data []byte) float64 {
	var intermediateResult uint64
	var initialValue []byte
	var initialStringValue = ""
	if td.BleVisitorConfig.DataConvert.StartIndex <= td.BleVisitorConfig.DataConvert.EndIndex {
		for index := td.BleVisitorConfig.DataConvert.StartIndex; index <= td.BleVisitorConfig.DataConvert.EndIndex; index++ {
			initialValue = append(initialValue, data[index])
		}
	} else {
		for index := td.BleVisitorConfig.DataConvert.StartIndex; index >= td.BleVisitorConfig.DataConvert.EndIndex; index-- {
			initialValue = append(initialValue, data[index])
		}
	}
	for _, value := range initialValue {
		initialStringValue = initialStringValue + strconv.Itoa(int(value))
	}
	initialByteValue, _ := strconv.ParseUint(initialStringValue, 10, 64)
	if td.BleVisitorConfig.DataConvert.ShiftLeft != 0 {
		intermediateResult = initialByteValue << td.BleVisitorConfig.DataConvert.ShiftLeft
	} else if td.BleVisitorConfig.DataConvert.ShiftRight != 0 {
		intermediateResult = initialByteValue >> td.BleVisitorConfig.DataConvert.ShiftRight
	} else {
		intermediateResult = initialByteValue
	}
	finalResult := float64(intermediateResult)
	for _, orderOfOperations := range td.BleVisitorConfig.DataConvert.OrderOfOperations {
		if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothAdd)) {
			finalResult = finalResult + orderOfOperations.OperationValue
		} else if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothSubtract)) {
			finalResult = finalResult - orderOfOperations.OperationValue
		} else if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothMultiply)) {
			finalResult = finalResult * orderOfOperations.OperationValue
		} else if strings.ToUpper(orderOfOperations.OperationType) == strings.ToUpper(string(v1alpha2.BluetoothDivide)) {
			finalResult = finalResult / orderOfOperations.OperationValue
		}
	}
	return finalResult
}
