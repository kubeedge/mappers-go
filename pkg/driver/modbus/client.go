/*
Copyright 2020 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package modbus

import (
	"errors"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/sailorvii/modbus"

	"github.com/kubeedge/mappers-go/pkg/common"
)

// ModbusTCP is the configurations of modbus TCP.
type ModbusTCP struct {
	SlaveID  byte
	DeviceIP string
	TCPPort  string
	Timeout  time.Duration
}

// ModbusRTU is the configurations of modbus RTU.
type ModbusRTU struct {
	SlaveID      byte
	SerialName   string
	BaudRate     int
	DataBits     int
	StopBits     int
	Parity       string
	RS485Enabled bool
	Timeout      time.Duration
}

// ModbusClient is the structure for modbus client.
type ModbusClient struct {
	Client  modbus.Client
	Handler interface{}
	Config  interface{}

	mu sync.Mutex
}

/*
* In modbus RTU mode, devices could connect to one serial port on RS485. However,
* the serial port doesn't support paralleled visit, and for one tcp device, it also doesn't support
* paralleled visit, so we expect one client for one port.
 */
var clients *sync.Map

var clientInit sync.Once

func initMap() {
	clientInit.Do(func() {
		if clients == nil {
			clients = new(sync.Map)
		}
	})
}

func newTCPClient(config ModbusTCP) *ModbusClient {
	initMap()

	addr := config.DeviceIP + ":" + config.TCPPort
	slave := addr + "/" + string(config.SlaveID)
	klog.Infoln("slave id: ", config.SlaveID)
	v, ok := clients.Load(slave)
	if ok {
		return v.(*ModbusClient)
	}

	handler := modbus.NewTCPClientHandler(addr)
	handler.Timeout = config.Timeout
	handler.IdleTimeout = config.Timeout
	handler.SlaveId = config.SlaveID

	client := ModbusClient{Client: modbus.NewClient(handler), Handler: handler, Config: config}
	clients.Store(slave, &client)
	return &client
}

func newRTUClient(config ModbusRTU) *ModbusClient {
	initMap()

	klog.Infoln("SerialName : ", config.SerialName)
	v, ok := clients.Load(config.SerialName)
	if ok {
		return v.(*ModbusClient)
	}

	handler := modbus.NewRTUClientHandler(config.SerialName)
	handler.BaudRate = config.BaudRate
	handler.DataBits = config.DataBits
	handler.Parity = parity(config.Parity)
	handler.StopBits = config.StopBits
	handler.SlaveId = config.SlaveID
	handler.Timeout = config.Timeout
	handler.IdleTimeout = config.Timeout
	handler.RS485.Enabled = config.RS485Enabled
	client := ModbusClient{Client: modbus.NewClient(handler), Handler: handler, Config: config}
	clients.Store(config.SerialName, &client)
	return &client
}

// NewClient allocate and return a modbus client.
// Client type includes TCP and RTU.
func NewClient(config interface{}) (*ModbusClient, error) {
	switch c := config.(type) {
	case ModbusTCP:
		return newTCPClient(c), nil
	case ModbusRTU:
		return newRTUClient(c), nil
	default:
		return &ModbusClient{}, errors.New("Wrong modbus type")
	}
}

// GetStatus get device status.
// Now we could only get the connection status.
func (c *ModbusClient) GetStatus() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.Client.Connect()
	if err == nil {
		return common.DEVSTOK
	}
	return common.DEVSTDISCONN
}

// Get get register.
func (c *ModbusClient) Get(registerType string, addr uint16, quantity uint16) (results []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch registerType {
	case "CoilRegister":
		results, err = c.Client.ReadCoils(addr, quantity)
	case "DiscreteInputRegister":
		results, err = c.Client.ReadDiscreteInputs(addr, quantity)
	case "HoldingRegister":
		results, err = c.Client.ReadHoldingRegisters(addr, quantity)
	case "InputRegister":
		results, err = c.Client.ReadInputRegisters(addr, quantity)
	default:
		return nil, errors.New("bad register type")
	}
	klog.V(2).Info("Get result: ", results)
	return results, err
}

