package instancepool

import (
	"context"
	"sync"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/models"
)

// DeviceInstancesName contains the name of device instance struct in the DIC.
var DeviceInstancesName = di.TypeInstanceToName(map[string]configmap.DeviceInstance{})

// DeviceModelsName contains the name of device model struct in the DIC.
var DeviceModelsName = di.TypeInstanceToName(map[string]configmap.DeviceModel{})

// ProtocolName contains the name of device protocol struct in the DIC.
var ProtocolName = di.TypeInstanceToName(map[string]configmap.Protocol{})

// ProtocolDriverName contains the name of device driver struct in the DIC.
var ProtocolDriverName = di.TypeInstanceToName((*models.ProtocolDriver)(nil))

// WgName contains the name of waitGroup struct in the DIC.
var WgName = di.TypeInstanceToName((*sync.WaitGroup)(nil))

// MutexName contains the name of mutex struct in the DIC.
var MutexName = di.TypeInstanceToName((*sync.Mutex)(nil))

// StopFunctionsName contains the name of device cancelFunc struct in the DIC.
var StopFunctionsName = di.TypeInstanceToName(map[string]context.CancelFunc(nil))

// ConnectInfoName contains the name of device connectInfo struct in the DIC.
var ConnectInfoName = di.TypeInstanceToName(map[string]configmap.ConnectInfo{})

// DeviceLockName contains the name of device lock struct in the DIC.
var DeviceLockName = di.TypeInstanceToName(map[string]common.Lock{})

// DeviceInstancesNameFrom helper function queries the DIC and returns device instance struct.
func DeviceInstancesNameFrom(get di.Get) map[string]*configmap.DeviceInstance {
	return get(DeviceInstancesName).(map[string]*configmap.DeviceInstance)
}

// DeviceModelsNameFrom helper function queries the DIC and returns device model struct.
func DeviceModelsNameFrom(get di.Get) map[string]*configmap.DeviceModel {
	return get(DeviceModelsName).(map[string]*configmap.DeviceModel)
}

// ProtocolNameFrom helper function queries the DIC and returns device protocol struct.
func ProtocolNameFrom(get di.Get) map[string]*configmap.Protocol {
	return get(ProtocolName).(map[string]*configmap.Protocol)
}

// ProtocolDriverNameFrom helper function queries the DIC and returns device driver struct.
func ProtocolDriverNameFrom(get di.Get) models.ProtocolDriver {
	return get(ProtocolDriverName).(models.ProtocolDriver)
}

// WgNameFrom helper function queries the DIC and returns waitGroup struct.
func WgNameFrom(get di.Get) *sync.WaitGroup {
	return get(WgName).(*sync.WaitGroup)
}

// MutexNameFrom helper function queries the DIC and returns mutex struct.
func MutexNameFrom(get di.Get) *sync.Mutex {
	return get(MutexName).(*sync.Mutex)
}

// StopFunctionsNameFrom helper function queries the DIC and returns device cancelFunc struct.
func StopFunctionsNameFrom(get di.Get) map[string]context.CancelFunc {
	return get(StopFunctionsName).(map[string]context.CancelFunc)
}

// ConnectInfoNameFrom helper function queries the DIC and returns device connectInfo struct.
func ConnectInfoNameFrom(get di.Get) map[string]*configmap.ConnectInfo {
	return get(ConnectInfoName).(map[string]*configmap.ConnectInfo)
}

// DeviceLockNameFrom helper function queries the DIC and returns device lock struct.
func DeviceLockNameFrom(get di.Get) map[string]*common.Lock {
	return get(DeviceLockName).(map[string]*common.Lock)
}
