package driver

import (
	"encoding/json"

	"github.com/kubeedge/mapper-framework/pkg/common"
)

// CustomizedDev is the customized device configuration and client information.
type CustomizedDev struct {
	Instance         common.DeviceInstance
	CustomizedClient *CustomizedClient
}

type CustomizedClient struct {
	deviceName string
	client     *OPCUAClient
	ProtocolConfig
}

type ProtocolConfig struct {
	URL string `json:"url"` // OPC UA server endpoint URL
	//optional fields are stated below, you can choose to implement them based on your protocol requirements
	UserName          string `json:"userName,omitempty"`          // Username for authentication
	Password          string `json:"password,omitempty"`          // Password for authentication
	SecurityPolicy    string `json:"securityPolicy,omitempty"`    // Security policy (None, Basic128Rsa15, Basic256, Basic256Sha256, etc.)
	SecurityMode      string `json:"securityMode,omitempty"`      // Security mode (None, Sign, SignAndEncrypt)
	Certificate       string `json:"certificate,omitempty"`       // Path to certificate file
	PrivateKey        string `json:"privateKey,omitempty"`        // Path to private key file
	RemoteCertificate string `json:"remoteCertificate,omitempty"` // Path to remote certificate file
	Timeout           int64  `json:"timeout,omitempty"`           // Connection timeout in seconds
}

type VisitorConfigData struct {
	NodeID string `json:"nodeID"` //OPC UA Node ID (e.g., "ns=2;i=2")
	//optional fields
	BrowseName string `json:"browseName,omitempty"` //Browse name of the node
	DataType   string `json:"dataType,omitempty"`   // Data type for the node value
}

type VisitorConfig struct {
	VisitorConfigData `json:",inline"`
}

// ParseConfigData parses the protocol config from customizedValues
func (pc *ProtocolConfig) ParseConfigData(configData []byte) error {
	return json.Unmarshal(configData, pc)
}

// ParseVisitorConfig parses the visitor config
func (vc *VisitorConfig) ParseVisitorConfig(visitorConfig []byte) error {
	return json.Unmarshal(visitorConfig, &vc.VisitorConfigData)
}
