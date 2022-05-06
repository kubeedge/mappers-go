package application

import (
	"context"
	"net/url"
	"sync"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/controller"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/httpadapter/requests"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/instancepool"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/mqttadapter"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
)

// AddDevice internal callback function
func AddDevice(addDeviceRequest requests.AddDeviceRequest, dic *di.Container) common.ErrKind {
	deviceInstance := addDeviceRequest.DeviceInstance
	instanceID := deviceInstance.ID
	deviceInstances := instancepool.DeviceInstancesNameFrom(dic.Get)
	protocols := instancepool.ProtocolNameFrom(dic.Get)
	// Check if the device ID added by the user is duplicated
	for k := range deviceInstances {
		if k == instanceID {
			klog.Errorf("The deviceInstance name: %s used in the uploaded file already exists in mapperService",instanceID)
			return common.KindDuplicateName
		}
	}
	var isFind bool
	// Check whether mapperService already has a protocol
	if _, ok := protocols[deviceInstance.ProtocolName]; !ok {
		if addDeviceRequest.Protocol != nil {
			for i := range addDeviceRequest.Protocol{
				if deviceInstance.ProtocolName == addDeviceRequest.Protocol[i].Name {
					isFind = true
					break
				}
			}
			if !isFind{
				klog.Errorf("http callback error : protocol mismatch , there is no protocol named %s matching the uploaded file",deviceInstance.ProtocolName)
				return common.KindEntityDoesNotExist
			}
		}
	}
	// Check whether mapperService already has a deviceModel
	deviceModels := instancepool.DeviceModelsNameFrom(dic.Get)
	isFind = false
	if _, ok := deviceModels[deviceInstance.Model]; !ok {
		if addDeviceRequest.DeviceModels != nil {
			for i := range addDeviceRequest.DeviceModels{
				if deviceInstance.Model == addDeviceRequest.DeviceModels[i].Name {
					isFind = true
					break
				}
			}
			if !isFind {
				klog.Errorf("http callback error : deviceModel mismatch , there is no deviceModel named %s matching the uploaded file",deviceInstance.Model)
				return common.KindEntityDoesNotExist
			}
		}
	}
	// Iterate through the json file uploaded by the user to find
	isFind = false
	for k := 0; k < len(deviceInstance.PropertyVisitors); k++ {
		modelName := deviceInstance.PropertyVisitors[k].ModelName
		propertyName := deviceInstance.PropertyVisitors[k].PropertyName
		l := 0
		for l = 0; l < len(addDeviceRequest.DeviceModels); l++ {
			if modelName == addDeviceRequest.DeviceModels[l].Name {
				m := 0
				for m = 0; m < len(addDeviceRequest.DeviceModels[l].Properties); m++ {
					if propertyName == addDeviceRequest.DeviceModels[l].Properties[m].Name {
						deviceInstance.PropertyVisitors[k].PProperty = addDeviceRequest.DeviceModels[l].Properties[m]
						isFind = true
						break
					}
				}
				if m == len(addDeviceRequest.DeviceModels[l].Properties) {
					klog.Errorf("Property : %s not found in the uploaded file",propertyName)
					return common.KindServerError
				}
				break
			}
		}
		if l == len(addDeviceRequest.DeviceModels) {
			klog.Errorf("Device model : %s not found in the uploaded file",modelName)
			return common.KindServerError
		}
	}

	// Traverse and find from mapper and existing deviceModel
	if !isFind {
		for k := 0; k < len(deviceInstance.PropertyVisitors); k++ {
			modelName := deviceInstance.PropertyVisitors[k].ModelName
			propertyName := deviceInstance.PropertyVisitors[k].PropertyName
			l := 0
			for _, value := range deviceModels {
				if modelName == value.Name {
					m := 0
					for m = 0; m < len(value.Properties); m++ {
						if propertyName == value.Properties[m].Name {
							deviceInstance.PropertyVisitors[k].PProperty = value.Properties[m]
							break
						}
					}
					if m == len(value.Properties) {
						klog.Errorf("Property : %s not exists",propertyName)
						return common.KindEntityDoesNotExist
					}
					break
				}
			}
			if l == len(deviceModels) {
				klog.Errorf("Device model : %s not exists",modelName)
				return common.KindEntityDoesNotExist
			}
		}
	}
	for k := 0; k < len(deviceInstance.Twins); k++ {
		name := deviceInstance.Twins[k].PropertyName
		l := 0
		for l = 0; l < len(deviceInstance.PropertyVisitors); l++ {
			if name == deviceInstance.PropertyVisitors[l].PropertyName {
				deviceInstance.Twins[k].PVisitor = &deviceInstance.PropertyVisitors[l]
				break
			}
		}
		if l == len(deviceInstance.PropertyVisitors) {
			klog.Errorf("PropertyVisitor : %s not found",name)
			return common.KindEntityDoesNotExist
		}
	}

	mqttClient := instancepool.MqttClientNameFrom(dic.Get)
	driver := instancepool.ProtocolDriverNameFrom(dic.Get)
	wg := instancepool.WgNameFrom(dic.Get)
	stopFunctions := instancepool.StopFunctionsNameFrom(dic.Get)
	connectInfo := instancepool.ConnectInfoNameFrom(dic.Get)
	mutex := instancepool.MutexNameFrom(dic.Get)
	deviceMutex := instancepool.DeviceLockNameFrom(dic.Get)
	// Create connect info for new device instance
	for _, visitorV := range deviceInstance.PropertyVisitors {
		driverName := common.DriverPrefix + deviceInstance.ID + visitorV.PropertyName
		connectInfo[driverName] = &configmap.ConnectInfo{
			ProtocolCommonConfig: deviceInstance.PProtocol.ProtocolCommonConfig,
			VisitorConfig:        visitorV.VisitorConfig,
			ProtocolConfig:       deviceInstance.PProtocol.ProtocolConfigs,
		}
	}
	// Create device lock for device instance
	deviceMutex[instanceID] = new(common.Lock)
	deviceMutex[instanceID].DeviceLock = new(sync.Mutex)
	var waitInit *sync.WaitGroup
	waitInit = new(sync.WaitGroup)
	waitInit.Add(1)
	go func() {
		ctx, cancelFunc := context.WithCancel(context.Background())
		mqttadapter.SendTwin(ctx, instanceID, deviceInstance, driver, mqttClient, wg, dic, deviceMutex[instanceID])
		mqttadapter.SendData(ctx, instanceID, deviceInstance, driver, mqttClient, wg, dic, deviceMutex[instanceID])
		mqttadapter.SendDeviceState(ctx, instanceID, deviceInstance, driver, mqttClient, wg, dic, deviceMutex[instanceID])
		stopFunctions[instanceID] = cancelFunc
		klog.V(1).Infof("Add %s successful\n", instanceID)
		waitInit.Done()
	}()
	waitInit.Wait()
	mutex.Lock()
	if addDeviceRequest.DeviceModels != nil {
		for i := 0; i < len(addDeviceRequest.DeviceModels); i++ {
			if _, ok := deviceModels[addDeviceRequest.DeviceModels[i].Name]; !ok {
				deviceModels[addDeviceRequest.DeviceModels[i].Name] = new(configmap.DeviceModel)
				deviceModels[addDeviceRequest.DeviceModels[i].Name] = addDeviceRequest.DeviceModels[i]
			}
		}
	}
	if addDeviceRequest.Protocol != nil {
		for i := 0; i < len(addDeviceRequest.Protocol); i++ {
			if _, ok := protocols[addDeviceRequest.Protocol[i].Name]; !ok {
				protocols[addDeviceRequest.Protocol[i].Name] = new(configmap.Protocol)
				protocols[addDeviceRequest.Protocol[i].Name] = addDeviceRequest.Protocol[i]
			}
		}
	}
	// Convert the json file passed in by the user into a mapper-related structure and update it.
	// deviceInstance can be stored directly because it has already judged whether the ID is repeated or not.
	deviceInstances[deviceInstance.ID] = new(configmap.DeviceInstance)
	deviceInstances[deviceInstance.ID] = deviceInstance
	mutex.Unlock()
	return ""
}

