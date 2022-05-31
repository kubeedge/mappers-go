package opcua

import (
	"github.com/kubeedge/mappers-go/pkg/common"
)

// OPCUADev is the OPCUA device configuration and client information.
type OPCUADev struct {
	Instance    common.DeviceInstance
	OPCUAClient *OPCUAClient
}

// VisitorConfigOPCUA common visitor configurations for opc-ua protocol
type VisitorConfigOPCUA struct {
	// Required: The ID of opc-ua node, e.g. "ns=1,i=1005"
	NodeID string `json:"nodeID,omitempty"`
	// The name of opc-ua node
	BrowseName string `json:"browseName,omitempty"`
}

// ProtocolConfigOPCUA configuration for opc-ua protocol.
type ProtocolConfigOPCUA struct {
	// Required: The URL for opc server endpoint.
	URL string `json:"url,omitempty"`
	// Username for access opc server.
	// +optional
	UserName string `json:"userName,omitempty"`
	// Password file for access opc server.
	// +optional
	Password string `json:"password,omitempty"`
	// Defaults to "None". The value could be "None", "Basic128Rsa15", "Basic256",
	// "Basic256Sha256", "Aes128Sha256RsaOaep", "Aes256Sha256RsaPss".
	// +optional
	SecurityPolicy string `json:"securityPolicy,omitempty"`
	// Defaults to "None". The value could be "None", "Sign" and "SignAndEncrypt".
	// +optional
	SecurityMode string `json:"securityMode,omitempty"`
	// Certificate file for access opc server.
	// +optional
	Certificate string `json:"certificate,omitempty"`
	// PrivateKey file for access opc server.
	// +optional
	PrivateKey string `json:"privateKey,omitempty"`
	// Timeout seconds for the opc server connection.???
	// +optional
	Timeout int64 `json:"timeout,omitempty"`
}

// ProtocolCommonConfigOPCUA is the opc-ua configuration.
type ProtocolCommonConfigOPCUA struct {
	CustomizedValues CustomizedValue `json:"customizedValues,omitempty"`
}

// CustomizedValue is the customized part for opc-ua protocol.
type CustomizedValue map[string]interface{}
