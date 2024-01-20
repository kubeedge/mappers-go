package driver

import (
	"sync"
)

func NewClient(commonProtocol TemplateProtocolCommonConfig,
	protocol TemplateProtocolConfig) (*CustomizedClient, error) {
	client := &CustomizedClient{
		TemplateProtocolCommonConfig: commonProtocol,
		TemplateProtocolConfig:       protocol,
		deviceMutex:                  sync.Mutex{},
	}
	return client, nil
}

func (c *CustomizedClient) InitDevice() error {
	// TODO: add init operation
	// you can use c.TemplateProtocolConfig and c.TemplateProtocolCommonConfig
	return nil
}

func (c *CustomizedClient) GetDeviceData(visitor *TemplateVisitorConfig) (interface{}, error) {
	// TODO: get device's data
	// you can use c.TemplateProtocolConfig,c.TemplateProtocolCommonConfig and visitor
	return nil, nil
}

func (c *CustomizedClient) SetDeviceData(data interface{}, visitor *TemplateVisitorConfig) error {
	// TODO: set device's data
	// you can use c.TemplateProtocolConfig,c.TemplateProtocolCommonConfig and visitor
	return nil
}

func (c *CustomizedClient) StopDevice() error {
	// TODO: stop device
	// you can use c.TemplateProtocolConfig and c.TemplateProtocolCommonConfig
	return nil
}
