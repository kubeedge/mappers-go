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

package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strconv"
)

func SwitchRegister(value []byte) []byte {

	for i := 0; i < len(value)/2; i = i+2 {
		j := len(value)-i-2
		value[i], value[j] = value[j], value[i]
		value[i+1], value[j+1] = value[j+1], value[i+1]
	}

	return value
}

func SwitchByte(value []byte) []byte {

	for i := 0; i < len(value); i = i+2 {
		value[i], value[i + 1] = value[i + 1], value[i]
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

	byteBuff := bytes.NewBuffer(value)
	switch dataType {
	case "int":
		switch len(value) {
		case 1:
			var tmp int8
			err := binary.Read(byteBuff, binary.BigEndian, &tmp)
			if err != nil {
				return "", err
			}
			data := float64(tmp) * scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 2:
			var tmp int16
			err := binary.Read(byteBuff, binary.BigEndian, &tmp)
			if err != nil {
				return "", err
			}
			data := float64(tmp) * scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 4:
			var tmp int32
			err := binary.Read(byteBuff, binary.BigEndian, &tmp)
			if err != nil {
				return "", err
			}
			data := float64(tmp) * scale
			sData := strconv.FormatInt(int64(data), 10)
			return sData, nil
		case 8:
			var tmp int64
			err := binary.Read(byteBuff, binary.BigEndian, &tmp)
			if err != nil {
				return "", err
			}
			data := float64(tmp) * scale
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
			data := math.Float64frombits(bits) * scale
			sData := strconv.FormatFloat(data,'f', 6, 64)
			return sData, nil
		}
	case "float":
		if len(value) != 4 {
			return "", errors.New("BytesToFloat bytes length is invalid")
		} else {
			bits := binary.BigEndian.Uint32(value)
			data := float64(math.Float32frombits(bits)) * scale
			sData := strconv.FormatFloat(data,'f', 6, 64)
			return sData, nil
		}
	case "boolean":
		return strconv.FormatBool(value[0] == 1), nil
	case "string":
		data := string(value)
		return data, nil
	default:
		return "", errors.New("Data type is not support")
	}
}
