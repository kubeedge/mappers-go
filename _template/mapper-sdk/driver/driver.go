package driver

import (
	"encoding/json"
	"fmt"
	"sync"
)

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

// Template Realize the structure
type Template struct {
	// TODO: add some vars to help your implement the SDK interfaces
	mutex sync.Mutex
	protocolConfig TemplateProtocolConfig
	protocolCommonConfig  TemplateProtocolCommonConfig
	visitorConfig         TemplateVisitorConfig
}

// InitDevice Sth that need to do in the first
// If you need mount a persistent connection, you should provide parameters in configmap's protocolCommon.
// and handle these parameters in the following function
func (d *Template) InitDevice(protocolCommon []byte) (err error) {
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &d.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return err
		}
	}
	// TODO: add init operation
	return nil
}

// SetConfig Parse the configmap's raw json message
// In the case of high concurrency, d.mutex helps you get the correct value
func (d *Template) SetConfig(protocolCommon, visitor, protocol []byte) (err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &d.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal protocolCommonConfig error: %v\n", err)
			return err
		}
	}
	if visitor != nil {
		if err = json.Unmarshal(visitor, &d.visitorConfig); err != nil {
			fmt.Printf("Unmarshal visitorConfig error: %v\n", err)
			return err
		}

	}
	if protocol != nil {
		if err = json.Unmarshal(protocol, &d.protocolConfig); err != nil {
			fmt.Printf("Unmarshal protocolConfig error: %v\n", err)
			return err
		}
	}
	return
}

// ReadDeviceData  is an interface that reads data from a specific device, data's dataType is consistent with configmap
func (d *Template) ReadDeviceData(protocolCommon, visitor, protocol []byte) (data interface{}, err error) {
	// Parse raw json message to get a Template instance
	err = d.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return nil, err
	}
	// TODO: get device's data by protocolCommonConfig,visitorConfig,protocolConfig
	return nil, err
}

// WriteDeviceData is an interface that write data to a specific device, data's dataType is consistent with configmap
func (d *Template) WriteDeviceData(data interface{}, protocolCommon, visitor, protocol []byte) (err error) {
	// Parse raw json message to get a Template instance
	err = d.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return err
	}
	// TODO: set device's config to data interface{}  by protocolCommonConfig,visitorConfig,protocolConfig
	return nil
}

// StopDevice is an interface to disconnect a specific device
// This function is called when mapper stops serving
func (d *Template) StopDevice() (err error) {
	// TODO: If you need to exit safely, set the exit operation, otherwise it can be ignored
	fmt.Println("----------Stop Template Successful----------")
	return nil
}



// GetDeviceStatus is an interface to get the device status true is OK , false is DISCONNECTED
func (d *Template) GetDeviceStatus(protocolCommon, visitor, protocol []byte) (status bool) {
	err := d.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return false
	}
	// TODO: get device's status by protocolCommonConfig,visitorConfig,protocolConfig
	return true
}
