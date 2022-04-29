//Package models defines the interfaces that developer need to implement
package models

// ProtocolDriver is a low-level device-specific interface used by other
// components of Mapper to interact with a specific class of devices.
type ProtocolDriver interface {
	// InitDevice provide configmap parsing to specific protocols
	InitDevice(protocolCommon []byte) (err error)

	// ReadDeviceData  is an interface that reads data from a specific device, data is a type of string
	ReadDeviceData(protocolCommon, visitor, protocol []byte) (data interface{}, err error)

	// WriteDeviceData is an interface that write data to a specific device, data's DataType is Consistent with configmap
	WriteDeviceData(data interface{}, protocolCommon, visitor, protocol []byte) (err error)

	// StopDevice is an interface to stop all devices
	StopDevice() (err error)

	// GetDeviceStatus is an interface to get the device status true is OK , false is DISCONNECTED
	GetDeviceStatus(protocolCommon, visitor, protocol []byte) (status bool)
}
