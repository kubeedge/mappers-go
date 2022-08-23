package onvif

import (
	"github.com/kubeedge/mappers-go/pkg/common"
)

// OnvifDev is the Onvif device configuration and client information.
type OnvifDev struct {
	Instance    common.DeviceInstance
	OnvifClient *OnvifClient
}

// OnvifVisitorConfig is the Onvif register configuration.
type OnvifVisitorConfig struct {
	ProtocolName string          `json:"protocolName"`
	ConfigData   CustomizedValue `json:"configData"`
}

// OnvifProtocolConfig is the protocol configuration.
type OnvifProtocolConfig struct {
	ProtocolName string          `json:"protocolName"`
	ConfigData   CustomizedValue `json:"configData"`
}

// OnvifProtocolCommonConfig is the Onvif protocol configuration.
type OnvifProtocolCommonConfig struct {
}

// CustomizedValue is the customized part for Onvif protocol.
type CustomizedValue map[string]interface{}
