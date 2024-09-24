package driver

import (
	"sync"
	"time"

	"github.com/kubeedge/mapper-framework/pkg/common"
)

// CustomizedDev is the customized device configuration and client information.
type CustomizedDev struct {
	CustomizedClient *CustomizedClient
	Instance         common.DeviceInstance
}

type CustomizedClient struct {
	// TODO add some variables to help you better implement device drivers
	deviceMutex sync.Mutex
	ProtocolConfig
	DeviceInfo       string                 `json:"deviceInfo"`
	ParsedDeviceInfo map[string]interface{} `json:"parsedDeviceInfo"`
}

type ProtocolConfig struct {
	ProtocolName string `json:"protocolName"`
	ConfigData   `json:"configData"`
}

type ConfigData struct {
	// MQTT protocol config data
	ClientID      string        `json:"clientID"`      // MQTT Client ID
	BrokerURL     string        `json:"brokerURL"`     // MQTT Broker URL
	Topic         string        `json:"topic"`         // Topic for publishing or subscribing
	Message       string        `json:"message"`       // Content of the message
	Username      string        `json:"username"`      // Username for MQTT broker authentication
	Password      string        `json:"password"`      // Password for MQTT broker authentication
	ConnectionTTL time.Duration `json:"connectionTTL"` // Connection timeout duration
	LastMessage   time.Time     `json:"lastMessage"`   // Timestamp of the last received message
	IsData        bool          `json:"isData"`        // Indicates if there is valid data
}

type VisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	DataType string `json:"dataType"`

	ClientID         string                 `json:"clientID"`      // MQTT Client ID
	DeviceInfo       string                 `json:"deviceInfo"`    // Device information, such as device identification or other important information.
	OperationInfo    OperationInfoType      `json:"operationInfo"` // Operation information, such as adding, deleting, modifying and so on.
	SerializedFormat SerializedFormatType   `json:"fileType"`      // Supported formats: json, xml and yaml.
	ParsedMessage    map[string]interface{} `json:"parsedMessage"` // The parsed message
}

// OperationInfoType defines the enumeration values for device operation.
type OperationInfoType uint

const (
	UPDATE OperationInfoType = iota // revision
)

// SerializedFormatType defines the enumeration values for serialized types.
type SerializedFormatType uint

const (
	JSON SerializedFormatType = iota // json
	YAML                             // yaml
	XML                              // xml
	JSONPATH
)
