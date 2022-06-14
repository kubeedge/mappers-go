// Package instancepool used to provide a pool to get MapperService's instance,
// like deviceInstances map[string]*configmap.DeviceInstance ,deviceModels map[string]*configmap.DeviceModel, etc.
package instancepool

import (
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/clients/mqttclient"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
)

// MqttClientName contains the name of device service's ConfigurationStruct implementation in the DIC.
var MqttClientName = di.TypeInstanceToName(mqttclient.MqttClient{})

// MqttClientNameFrom helper function queries the DIC and returns device service's ConfigurationStruct implementation.
func MqttClientNameFrom(get di.Get) mqttclient.MqttClient {
	return get(MqttClientName).(mqttclient.MqttClient)
}
