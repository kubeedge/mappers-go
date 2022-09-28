package global

import (
	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
)

type DevPanel interface {
	DevStart()
	DevInit(cfg *config.Config) error
	UpdateDev(model *common.DeviceModel, device *common.DeviceInstance, protocol *common.Protocol)
	UpdateDevTwins(deviceID string, twins []common.Twin) error
	DealDeviceTwinGet(deviceID string, twinName string) (interface{}, error)
	GetDevice(deviceID string) (interface{}, error)
	RemoveDevice(deviceID string) error
	GetModel(modelName string) (common.DeviceModel, error)
	UpdateModel(model *common.DeviceModel)
	RemoveModel(modelName string)
}
