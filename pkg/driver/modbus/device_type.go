package modbus

import (
	"github.com/kubeedge/mappers-go/pkg/common"
)

// ModbusDev is the modbus device configuration and client information.
type ModbusDev struct {
	Instance     common.DeviceInstance
	ModbusClient *ModbusClient
}

// ModbusVisitorConfig is the modbus register configuration.
type ModbusVisitorConfig struct {
	Register       string  `json:"register"`
	Offset         uint16  `json:"offset"`
	Limit          int     `json:"limit"`
	Scale          float64 `json:"scale,omitempty"`
	IsSwap         bool    `json:"isSwap,omitempty"`
	IsRegisterSwap bool    `json:"isRegisterSwap,omitempty"`
}

// ModbusProtocolConfig is the protocol configuration.
type ModbusProtocolConfig struct {
	SlaveID int16 `json:"slaveID,omitempty"`
}

// ModbusProtocolCommonConfig is the modbus protocol configuration.
type ModbusProtocolCommonConfig struct {
	COM              COMStruct       `json:"com,omitempty"`
	TCP              TCPStruct       `json:"tcp,omitempty"`
	CustomizedValues CustomizedValue `json:"customizedValues,omitempty"`
}

// CustomizedValue is the customized part for modbus protocol.
type CustomizedValue map[string]interface{}

// COMStruct is the serial configuration.
type COMStruct struct {
	SerialPort string `json:"serialPort"`
	BaudRate   int64  `json:"baudRate"`
	DataBits   int64  `json:"dataBits"`
	Parity     string `json:"parity"`
	StopBits   int64  `json:"stopBits"`
}

// TCPStruct is the TCP configuration.
type TCPStruct struct {
	IP   string `json:"ip"`
	Port int64  `json:"port"`
}
