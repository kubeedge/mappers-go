package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type VirtualDeviceProtocolConfig struct {
	ProtocolName       string `json:"protocolName"`
	ProtocolConfigData `json:"configData"`
}

type ProtocolConfigData struct {
	DeviceID int `json:"deviceID,omitempty"`
}

type VirtualDeviceProtocolCommonConfig struct {
	CommonCustomizedValues `json:"customizedValues"`
}

type CommonCustomizedValues struct {
	ProtocolID int `json:"protocolID"`
}
type VirtualDeviceVisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	DataType string `json:"dataType"`
}

// VirtualDevice Realize the structure of random number
type VirtualDevice struct {
	mutex                 sync.Mutex
	virtualProtocolConfig VirtualDeviceProtocolConfig
	protocolCommonConfig  VirtualDeviceProtocolCommonConfig
	visitorConfig         VirtualDeviceVisitorConfig
	client                map[int]int64
}

// InitDevice Sth that need to do in the first
// If you need mount a persistent connection, you should provIDe parameters in configmap's protocolCommon.
// and handle these parameters in the following function
func (vd *VirtualDevice) InitDevice(protocolCommon []byte) (err error) {
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &vd.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return err
		}
	}
	fmt.Printf("InitDevice%d...\n", vd.protocolCommonConfig.ProtocolID)
	return nil
}

// SetConfig Parse the configmap's raw json message
func (vd *VirtualDevice) SetConfig(protocolCommon, visitor, protocol []byte) (dataType string, deviceID int, err error) {
	vd.mutex.Lock()
	defer vd.mutex.Unlock()
	vd.NewClient()
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &vd.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return "", 0, err
		}
	}
	if visitor != nil {
		if err = json.Unmarshal(visitor, &vd.visitorConfig); err != nil {
			fmt.Printf("Unmarshal visitorConfig error: %v\n", err)
			return "", 0, err
		}
	}

	if protocol != nil {
		if err = json.Unmarshal(protocol, &vd.virtualProtocolConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolConfig error: %v\n", err)
			return "", 0, err
		}
	}
	dataType = vd.visitorConfig.DataType
	deviceID = vd.virtualProtocolConfig.DeviceID
	return
}

// ReadDeviceData  is an interface that reads data from a specific device, data's dataType is consistent with configmap
func (vd *VirtualDevice) ReadDeviceData(protocolCommon, visitor, protocol []byte) (data interface{}, err error) {
	// Parse raw json message to get a virtualDevice instance
	DataTye, DeviceID, err := vd.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return nil, err
	}
	if DataTye == "int" {
		if vd.client[DeviceID] == 0 {
			return 0, errors.New("vd.limit should not be 0")
		}
		return rand.Intn(int(vd.client[DeviceID])), nil
	} else if DataTye == "float" {
		if vd.client[DeviceID] == 0 {
			return 0, errors.New("vd.limit should not be 0")
		}
		// Simulate device that have time delay
		// time.Sleep(time.Second)
		return rand.Float64(), nil
	} else {
		return "", errors.New("dataType don't exist")
	}
}

// WriteDeviceData is an interface that write data to a specific device, data's dataType is consistent with configmap
func (vd *VirtualDevice) WriteDeviceData(data interface{}, protocolCommon, visitor, protocol []byte) (err error) {
	// Parse raw json message to get a virtualDevice instance
	_, DeviceID, err := vd.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return err
	}
	vd.client[DeviceID] = data.(int64)
	return nil
}

// StopDevice is an interface to disconnect a specific device
// This function is called when mapper stops serving
func (vd *VirtualDevice) StopDevice() (err error) {
	// in this func, u can get ur device-instance in the client map, and give a safety exit
	fmt.Println("----------Stop Virtual Device Successful----------")
	return nil
}

// NewClient create device-instance, if device-instance exit, set the limit as 100.
// Control a group of devices through singleton pattern
func (vd *VirtualDevice) NewClient() {
	if vd.client == nil {
		vd.client = make(map[int]int64)
	}
	if _, ok := vd.client[vd.virtualProtocolConfig.DeviceID]; ok {
		if vd.client[vd.virtualProtocolConfig.DeviceID] == 0 {
			vd.client[vd.virtualProtocolConfig.DeviceID] = 100
		}
	}
}

// GetDeviceStatus is an interface to get the device status true is OK , false is DISCONNECTED
func (vd *VirtualDevice) GetDeviceStatus(protocolCommon, visitor, protocol []byte) (status bool) {
	_, _, err := vd.SetConfig(protocolCommon, visitor, protocol)
	return err == nil
}