// Set set register.
func (c *ModbusClient) Set(registerType string, addr uint16, value uint16) (results []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	klog.V(1).Info("Set:", registerType, addr, value)

	switch registerType {
	case "CoilRegister":
		var valueSet uint16
		switch value {
		case 0:
			valueSet = 0x0000
		case 1:
			valueSet = 0xFF00
		default:
			return nil, errors.New("wrong value")
		}
		results, err = c.Client.WriteSingleCoil(addr, valueSet)
	case "HoldingRegister":
		results, err = c.Client.WriteSingleRegister(addr, value)
	default:
		return nil, errors.New("bad register type")
	}
	klog.V(1).Info("Set result:", err, results)
	return results, err
}

// SetString set string.
func (c *ModbusClient) SetString(registerType string, offset uint16, limit int, value string) (results []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	klog.V(1).InfoS("ModbusClient Set:", "register", registerType, "offset", offset, "limit", limit, "value", value)

	switch registerType {
	case "CoilRegister":
		var valueSet uint16
		switch value {
		case "0":
			valueSet = 0x0000
		case "1":
			valueSet = 0xFF00
		default:
			return nil, errors.New("wrong value")
		}
		results, err = c.Client.WriteSingleCoil(offset, valueSet)
	case "HoldingRegister":
		valueBytes := make([]byte, limit*2)
		copy(valueBytes, value)
		results, err = c.Client.WriteMultipleRegisters(offset, uint16(limit), valueBytes)
		if err != nil {
			klog.ErrorS(err, "Failed to set HoldingRegister", "offset", offset, "limit", limit, "value", value)
		}
	default:
		return nil, errors.New("bad register type")
	}
	klog.V(1).InfoS("ModbusClient Set result", "results", results)
	return results, err
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

// Reconnect close the connection and reconnect to device.
func (c *ModbusClient) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.Client.Close()
	if err != nil {
		klog.Errorf("fail to close the modbus connection with error: %+v", err)
		return err
	}

	err = c.Client.Connect()
	if err != nil {
		klog.Errorf("fail to connect the modbus connection with error: %+v", err)
		return err
	}

	return nil
}

// GetWithRetry get register with retry.
func (c *ModbusClient) GetWithRetry(registerType string, addr uint16, quantity uint16, retryTime int) (results []byte, err error) {
	for i := 0; i < retryTime; i++ {
		results, err = c.Get(registerType, addr, quantity)
		if err == nil {
			return results, nil
		}

		klog.Warningf("fail to get register, reconnect and retry")
		errConn := c.Reconnect()
		if errConn != nil {
			klog.Errorf("fail to reconnect with error: %+v", errConn)
			continue
		}
	}
	return results, err
}

// SetWithRetry set register with retry.
func (c *ModbusClient) SetWithRetry(registerType string, addr uint16, value uint16, retryTime int) (results []byte, err error) {
	for i := 0; i < retryTime; i++ {
		results, err = c.Set(registerType, addr, value)
		if err == nil {
			return results, nil
		}

		klog.Warningf("fail to set register, reconnect and retry")
		errConn := c.Reconnect()
		if errConn != nil {
			klog.Errorf("fail to reconnect with error: %+v", errConn)
			continue
		}
	}
	return results, err
}

// SetStringWithRetry set string with retry.
func (c *ModbusClient) SetStringWithRetry(registerType string, offset uint16, limit int, value string, retryTime int) (results []byte, err error) {
	for i := 0; i < retryTime; i++ {
		results, err = c.SetString(registerType, offset, limit, value)
		if err == nil {
			return results, nil
		}

		klog.Warningf("fail to set string, reconnect and retry")
		errConn := c.Reconnect()
		if errConn != nil {
			klog.Errorf("fail to reconnect with error: %+v", errConn)
			continue
		}
	}
	return results, err
}
