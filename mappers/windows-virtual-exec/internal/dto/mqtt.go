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

package dto

type BaseMessage struct {
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
}

type TwinValue struct {
	Value    *string       `json:"value,omitempty"`
	Metadata ValueMetadata `json:"metadata,omitempty"`
}

type ValueMetadata struct {
	Timestamp int64 `json:"timestamp,omitempty"`
}

type TypeMetadata struct {
	Type string `json:"type,omitempty"`
}

type TwinVersion struct {
	CloudVersion int64 `json:"cloud"`
	EdgeVersion  int64 `json:"edge"`
}

type MsgTwin struct {
	Expected        *TwinValue    `json:"expected,omitempty"`
	Actual          *TwinValue    `json:"actual,omitempty"`
	Optional        *bool         `json:"optional,omitempty"`
	Metadata        *TypeMetadata `json:"metadata,omitempty"`
	ExpectedVersion *TwinVersion  `json:"expected_version,omitempty"`
	ActualVersion   *TwinVersion  `json:"actual_version,omitempty"`
}

type DeviceTwinUpdate struct {
	BaseMessage
	Twin map[string]*MsgTwin `json:"twin"`
}

type DeviceTwinResult struct {
	BaseMessage
	Twin map[string]*MsgTwin `json:"twin"`
}

type DeviceTwinDelta struct {
	BaseMessage
	Twin  map[string]*MsgTwin `json:"twin"`
	Delta map[string]string   `json:"delta"`
}

type MsgAttr struct {
	Value    string        `json:"value"`
	Optional *bool         `json:"optional,omitempty"`
	Metadata *TypeMetadata `json:"metadata,omitempty"`
}

type DeviceStatusUpdate struct {
	BaseMessage
	State      string              `json:"state,omitempty"`
	Attributes map[string]*MsgAttr `json:"attributes"`
}

type DeviceListUpdate struct {
	BaseMessage
	AddedDevices   []DeviceInfo `json:"added_devices"`
	RemovedDevices []DeviceInfo `json:"removed_devices"`
}

type DeviceList struct {
	BaseMessage
	Devices []DeviceInfo `json:"devices"`
}

type DeviceInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MissionDelta struct {
	BaseMessage
	Twin struct {
		ExecCommand     *MsgTwin `json:"exec-command"`
		ExecFileName    *MsgTwin `json:"exec-file-name"`
		ExecFileContent *MsgTwin `json:"exec-file-content"`
		Output          *MsgTwin `json:"output"`
		Status          *MsgTwin `json:"status"`
	} `json:"twin"`
	Delta struct {
		ExecCommand     string `json:"exec-command"`
		ExecFileName    string `json:"exec-file-name"`
		ExecFileContent string `json:"exec-file-content"`
		Output          string `json:"output"`
		Status          string `json:"status"`
	} `json:"delta"`
}

type MissionTwins struct {
	BaseMessage
	Twin struct {
		ExecCommand     *MsgTwin `json:"exec-command"`
		ExecFileName    *MsgTwin `json:"exec-file-name"`
		ExecFileContent *MsgTwin `json:"exec-file-content"`
		Output          *MsgTwin `json:"output"`
		Status          *MsgTwin `json:"status"`
	} `json:"twin"`
}
