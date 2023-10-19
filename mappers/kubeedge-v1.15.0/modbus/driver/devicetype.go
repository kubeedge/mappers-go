package driver

import (
	"sync"
	"time"

	"github.com/sailorvii/modbus"

	"github.com/kubeedge/modbus/pkg/common"
)

// CustomizedDev is the customized device configuration and client information.
type CustomizedDev struct {
	Instance         common.DeviceInstance
	CustomizedClient *CustomizedClient
}

type CustomizedClient struct {
	// TODO add some variables to help you better implement device drivers
	deviceMutex sync.Mutex
	ProtocolConfig
	ModbusProtocolConfig
	ModbusClient modbus.Client
}

type ProtocolConfig struct {
	ProtocolName string `json:"protocolName"`
	ConfigData   `json:"configData"`
}

type ConfigData struct {
	// TODO: add your protocol config data
	SlaveID    byte
	SerialPort string
	BaudRate   int
	DataBits   int
	StopBits   int
	Parity     string
	Timeout    int
}

type ModbusProtocolConfig struct {
	SlaveID    byte
	SerialPort string
	BaudRate   int
	DataBits   int
	StopBits   int
	Parity     string
	Timeout    time.Duration
}

type VisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	// TODO: add your visitor config data
	DataType string `json:"dataType"`
	Register       string  `json:"register"`
	Offset         uint16  `json:"offset"`
	Limit          int     `json:"limit"`
	Scale          float64 `json:"scale,omitempty"`
	IsSwap         bool    `json:"isSwap,omitempty"`
	IsRegisterSwap bool    `json:"isRegisterSwap,omitempty"`
}
