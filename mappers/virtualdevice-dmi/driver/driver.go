package driver

import (
	"fmt"
	"math/rand"
	"sync"

	"k8s.io/klog/v2"
)

func NewClient(commonProtocol CustomizedDeviceProtocolCommonConfig,
	protocol CustomizedDeviceProtocolConfig) (*CustomizedClient, error) {
	client := &CustomizedClient{
		CustomizedDeviceProtocolCommonConfig: commonProtocol,
		CustomizedDeviceProtocolConfig:       protocol,
		deviceMutex:                          sync.Mutex{},
	}
	return client, nil
}

func (c *CustomizedClient) InitDevice() error {
	// If your devices need to be initialized, do it here.
	klog.Infof("Init device%d successful, protocolID: %v", c.DeviceID, c.ProtocolID)
	klog.Infof("I can get Info: %v %v ", c.Com.SerialPort, c.Com.BaudRate)
	return nil
}

func (c *CustomizedClient) GetDeviceData(visitor *CustomizedDeviceVisitorConfig) (interface{}, error) {
	if visitor.DataType == "int" {
		if c.intMaxValue <= 0 {
			return nil, fmt.Errorf("max value is %d, should > 0", c.intMaxValue)
		}
		return rand.Intn(c.intMaxValue), nil
	} else if visitor.DataType == "float" {
		return rand.Float64(), nil
	} else {
		return nil, fmt.Errorf("unrecognized data type: %s", visitor.DataType)
	}
}

func (c *CustomizedClient) SetDeviceData(data interface{}, visitor *CustomizedDeviceVisitorConfig) error {
	if visitor.DataType == "int" {
		c.intMaxValue = int(data.(int64))
	} else {
		return fmt.Errorf("unrecognized data type: %s", visitor.DataType)
	}
	return nil
}

func (c *CustomizedClient) StopDevice() error {
	klog.Infof("Stop device%d successful", c.DeviceID)
	return nil
}

func (c *CustomizedClient) GetDeviceStatus() {
	// TODO health check
}
