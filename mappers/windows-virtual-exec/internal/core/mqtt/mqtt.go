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

package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"regexp"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/dto"
)

// Joint the topic like topic := fmt.Sprintf(TopicTwinUpdateDelta, deviceID)
const (
	TopicRevTwinUpdateDelta = "$hw/events/device/%s/twin/update/delta"

	TopicPubTwinUpdateRequest  = "$hw/events/device/%s/twin/update"
	TopicRecTwinUpdateResponse = "$hw/events/device/%s/twin/update/result"

	TopicPubTwinInfoRequest  = "$hw/events/device/%s/twin/get"
	TopicRecTwinInfoResponse = "$hw/events/device/%s/twin/get/result"

	TopicPubDeviceStateUpdateRequest  = "$hw/events/device/%s/state/update"
	TopicRecDeviceStateUpdateResponse = "$hw/events/device/%s/state/update/result"

	TopicPubNodeDeviceListRequest  = "$hw/events/node/%s/membership/get"
	TopicRecModeDeviceListResponse = "$hw/events/node/%s/membership/get/result"

	TopicRecNodeDeviceUpdate = "$hw/events/node/%s/membership/updated"
)

var client *Client

func GetClient() *Client {
	return client
}

// MqttClient is parameters for Mqtt client.
type Client struct {
	Qos        byte
	Retained   bool
	IP         string
	User       string
	Passwd     string
	Cert       string
	PrivateKey string
	Client     mqtt.Client
}

func InitClient(ip, user, passwd, cert, privkey string) error {
	client = &Client{
		IP:         ip,
		User:       user,
		Passwd:     passwd,
		Cert:       cert,
		PrivateKey: privkey,
	}
	return client.Connect()
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
func (mc *Client) Connect() error {
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
func (mc *Client) Publish(topic string, payload interface{}) error {
	if tc := mc.Client.Publish(topic, mc.Qos, mc.Retained, payload); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}
	return nil
}

// Subscribe subscribe a Mqtt topic.
func (mc *Client) Subscribe(topic string, onMessage mqtt.MessageHandler) error {
	if tc := mc.Client.Subscribe(topic, mc.Qos, onMessage); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}
	return nil
}

// getTimestamp get current timestamp.
func getTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func CreateEmptyMessage() (msg []byte) {
	var emptyMsg dto.BaseMessage

	emptyMsg.Timestamp = getTimestamp()

	msg, _ = json.Marshal(emptyMsg)
	return
}

// CreateMessageTwinUpdate create twin update message.
func CreateMessageTwinUpdate(info map[string]string) (msg []byte, err error) {
	var updateMsg dto.DeviceTwinUpdate

	updateMsg.BaseMessage.Timestamp = getTimestamp()
	updateMsg.Twin = map[string]*dto.MsgTwin{}

	for k := range info {
		value := info[k]
		updateMsg.Twin[k] = &dto.MsgTwin{}
		updateMsg.Twin[k].Actual = &dto.TwinValue{Value: &value}
		//updateMsg.Twin[k].Metadata = &dto.TypeMetadata{Type: "string"}
	}

	msg, err = json.Marshal(updateMsg)
	return
}

// CreateMessageState create device status message.
func CreateMessageState(state string) (msg []byte, err error) {
	var stateMsg dto.DeviceStatusUpdate

	stateMsg.BaseMessage.Timestamp = getTimestamp()
	stateMsg.State = state

	msg, err = json.Marshal(stateMsg)
	return
}

// GetDeviceID extract the device ID from Mqtt topic.
func GetDeviceID(topic string) (id string) {
	re := regexp.MustCompile(`hw/events/device/(.+?)/`)
	return re.FindStringSubmatch(topic)[1]
}

func GetNodeID(topic string) (id string) {
	re := regexp.MustCompile(`hw/events/node/(.+?)/`)
	return re.FindStringSubmatch(topic)[1]
}
