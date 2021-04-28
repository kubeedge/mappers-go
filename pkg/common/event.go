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
	"crypto/tls"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/pflag"
)

// Joint the topic like topic := fmt.Sprintf(TopicTwinUpdateDelta, deviceID)
const (
	TopicTwinUpdateDelta = "$hw/events/device/%s/twin/update/delta"
	TopicTwinUpdate      = "$hw/events/device/%s/twin/update"
	TopicStateUpdate     = "$hw/events/device/%s/state/update"
	TopicDataUpdate      = "$ke/events/device/%s/data/update"
)

var ErrorProtocolNotExpected = errors.New("protocol not expected")

// MqttClient is parameters for Mqtt client.
type MqttClient struct {
	Qos        byte
	Retained   bool
	IP         string
	User       string
	Passwd     string
	Cert       string
	PrivateKey string
	Client     mqtt.Client
}

// Mqtt is the Mqtt configuration.
type Mqtt struct {
	ServerAddress string `yaml:"server,omitempty"`
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Cert          string `yaml:"certification,omitempty"`
	PrivateKey    string `yaml:"privatekey,omitempty"`
}

func ParseMqttConfig(mqtt *Mqtt) {
	pflag.StringVar(&mqtt.ServerAddress, "mqtt-address", mqtt.ServerAddress, "MQTT broker address")
	pflag.StringVar(&mqtt.Username, "mqtt-username", mqtt.Username, "username")
	pflag.StringVar(&mqtt.Password, "mqtt-password", mqtt.Password, "password")
	pflag.StringVar(&mqtt.Cert, "mqtt-certification", mqtt.Cert, "certification file path")
	pflag.StringVar(&mqtt.PrivateKey, "mqtt-priviatekey", mqtt.PrivateKey, "private key file path")
	pflag.Parse()
}

// newTLSConfig new TLS configuration.
// Only one side check. Mqtt broker check the cert from client.
func newTLSConfig(certfile string, privateKey string) (*tls.Config, error) {
	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(certfile, privateKey)
	if err != nil {
		return nil, err
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}, nil
}

// Connect connect to the Mqtt server.
func (mc *MqttClient) Connect() error {
	opts := mqtt.NewClientOptions().AddBroker(mc.IP).SetClientID("").SetCleanSession(true)
	if mc.Cert != "" {
		tlsConfig, err := newTLSConfig(mc.Cert, mc.PrivateKey)
		if err != nil {
			return err
		}
		opts.SetTLSConfig(tlsConfig)
	} else {
		opts.SetUsername(mc.User)
		opts.SetPassword(mc.Passwd)
	}

	mc.Client = mqtt.NewClient(opts)
	// The token is used to indicate when actions have completed.
	if tc := mc.Client.Connect(); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}

	mc.Qos = 0          // At most 1 time
	mc.Retained = false // Not retained
	return nil
}

// Publish publish Mqtt message.
func (mc *MqttClient) Publish(topic string, payload interface{}) error {
	if tc := mc.Client.Publish(topic, mc.Qos, mc.Retained, payload); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}
	return nil
}

// Subscribe subsribe a Mqtt topic.
func (mc *MqttClient) Subscribe(topic string, onMessage mqtt.MessageHandler) error {
	if tc := mc.Client.Subscribe(topic, mc.Qos, onMessage); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}
	return nil
}

// getTimestamp get current timestamp.
func getTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// CreateMessageTwinUpdate create twin update message.
func CreateMessageTwinUpdate(name string, valueType string, value string) (msg []byte, err error) {
	var updateMsg DeviceTwinUpdate

	updateMsg.BaseMessage.Timestamp = getTimestamp()
	updateMsg.Twin = map[string]*MsgTwin{}
	updateMsg.Twin[name] = &MsgTwin{}
	updateMsg.Twin[name].Actual = &TwinValue{Value: &value}
	updateMsg.Twin[name].Metadata = &TypeMetadata{Type: valueType}

	msg, err = json.Marshal(updateMsg)
	return
}

