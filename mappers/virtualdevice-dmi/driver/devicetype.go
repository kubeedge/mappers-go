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
	CustomizedDeviceProtocolCommonConfig
	CustomizedDeviceProtocolConfig
}

type CustomizedDeviceProtocolConfig struct {
	ProtocolName       string `json:"protocolName"`
	ProtocolConfigData `json:"configData"`
}

type ProtocolConfigData struct {
	DeviceID int `json:"deviceID,omitempty"`
}

type CustomizedDeviceProtocolCommonConfig struct {
	Com                    `json:"com"`
	CommonCustomizedValues `json:"customizedValues"`
}

type Com struct {
	SerialPort string `json:"serialPort"`
	DataBits   int    `json:"dataBits"`
	BaudRate   int    `json:"baudRate"`
	Parity     string `json:"parity"`
	StopBits   int    `json:"stopBits"`
}

type CommonCustomizedValues struct {
	ProtocolID int `json:"protocolID"`
}

type CustomizedDeviceVisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	DataType string `json:"dataType"`
}