// DeleteDevice internal callback function
func DeleteDevice(instanceID string, dic *di.Container) (kind common.ErrKind) {
	stopFunctions := instancepool.StopFunctionsNameFrom(dic.Get)
	deviceInstances := instancepool.DeviceInstancesNameFrom(dic.Get)
	deviceModels := instancepool.DeviceModelsNameFrom(dic.Get)
	protocol := instancepool.ProtocolNameFrom(dic.Get)
	mutex := instancepool.MutexNameFrom(dic.Get)
	mutex.Lock()
	defer mutex.Unlock()
	if cancelFunc, ok := stopFunctions[instanceID]; ok {
		modelNameDeleted := deviceInstances[instanceID].Model
		protocolNameDeleted := deviceInstances[instanceID].ProtocolName
		if _, ok := deviceInstances[instanceID]; ok {
			delete(deviceInstances, instanceID)
		}
		cancelFunc()
		modelFlag := true
		protocolFlag := true
		for k, v := range deviceInstances {
			if k != instanceID {
				if v.Model == modelNameDeleted {
					modelFlag = false
					break
				}
			}
		}
		if modelFlag {
			delete(deviceModels, modelNameDeleted)
		}
		for k, v := range deviceInstances {
			if k != instanceID {
				if v.ProtocolName == protocolNameDeleted {
					protocolFlag = false
					break
				}
			}
		}
		if protocolFlag {
			delete(protocol, protocolNameDeleted)
		}
		delete(stopFunctions, instanceID)
		klog.V(1).Infof("Remove %s successful\n", instanceID)
	} else {
		klog.Errorf("Remove %s failed,there is no such instanceId\n", instanceID)
		return common.KindEntityDoesNotExist
	}
	return ""
}

