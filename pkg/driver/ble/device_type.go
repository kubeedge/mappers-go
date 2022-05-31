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

package ble

import (
	"github.com/kubeedge/mappers-go/pkg/common"
)

// BleDev is the ble device configuration and client information.
type BleDev struct {
	Instance  common.DeviceInstance
	BleClient *BleClient
}

// BleVisitorConfig is the BLE register configuration.
type BleVisitorConfig struct {
	CharacteristicUUID string            `json:"characteristicUUID"`
	DataWrite          map[string][]byte `json:"dataWrite"`
	DataConvert        DataConvert       `json:"dataConverter"`
}

type DataConvert struct {
	StartIndex        int                 `json:"startIndex"`
	EndIndex          int                 `json:"endIndex"`
	ShiftLeft         uint                `json:"shiftLeft"`
	ShiftRight        uint                `json:"shiftRight"`
	OrderOfOperations []OrderOfOperations `json:"orderOfOperations"`
}

type OrderOfOperations struct {
	OperationType  string  `json:"operationType"`
	OperationValue float64 `json:"operationValue"`
}

type BleProtocolConfig struct {
	MacAddress string `json:"macAddress"`
}
