package dmi

import (
	dmiapi "github.com/kubeedge/mappers-go/pkg/apis/dmi/v1"
)

// DeviceManagerService interface should be implemented by a device mapper or edgecore.
// The methods should be thread-safe.
type DeviceManagerService interface {
	DeviceMapperManager
}

// DeviceMapperManager contains methods for mapper name, version and API version.
type DeviceMapperManager interface {
	// GetMapper returns the device mapper name, device mapper version and device mapper API version
	//GetMapper(mapperName string) (*dmiapi.MapperInfo, error)
	//
	//HealthCheck(mapperName string) (string, error)

	MapperRegister(mapper *dmiapi.MapperInfo) error
}