// ReadDeviceData internal callback function
func ReadDeviceData(deviceID string, propertyName string, dic *di.Container) (string, common.ErrKind) {
	deviceInstance := instancepool.DeviceInstancesNameFrom(dic.Get)
	if _, ok := deviceInstance[deviceID]; !ok {
		return "", common.KindEntityDoesNotExist
	}
	index := -1
	for i, twin := range deviceInstance[deviceID].Twins {
		if twin.PropertyName == propertyName {
			index = i
			break
		}
	}
	if index == -1 {
		return "", common.KindEntityDoesNotExist
	}
	mapMutex := instancepool.DeviceLockNameFrom(dic.Get)
	protocolDriver := instancepool.ProtocolDriverNameFrom(dic.Get)
	response, err := controller.GetDeviceData(deviceID, deviceInstance[deviceID].Twins[index], protocolDriver, mapMutex[deviceID], dic)
	if err != nil {
		klog.Errorf("Get %s data error:", deviceID, err.Error())
		return response, common.KindServerError
	}
	return response, ""
}

// WriteDeviceData internal callback function
func WriteDeviceData(deviceID string, values url.Values, dic *di.Container) common.ErrKind {
	deviceInstance := instancepool.DeviceInstancesNameFrom(dic.Get)
	if _, ok := deviceInstance[deviceID]; !ok {
		return common.KindInvalidID
	}
	for k, v := range values {
		for i, twin := range deviceInstance[deviceID].Twins {
			if twin.PropertyName == k {
				rollback := deviceInstance[deviceID].Twins[i].Desired.Value
				deviceInstance[deviceID].Twins[i].Desired.Value = v[0]
				mapMutex := instancepool.DeviceLockNameFrom(dic.Get)
				protocolDriver := instancepool.ProtocolDriverNameFrom(dic.Get)
				err := controller.SetVisitor(deviceID, deviceInstance[deviceID].Twins[i], protocolDriver, mapMutex[deviceID], dic)
				if err != nil {
					klog.Errorf("Set %s data error:", deviceID, err.Error())
					deviceInstance[deviceID].Twins[i].Desired.Value = rollback
					return common.KindNotAllowed
				}
				if len(deviceInstance[deviceID].Twins[i].Desired.Value) > 30{
					klog.V(4).Infof("Set %s : %s value to %s......", deviceID, twin.PropertyName[:30])
				}else{
					klog.V(4).Infof("Set %s : %s value to %s", deviceID, twin.PropertyName, deviceInstance[deviceID].Twins[i].Desired.Value)
				}
				return ""
			}
		}
		return common.KindServerError
	}
	return common.KindEntityDoesNotExist
}
