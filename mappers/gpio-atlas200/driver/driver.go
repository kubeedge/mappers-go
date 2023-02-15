package driver

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type GPIOProtocolConfig struct {
	ProtocolName       string `json:"protocolName"`
	ProtocolConfigData `json:"configData"`
}

type ProtocolConfigData struct {
}

type GPIOProtocolCommonConfig struct {
	CommonCustomizedValues `json:"customizedValues"`
}

type CommonCustomizedValues struct {
}
type GPIOVisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	Pin int `json:"pin"`
}

// GPIO Realize the structure of random number
type GPIO struct {
	mutex                sync.Mutex
	protocolConfig       GPIOProtocolConfig
	protocolCommonConfig GPIOProtocolCommonConfig
	visitorConfig        GPIOVisitorConfig
}

// InitDevice Sth that need to do in the first
// If you need mount a persistent connection, you should provide parameters in configmap's protocolCommon.
// and handle these parameters in the following function
func (d *GPIO) InitDevice(protocolCommon []byte) (err error) {
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &d.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return err
		}
	}
	fmt.Println("GPIO devices do not need to be initialized")
	return nil
}

// SetConfig Parse the configmap's raw json message
func (d *GPIO) SetConfig(protocolCommon, visitor, protocol []byte) (pin int, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &d.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return 0, err
		}
	}
	if visitor != nil {
		if err = json.Unmarshal(visitor, &d.visitorConfig); err != nil {
			fmt.Printf("Unmarshal visitorConfig error: %v\n", err)
			return 0, err
		}
	}
	if protocol != nil {
		if err = json.Unmarshal(protocol, &d.protocolConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolConfig error: %v\n", err)
			return 0, err
		}
	}
	return d.visitorConfig.Pin, nil
}

// ReadDeviceData  is an interface that reads data from a specific device, data is a type of string
func (d *GPIO) ReadDeviceData(protocolCommon, visitor, protocol []byte) (data interface{}, err error) {
	// Parse raw json message to get a virtualDevice instance
	pin, err := d.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return nil, err
	}
	pinClient := Pin(pin)
	if pinClient.Read() == '0' {
		return "OFF", nil
	}
	return "ON", nil
}

// WriteDeviceData is an interface that write data to a specific device, data's DataType is Consistent with configmap
func (d *GPIO) WriteDeviceData(data interface{}, protocolCommon, visitor, protocol []byte) (err error) {
	// Parse raw json message to get a virtualDevice instance
	pin, err := d.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return err
	}
	status := data.(string)
	pinClient := Pin(pin)
	if strings.ToUpper(status) == "OFF" {
		pinClient.SetOutPut()
		pinClient.SetLow()

	} else if strings.ToUpper(status) == "ON" {
		pinClient.SetOutPut()
		pinClient.SetHight()
	} else {
		fmt.Println("the command should be \"ON\" or \"OFF\"")
	}
	return nil
}

// StopDevice is an interface to disconnect a specific device
// This function is called when mapper stops serving
func (d *GPIO) StopDevice() (err error) {
	// in this func, u can get ur device-instance in the client map, and give a safety exit
	fmt.Println("----------Stop gpio Device Successful----------")
	return nil
}

// GetDeviceStatus is an interface to get the device status true is OK , false is DISCONNECTED
func (d *GPIO) GetDeviceStatus(protocolCommon, visitor, protocol []byte) (status bool) {
	err := Open()
	defer Close()
	if err != nil {
		return false
	}
	return true
}
