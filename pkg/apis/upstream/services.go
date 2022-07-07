package upstream

import (
	dmiapi "github.com/kubeedge/kubeedge/edge/pkg/apis/dmi/upstream/v1"
)

// DeviceManagerService interface should be implemented by a device mapper or edgecore.
// The methods should be thread-safe.
type DeviceManagerService interface {
	DeviceMapperManager
	DeviceManager
}

// DeviceMapperManager contains methods for mapper name, version and API version.
type DeviceMapperManager interface {
	// GetMapper returns the device mapper name, device mapper version and device mapper API version
	MapperRegister(mapper *dmiapi.MapperInfo) error
}

// DeviceManager contains methods to manipulate devices managed by a
// device mapper. The methods are thread-safe.
type DeviceManager interface {
	// ReportDeviceStatus updates the reported status of the device from mapper.
	ReportDeviceStatus(deviceName string, reportedDevice *dmiapi.DeviceStatus) error
}
