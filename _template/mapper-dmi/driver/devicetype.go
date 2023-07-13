package driver

import (
	"sync"

	"github.com/kubeedge/mappers-go/pkg/common"
)

// CustomizedDev is the customized device configuration and client information.
type CustomizedDev struct {
	Instance         common.DeviceInstance
	CustomizedClient *CustomizedClient
}

type CustomizedClient struct {
	intMaxValue int
	deviceMutex sync.Mutex
	TemplateProtocolCommonConfig
	TemplateProtocolConfig
}

type TemplateProtocolConfig struct {
	ProtocolName       string `json:"protocolName"`
	ProtocolConfigData `json:"configData"`
}

type ProtocolConfigData struct {
	// TODO: add your config data according to configmap
}

type TemplateProtocolCommonConfig struct {
	CommonCustomizedValues `json:"customizedValues"`
}

type CommonCustomizedValues struct {
	// TODO: add your CommonCustomizedValues according to configmap
}
type TemplateVisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	// TODO: add your Visitor ConfigData according to configmap
}
