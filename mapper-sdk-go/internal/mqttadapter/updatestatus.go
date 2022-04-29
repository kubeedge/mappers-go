package mqttadapter

import (
	"encoding/json"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/clients/mqttclient"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/controller"
	"k8s.io/klog/v2"
	"time"
)

// StatusData the structure of device status.
type StatusData struct {
	topic      string
	MqttClient mqttclient.MqttClient
	driverUnit DriverUnit
}

// Run start timer function to get device's status and send it to mqtt broker
func (sd *StatusData) Run() {
	var err error
	sData := controller.GetDeviceStatus(sd.driverUnit.instanceID, sd.driverUnit.twin, sd.driverUnit.drivers, sd.driverUnit.mutex, sd.driverUnit.dic)
	var payload []byte
	if payload, err = CreateMessageState(sData); err != nil {
		klog.Errorf("Create message state failed: %v", err)
		return
	}
	//  push payload to MQTT broker
	if err = sd.MqttClient.Publish(sd.topic, payload); err != nil {
		klog.Errorf("Publish failed: %v", err)
		return
	}
}

// CreateMessageState create binary data for structure of DeviceData
func CreateMessageState(state string) (msg []byte, err error) {
	var stateMsg DeviceUpdate
	stateMsg.BaseMessage.Timestamp = time.Now().UnixNano() / 1e6
	stateMsg.State = state
	msg, err = json.Marshal(stateMsg)
	return
}
