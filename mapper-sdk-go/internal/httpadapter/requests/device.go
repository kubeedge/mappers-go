// Package requests used to call define add device request's struct
package requests

import "github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"

// AddDeviceRequest the struct of device request
type AddDeviceRequest struct {
	DeviceInstance *configmap.DeviceInstance `json:"deviceInstances"`
	DeviceModels   []*configmap.DeviceModel  `json:"deviceModels"`
	Protocol       []*configmap.Protocol     `json:"protocols"`
}
