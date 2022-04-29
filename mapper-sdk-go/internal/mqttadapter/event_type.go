// Package mqttadapter is a package to process mqtt message
// publish messages to EdgeCore and subscribe to EdgeCore messages
package mqttadapter

// BaseMessage the base structure of mqttadapter message.
type BaseMessage struct {
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
}

// TwinValue the structure of twin value.
type TwinValue struct {
	Value    *string       `json:"value,omitempty"`
	Metadata ValueMetadata `json:"metadata,omitempty"`
}

// ValueMetadata the meta of value.
type ValueMetadata struct {
	Timestamp int64 `json:"timestamp,omitempty"`
}

// TypeMetadata the meta of value type.
type TypeMetadata struct {
	Type string `json:"type,omitempty"`
}

// TwinVersion twin version.
type TwinVersion struct {
	CloudVersion int64 `json:"cloud"`
	EdgeVersion  int64 `json:"edge"`
}

// MsgTwin the structure of device twin.
type MsgTwin struct {
	Expected        *TwinValue    `json:"expected,omitempty"`
	Actual          *TwinValue    `json:"actual,omitempty"`
	Optional        *bool         `json:"optional,omitempty"`
	Metadata        *TypeMetadata `json:"metadata,omitempty"`
	ExpectedVersion *TwinVersion  `json:"expected_version,omitempty"`
	ActualVersion   *TwinVersion  `json:"actual_version,omitempty"`
}

// DeviceTwinUpdate the structure of device twin update.
type DeviceTwinUpdate struct {
	BaseMessage
	Twin map[string]*MsgTwin `json:"twin"`
}

// DeviceTwinResult device get result.
type DeviceTwinResult struct {
	BaseMessage
	Twin map[string]*MsgTwin `json:"twin"`
}

// DeviceTwinDelta twin delta.
type DeviceTwinDelta struct {
	BaseMessage
	Twin  map[string]*MsgTwin `json:"twin"`
	Delta map[string]string   `json:"delta"`
}

// DataMetadata data metadata.
type DataMetadata struct {
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
}

// DataValue data value.
type DataValue struct {
	Value    string       `json:"value"`
	Metadata DataMetadata `json:"metadata"`
}

// DeviceData device data structure.
type DeviceData struct {
	BaseMessage
	Data map[string]*DataValue `json:"data"`
}

//MsgAttr the struct of device attr
type MsgAttr struct {
	Value    string        `json:"value"`
	Optional *bool         `json:"optional,omitempty"`
	Metadata *TypeMetadata `json:"metadata,omitempty"`
}

//DeviceUpdate device update.
type DeviceUpdate struct {
	BaseMessage
	State      string              `json:"state,omitempty"`
	Attributes map[string]*MsgAttr `json:"attributes"`
}
