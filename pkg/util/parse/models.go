package parse

import "github.com/kubeedge/kubeedge/pkg/apis/devices/v1alpha2"

type DeviceData struct {
	Device      *v1alpha2.Device      `json:"device"`
	DeviceModel *v1alpha2.DeviceModel `json:"device_model"`
}

type TwinResultResponse struct {
	PropertyName string `json:"property_name"`
	Payload      []byte `json:"payload"`
}
