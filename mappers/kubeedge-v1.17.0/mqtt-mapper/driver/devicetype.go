package driver

import (
	"sync"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kubeedge/mapper-framework/pkg/common"
)

// CustomizedDev is the customized device configuration and client information.
type CustomizedDev struct {
	CustomizedClient *CustomizedClient
	Instance         common.DeviceInstance
}

type CustomizedClient struct {
	// TODO add some variables to help you better implement device drivers
	deviceMutex    sync.Mutex
	ProtocolConfig
	MessageQueue MessageQueue
}

type ProtocolConfig struct {
	ProtocolName string   `json:"protocolName"`
	ConfigData            `json:"configData"`
}

type ConfigData struct {
	// MQTT protocol config data
    ClientID     string `json:"clientID"`  // MQTT Client ID
    Topic        string `json:"topic"`      // Client need to specify a topic when publishing or subsribing.
    Message      string `json:"message"`    // Content of the message
}


type VisitorConfig struct {
	ProtocolName      string  `json:"protocolName"`
	VisitorConfigData         `json:"configData"`
}

type VisitorConfigData struct {
	DataType string `json:"dataType"`

	ClientID          string                `json:"clientID"`     // MQTT Client ID
	DeviceInfo        string                `json:"deviceInfo"`    // Device information, such as device identification or other important information.
	OperationInfo     OperationInfoType     `json:"operationInfo"` // Operation information, such as adding, deleting, modifying and so on.
	SerializedFormat  SerializedFormatType  `json:"fileType"`      // Supported formats: json, xml and yaml.
	ParsedMessage     interface{}           `json:"parsedMessage"` // The parsed message
}

// OperationInfoType defines the enumeration values for device operation.
type OperationInfoType uint

const (
	FULLTEXTMODIFY OperationInfoType = iota   // full text revision
	PATHMODIFY                          	  // path revision
	VALUEMODIFY                               // value revision
)

// SerializedFormatType defines the enumeration values for serialized types.
type SerializedFormatType uint

const (
	JSON SerializedFormatType = iota // json
	YAML                             // yaml
	XML                              // xml
)

// MessageQueue defines generic message queue operations and contains three methods:
// Publish is used to publish a message to the specified topic, the type of the message is interface{} in order to support multiple message formats.
// Subscribe subscribes to the specified topic, return the received message.
// Unsubscribe unsubscribes to the specified topic.
type MessageQueue interface {
	Publish(topic string, message interface{}) error
	Subscribe(topic string) (interface{}, error)
	Unsubscribe(topic string) error
}

type MqttMessageQueue struct {
    client mqtt.Client
}
