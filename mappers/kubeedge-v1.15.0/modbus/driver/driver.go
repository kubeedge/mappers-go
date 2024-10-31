package driver

import (
	"errors"
	"sync"
	"time"

	"github.com/sailorvii/modbus"
	"k8s.io/klog/v2"
)

var clients *sync.Map

var clientInit sync.Once

func initMap() {
	clientInit.Do(func() {
		if clients == nil {
			clients = new(sync.Map)
		}
	})
}

func NewClient(protocol ProtocolConfig) (*CustomizedClient, error) {
	modbusProtocolConfig := ModbusProtocolConfig{
		SlaveID:    protocol.SlaveID,
		SerialPort: protocol.SerialPort,
		BaudRate:   protocol.BaudRate,
		DataBits:   protocol.DataBits,
		StopBits:   protocol.StopBits,
		Parity:     protocol.Parity,
		Timeout:    time.Duration(protocol.Timeout),
	}
	client := &CustomizedClient{
		ProtocolConfig:       protocol,
		deviceMutex:          sync.Mutex{},
		ModbusProtocolConfig: modbusProtocolConfig,
	}
	return client, nil
}

func (c *CustomizedClient) InitDevice() error {
	// TODO: add init operation
	// you can use c.ProtocolConfig
	initMap()
	klog.Infoln("SerialPort : ", c.ModbusProtocolConfig.SerialPort)
	v, ok := clients.Load(c.ModbusProtocolConfig.SerialPort)
	if ok {
		c.ModbusClient = v.(modbus.Client)
		return nil
	}

	handler := modbus.NewRTUClientHandler(c.ModbusProtocolConfig.SerialPort)
	handler.BaudRate = c.ModbusProtocolConfig.BaudRate
	handler.DataBits = c.ModbusProtocolConfig.DataBits
	handler.Parity = parity(c.ModbusProtocolConfig.Parity)
	handler.StopBits = c.ModbusProtocolConfig.StopBits
	handler.SlaveId = c.ModbusProtocolConfig.SlaveID
	handler.Timeout = c.ModbusProtocolConfig.Timeout
	handler.IdleTimeout = c.ModbusProtocolConfig.Timeout
	client := modbus.NewClient(handler)
	clients.Store(c.ModbusProtocolConfig.SerialPort, &client)
	c.ModbusClient = client

	return nil
}

func (c *CustomizedClient) GetDeviceData(visitor *VisitorConfig) (interface{}, error) {
	// TODO: add the code to get device's data
	// you can use c.ProtocolConfig and visitor
	c.deviceMutex.Lock()
	defer c.deviceMutex.Unlock()

	var results []byte
	var err error
	switch visitor.Register {
	case "CoilRegister":
		results, err = c.ModbusClient.ReadCoils(visitor.Offset, uint16(visitor.Limit))
	case "DiscreteInputRegister":
		results, err = c.ModbusClient.ReadDiscreteInputs(visitor.Offset, uint16(visitor.Limit))
	case "HoldingRegister":
		results, err = c.ModbusClient.ReadHoldingRegisters(visitor.Offset, uint16(visitor.Limit))
	case "InputRegister":
		results, err = c.ModbusClient.ReadInputRegisters(visitor.Offset, uint16(visitor.Limit))
	default:
		return nil, errors.New("Bad register type")
	}
	klog.V(2).Info("Get result: ", results)
	return results, err
}

func (c *CustomizedClient) SetDeviceData(data interface{}, visitor *VisitorConfig) error {
	// TODO: set device's data
	// you can use c.ProtocolConfig and visitor
	var results []byte
	var err error

	c.deviceMutex.Lock()
	defer c.deviceMutex.Unlock()

	klog.V(1).Info("Set:", visitor.Register, visitor.Offset, uint16(visitor.Limit))

	switch visitor.Register {
	case "CoilRegister":
		var valueSet uint16
		switch uint16(visitor.Limit) {
		case 0:
			valueSet = 0x0000
		case 1:
			valueSet = 0xFF00
		default:
			return errors.New("Wrong value")
		}
		results, err = c.ModbusClient.WriteSingleCoil(visitor.Offset, valueSet)
	case "HoldingRegister":
		results, err = c.ModbusClient.WriteSingleRegister(visitor.Offset, uint16(visitor.Limit))
	default:
		return errors.New("Bad register type")
	}
	klog.V(1).Info("Set result:", err, results)
	return nil
}

func (c *CustomizedClient) StopDevice() error {
	// TODO: stop device
	// you can use c.ProtocolConfig
	err := c.ModbusClient.Close()
	if err != nil {
		return err
	}
	return nil
}

// parity convert into the format that modbus driver requires.
func parity(ori string) string {
	var p string
	switch ori {
	case "even":
		p = "E"
	case "odd":
		p = "O"
	default:
		p = "N"
	}
	return p
}
