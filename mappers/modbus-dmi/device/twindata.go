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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"k8s.io/klog/v2"

	dmiapi "github.com/kubeedge/kubeedge/pkg/apis/dmi/v1alpha1"

	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/modbus"
	"github.com/kubeedge/mappers-go/pkg/util/grpcclient"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

// TwinData is the timer structure for getting twin/data.
type TwinData struct {
	DeviceID      string
	DeviceName    string
	Client        *modbus.ModbusClient
	Name          string
	Type          string
	VisitorConfig *modbus.ModbusVisitorConfig
	Results       []byte
	LastValue     string
	Topic         string
}

func SwitchRegister(value []byte) []byte {
	for i := 0; i < len(value)/2; i = i + 2 {
		j := len(value) - i - 2
		value[i], value[j] = value[j], value[i]
		value[i+1], value[j+1] = value[j+1], value[i+1]
	}
	return value
}

func SwitchByte(value []byte) []byte {
	if len(value) < 2 {
		return value
	}
	for i := 0; i < len(value); i = i + 2 {
		value[i], value[i+1] = value[i+1], value[i]
	}
	return value
}

func TransferData(isRegisterSwap bool, isSwap bool,
	dataType string, scale float64,
	value []byte) (string, error) {
	// accord IsSwap/IsRegisterSwap to transfer byte array
	if isRegisterSwap {
		SwitchRegister(value)
	}
	if isSwap {
		SwitchByte(value)
	}

	switch dataType {
	case "int":
		switch len(value) {
		case 1:
			data := float64(value[0]) * scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 2:
			data := float64(binary.BigEndian.Uint16(value)) * scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 4:
			data := float64(binary.BigEndian.Uint32(value)) * scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 8:
			data := float64(binary.BigEndian.Uint64(value)) * scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		default:
			return "", errors.New("BytesToInt bytes length is invalid")
		}
	case "double":
		if len(value) != 8 {
			return "", errors.New("BytesToDouble bytes length is invalid")
		}
		bits := binary.BigEndian.Uint64(value)
		data := math.Float64frombits(bits) * scale
		sData := strconv.FormatFloat(data, 'f', 6, 64)
		return sData, nil
	case "float":
		if len(value) != 4 {
			return "", errors.New("BytesToFloat bytes length is invalid")
		}
		bits := binary.BigEndian.Uint32(value)
		data := float64(math.Float32frombits(bits)) * scale
		sData := strconv.FormatFloat(data, 'f', 2, 64)
		return sData, nil
	case "boolean":
		return strconv.FormatBool(value[0] == 0xFF), nil
	case "string":
		for i, b := range value {
			if !isUpper(b) && !isLowercase(b) && !isNumber(b) && !isSpecial(b) {
				value[i] = ' '
			}
		}
		data := strings.ReplaceAll(string(value), " ", "")
		return data, nil
	default:
		return "", errors.New("data type is not support")
	}
}

func isUpper(b byte) bool {
	return 'A' <= b && b <= 'Z'
}

func isLowercase(b byte) bool {
	return 'a' <= b && b <= 'z'
}

func isNumber(b byte) bool {
	return '0' <= b && b <= '9'
}

func isSpecial(b byte) bool {
	whiteList := map[byte]byte{
		'/': '/',
		'-': '-',
		'_': '_',
		'.': '.',
		'%': '%',
		'+': '+',
		',': ',',
		'=': '=',
		'@': '@',
		'#': '#',
		':': ':',
		'^': '^',
		'~': '~',
		'?': '?',
		'&': '&',
		'!': '!',
		'*': '*',
	}
	_, ok := whiteList[b]
	return ok
}

func (td *TwinData) GetPayload() ([]byte, bool, error) {
	var err error

	td.Results, err = td.Client.Get(td.VisitorConfig.Register, td.VisitorConfig.Offset, uint16(td.VisitorConfig.Limit))
	if err != nil {
		return nil, false, fmt.Errorf("get register failed: %v", err)
	}
	// transfer data according to the dpl configuration
	sData, err := TransferData(td.VisitorConfig.IsRegisterSwap,
		td.VisitorConfig.IsSwap, td.Type, td.VisitorConfig.Scale, td.Results)
	if err != nil {
		return nil, false, fmt.Errorf("transfer Data failed: %v", err)
	}

	// do not report if the twin data is not changed to prevent triggering traffic limiting
	changed := sData != td.LastValue
	td.LastValue = sData
	// construct payload
	var payload []byte
	if strings.Contains(td.Topic, "$hw") {
		if payload, err = common.CreateMessageTwinUpdate(td.Name, td.Type, sData); err != nil {
			return nil, false, fmt.Errorf("create message twin update failed: %v", err)
		}
	} else {
		if payload, err = common.CreateMessageData(td.Name, td.Type, sData); err != nil {
			return nil, false, fmt.Errorf("create message data failed: %v", err)
		}
	}
	klog.V(2).Infof("Get the %s value as %s", td.Name, sData)
	return payload, changed, nil
}

// Run timer function.
func (td *TwinData) Run() {
	payload, changed, err := td.GetPayload()
	if err != nil {
		klog.Errorf("twindata %s get payload failed, err: %s", td.Name, err)
		return
	}
	if !changed {
		return
	}

	var msg common.DeviceTwinUpdate
	if err = json.Unmarshal(payload, &msg); err != nil {
		klog.Errorf("twindata %s Unmarshal failed, err: %s", td.Name, err)
		return
	}

	twins := parse.ConvMsgTwinToGrpc(msg.Twin)

	var rdsr = &dmiapi.ReportDeviceStatusRequest{
		DeviceName: td.DeviceID,
		ReportedDevice: &dmiapi.DeviceStatus{
			Twins: twins,
			State: "OK",
		},
	}
	if err := grpcclient.ReportDeviceStatus(rdsr); err != nil {
		klog.Errorf("fail to report device status of %s with err: %+v", rdsr.DeviceName, err)
	}
}
