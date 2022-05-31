package parse

import "github.com/kubeedge/kubeedge/cloud/pkg/apis/devices/v1alpha2"

type DeviceData struct {
	Device      *v1alpha2.Device      `json:"device"`
	DeviceModel *v1alpha2.DeviceModel `json:"device_model"`
}
