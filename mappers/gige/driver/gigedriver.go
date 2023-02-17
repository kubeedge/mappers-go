package driver

/*
#include <dlfcn.h>
#include <stdlib.h>
int open_device(unsigned int** device,char* deviceId,char** error);
int get_value (unsigned int* device, char* feature, char** value,char** error);
int set_value (unsigned int* device, char* feature, char* value,char** error);
int close_device (unsigned int* device);
#cgo LDFLAGS: -ldl
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
	"sync"
	"time"
	"unsafe"
)

type GigEVisionDeviceProtocolCommonConfig struct {
	CommonCustomizedValues `json:"customizedValues"`
}

type CommonCustomizedValues struct {
	DeviceSN string `json:"deviceSN"`
}

type GigEVisionDeviceVisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	FeatureName string `json:"FeatureName"`
}

type GigEVisionDevice struct {
	mutex                sync.RWMutex
	protocolCommonConfig GigEVisionDeviceProtocolCommonConfig
	visitorConfig        GigEVisionDeviceVisitorConfig
	deviceMeta           map[string]*DeviceMeta
}

type DeviceMeta struct {
	dev              *C.uint
	FeatureName      string
	deviceStatus     bool
	imageFormat      string
	imageURL         string
	ImageTrigger     string
	ImagePostingFlag bool
	maxRetryTimes    int
}

func (gigEClient *GigEVisionDevice) InitDevice(protocolCommon []byte) (err error) {
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &gigEClient.protocolCommonConfig); err != nil {
			klog.Errorf("Unmarshal ProtocolCommonConfig error: %v.", err)
			return err
		}
	}
	err = gigEClient.NewClient(gigEClient.protocolCommonConfig.DeviceSN)
	if err != nil {
		klog.Errorf("Failed to new a GigE client: %v.", err)
		return err
	}
	return nil
}

func (gigEClient *GigEVisionDevice) ParseConfig(protocolCommon, visitor []byte) (deviceSN string, err error) {
	gigEClient.mutex.Lock()
	defer gigEClient.mutex.Unlock()
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &gigEClient.protocolCommonConfig); err != nil {
			klog.Errorf("Unmarshal ProtocolCommonConfig error: %v.", err)
			return "", err
		}
	}
	if visitor != nil {
		if err = json.Unmarshal(visitor, &gigEClient.visitorConfig); err != nil {
			klog.Errorf("Unmarshal visitorConfig error: %v.", err)
			return "", err
		}
	}
	gigEClient.deviceMeta[gigEClient.protocolCommonConfig.DeviceSN].FeatureName = gigEClient.visitorConfig.FeatureName
	return gigEClient.protocolCommonConfig.DeviceSN, nil
}

// ReadDeviceData  is an interface that reads data from a specific device, data is a type of string
func (gigEClient *GigEVisionDevice) ReadDeviceData(protocolCommon, visitor, _ []byte) (data interface{}, err error) {
	deviceSN, err := gigEClient.ParseConfig(protocolCommon, visitor)
	if err != nil {
		return nil, err
	}
	if !gigEClient.deviceMeta[deviceSN].deviceStatus {
		err = fmt.Errorf("device %s is unreachable and failed to read device %s data", deviceSN, gigEClient.deviceMeta[deviceSN].FeatureName)
		return nil, err
	}
	data, err = gigEClient.Get(deviceSN)
	if err != nil {
		return nil, err
	}
	return data, err
}

// WriteDeviceData is an interface that write data to a specific device, data's DataType is Consistent with configmap
func (gigEClient *GigEVisionDevice) WriteDeviceData(data interface{}, protocolCommon, visitor, _ []byte) (err error) {
	deviceSN, err := gigEClient.ParseConfig(protocolCommon, visitor)
	if err != nil {
		return err
	}
	if !gigEClient.deviceMeta[deviceSN].deviceStatus {
		err = fmt.Errorf("device %s is unreachable and failed to get %s", deviceSN, gigEClient.deviceMeta[deviceSN].FeatureName)
		return err
	}
	err = gigEClient.Set(deviceSN, data)
	if err != nil {
		return err
	}
	return nil
}

// StopDevice is an interface to disconnect a specific device
func (gigEClient *GigEVisionDevice) StopDevice() (err error) {
	for s := range gigEClient.deviceMeta {
		if gigEClient.deviceMeta[s].deviceStatus {
			C.close_device(gigEClient.deviceMeta[s].dev)
			gigEClient.deviceMeta[s].dev = nil
			gigEClient.deviceMeta[s].deviceStatus = false
		}
	}
	fmt.Println("----------Stop GigE Device Successful----------")
	return nil
}

// GetDeviceStatus is an interface to get the device status, true is OK , false is DISCONNECTED
func (gigEClient *GigEVisionDevice) GetDeviceStatus(protocolCommon, visitor, _ []byte) (status bool) {
	deviceSN, err := gigEClient.ParseConfig(protocolCommon, visitor)
	if err == nil {
		return false
	}
	return gigEClient.deviceMeta[deviceSN].deviceStatus
}

func (gigEClient *GigEVisionDevice) ReconnectDevice(DeviceSN string) {
	var msg *C.char
	var dev *C.uint
	defer C.free(unsafe.Pointer(msg))
	if gigEClient.deviceMeta[DeviceSN].dev != nil {
		C.close_device(gigEClient.deviceMeta[DeviceSN].dev)
	}
	gigEClient.deviceMeta[DeviceSN].dev = nil
	retryTimes := 0
	for retryTimes < gigEClient.deviceMeta[DeviceSN].maxRetryTimes {
		time.Sleep(5 * time.Second)
		signal := C.open_device(&dev, C.CString(DeviceSN), &msg)
		if signal != 0 {
			klog.Errorf("Failed to restart device %s: %s.", DeviceSN, (string)(C.GoString(msg)))
		} else {
			gigEClient.deviceMeta[DeviceSN].dev = dev
			gigEClient.deviceMeta[DeviceSN].deviceStatus = true
			break
		}
		retryTimes++
	}
	fmt.Printf("Device %s restart success!\n", DeviceSN)
}
