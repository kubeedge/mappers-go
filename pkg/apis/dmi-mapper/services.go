package dmi_mapper

import (
	dmiapi "github.com/kubeedge/mappers-go/pkg/apis/dmi-mapper/v1"
)

type DeviceManagerService interface {
	DeviceManager
}

type DeviceManager interface {

	// CreateDevice creates a new device.
	CreateDevice(config *dmiapi.DeviceConfig) (string, error)

	// RemoveDevice removes the device from platform by device name.
	RemoveDevice(deviceName string) error

	// UpdateDevice update device meta data.
	UpdateDevice(deviceName string, config *dmiapi.DeviceConfig) error

	// UpdateDeviceStatus updates the status of the device.
	UpdateDeviceStatus(deviceName string, desiredDevice *dmiapi.DeviceStatus) error

	// GetDevice returns the status of the device.
	GetDevice(deviceName string) (*dmiapi.DeviceStatus, error)
}
