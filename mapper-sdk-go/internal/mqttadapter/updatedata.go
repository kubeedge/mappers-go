package mqttadapter

import (
	"encoding/json"
	"time"
)

// CreateMessageData create binary data for structure of DeviceData
func CreateMessageData(name string, valueType string, value string) (msg []byte, err error) {
	var dataMsg DeviceData

	dataMsg.BaseMessage.Timestamp = time.Now().UnixNano() / 1e6
	dataMsg.Data = map[string]*DataValue{}
	dataMsg.Data[name] = &DataValue{}
	dataMsg.Data[name].Value = value
	dataMsg.Data[name].Metadata.Type = valueType
	dataMsg.Data[name].Metadata.Timestamp = time.Now().UnixNano() / 1e6

	msg, err = json.Marshal(dataMsg)
	return
}