// CreateMessageData create data message.
func CreateMessageData(name string, valueType string, value string) (msg []byte, err error) {
	var dataMsg DeviceData

	dataMsg.BaseMessage.Timestamp = getTimestamp()
	dataMsg.Data = map[string]*DataValue{}
	dataMsg.Data[name] = &DataValue{}
	dataMsg.Data[name].Value = value
	dataMsg.Data[name].Metadata.Type = valueType
	dataMsg.Data[name].Metadata.Timestamp = getTimestamp()

	msg, err = json.Marshal(dataMsg)
	return
}

// CreateMessageState create device status message.
func CreateMessageState(state string) (msg []byte, err error) {
	var stateMsg DeviceUpdate

	stateMsg.BaseMessage.Timestamp = getTimestamp()
	stateMsg.State = state

	msg, err = json.Marshal(stateMsg)
	return
}

// GetDeviceID extract the device ID from Mqtt topic.
func GetDeviceID(topic string) (id string) {
	re := regexp.MustCompile(`hw/events/device/(.+)/twin/update/delta`)
	return re.FindStringSubmatch(topic)[1]
}

// validateProfilePropertyVisitor validate device instance propertyVisitors
func validateProfilePropertyVisitor(instance *DeviceInstance, deviceModels []DeviceModel) error {
	for k := 0; k < len(instance.PropertyVisitors); k++ {
		modelName := instance.PropertyVisitors[k].ModelName
		propertyName := instance.PropertyVisitors[k].PropertyName
		l := 0
		for l = 0; l < len(deviceModels); l++ {
			if modelName == deviceModels[l].Name {
				m := 0
				for m = 0; m < len(deviceModels[l].Properties); m++ {
					if propertyName == deviceModels[l].Properties[m].Name {
						instance.PropertyVisitors[k].PProperty = deviceModels[l].Properties[m]
						break
					}
				}

				if m == len(deviceModels[l].Properties) {
					err := errors.New("Property not found")
					return err
				}
				break
			}
		}
		if l == len(deviceModels) {
			err := errors.New("Device model not found")
			return err
		}
	}
	return nil
}

// validateProfileTwin validate device instance twins
func validateProfileTwin(instance *DeviceInstance) error {
	for k := 0; k < len(instance.Twins); k++ {
		name := instance.Twins[k].PropertyName
		l := 0
		for l = 0; l < len(instance.PropertyVisitors); l++ {
			if name == instance.PropertyVisitors[l].PropertyName {
				instance.Twins[k].PVisitor = &instance.PropertyVisitors[l]
				break
			}
		}
		if l == len(instance.PropertyVisitors) {
			return errors.New("PropertyVisitor not found")
		}
	}
	return nil
}

// validateProfileData validate device instance data
func validateProfileData(instance *DeviceInstance) error {
	for k := 0; k < len(instance.Datas.Properties); k++ {
		name := instance.Datas.Properties[k].PropertyName
		l := 0
		for l = 0; l < len(instance.PropertyVisitors); l++ {
			if name == instance.PropertyVisitors[l].PropertyName {
				instance.Datas.Properties[k].PVisitor = &instance.PropertyVisitors[l]
				break
			}
		}
		if l == len(instance.PropertyVisitors) {
			return errors.New("PropertyVisitor not found")
		}
	}
	return nil
}

// validateProfileProtocol validate device protocol
func validateProfileProtocol(instance *DeviceInstance, protocols []Protocol, expectedProtocol string) error {
	j := 0
	for j = 0; j < len(protocols); j++ {
		if instance.ProtocolName == protocols[j].Name {
			instance.PProtocol = protocols[j]
			break
		}
	}
	if j == len(protocols) {
		err := errors.New("Protocol not found")
		return err
	}

	if instance.PProtocol.Protocol != expectedProtocol {
		return ErrorProtocolNotExpected
	}
	return nil
}

// ValidateProfileDeviceInstance validate device instance
func ValidateProfileDeviceInstance(instance *DeviceInstance, deviceProfile *DeviceProfile, protocol string) error {
	if err := validateProfileProtocol(instance, deviceProfile.Protocols, protocol); err != nil {
		return err
	}

	if err := validateProfilePropertyVisitor(instance, deviceProfile.DeviceModels); err != nil {
		return err
	}

	if err := validateProfileTwin(instance); err != nil {
		return err
	}

	if err := validateProfileData(instance); err != nil {
		return err
	}

	return nil
}
