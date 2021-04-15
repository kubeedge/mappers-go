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
	"strconv"
	"strings"
	"bytes"
	"encoding/binary"
	"errors"
	"math"

	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/modbus/driver"
	"github.com/kubeedge/mappers-go/pkg/modbus/globals"
	"github.com/kubeedge/kubeedge/mappers/modbus-go/configmap"
	"k8s.io/klog/v2"
)

// TwinData is the timer structure for getting twin/data.
type TwinData struct {
	Client       *driver.ModbusClient
	Name         string
	Type         string
	VisitorConfig *configmap.ModbusVisitorConfig
	Results      []byte
	Topic        string
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
	sData, err := TransferData(td, td.Results)
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
		klog.Error(err)
	}

	klog.V(2).Infof("Update value: %s, topic: %s", sData, td.Topic)
}

func TransferBytes(isRegisterSwap bool, isSwap bool, value []byte) []byte {

	if isRegisterSwap && isSwap {
		for i := 0; i < len(value)/2; i++ {
			j := len(value) - i - 1
			value[i], value[j] = value[j], value[i]
		}
	} else if isRegisterSwap {
		for i := 0; i < len(value)/2; i = i+2 {
			j := len(value)-i-2
			value[i], value[j] = value[j], value[i]
			value[i+1], value[j+1] = value[j+1], value[i]
		}
	} else if isSwap {
		for i := 0; i < len(value)/2; i++ {
			j := i*2 + 1
			value[i*2], value[j] = value[j], value[i*2]
		}
	}

	return value
}

func TransferData(td *TwinData, value []byte) (string, error) {

	// accord IsSwap/IsRegisterSwap to transfer byte array
	TransferBytes(td.VisitorConfig.IsRegisterSwap, td.VisitorConfig.IsSwap, value)

	byteBuff := bytes.NewBuffer(value)
	switch td.Type {
	case "int":
		switch len(value) {
		case 1:
			var tmp int8
			err := binary.Read(byteBuff, binary.BigEndian, &tmp)
			if err != nil {
				return "", err
			}
			data := float64(tmp) * td.VisitorConfig.Scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 2:
			var tmp int16
			err := binary.Read(byteBuff, binary.BigEndian, &tmp)
			if err != nil {
				return "", err
			}
			data := float64(tmp) * td.VisitorConfig.Scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 4:
			var tmp int32
			err := binary.Read(byteBuff, binary.BigEndian, &tmp)
			if err != nil {
				return "", err
			}
			data := float64(tmp) * td.VisitorConfig.Scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		default:
			return "", errors.New("BytesToInt bytes length is invalid")
		}
	case "double":
		if len(value) != 8 {
			return "", errors.New("BytesToDouble bytes length is invalid")
		} else {
			bits := binary.BigEndian.Uint64(value)
			data := math.Float64frombits(bits) * td.VisitorConfig.Scale
			sData := strconv.FormatFloat(data,'f', 6, 64)
			return sData, nil
		}
	case "float":
		if len(value) != 4 {
			return "", errors.New("BytesToFloat bytes length is invalid")
		} else {
			bits := binary.BigEndian.Uint32(value)
			data := float64(math.Float32frombits(bits)) * td.VisitorConfig.Scale
			sData := strconv.FormatFloat(data,'f', 6, 64)
			return sData, nil
		}
	case "boolean":
		var data bool
		err := binary.Read(byteBuff, binary.BigEndian, &data)
		if err != nil {
			return "", err
		}
		sData := strconv.FormatBool(data)
		return sData, nil
	case "string":
		var data string
		err := binary.Read(byteBuff, binary.BigEndian, &data)
		if err != nil {
			return "", err
		}
		return data, nil
	default:
		return "", errors.New("Data type is not support")
	}
}

