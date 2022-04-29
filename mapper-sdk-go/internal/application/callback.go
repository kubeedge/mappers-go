package application

import (
	"context"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/controller"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/httpadapter/requests"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/instancepool"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/mqttadapter"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
	"k8s.io/klog/v2"
	"net/url"
	"sync"
)

// AddDevice internal callback function
func AddDevice(addDeviceRequest requests.AddDeviceRequest, dic *di.Container) (kind common.ErrKind) {
	instanceID := addDeviceRequest.DeviceInstance.ID
	deviceInstance := addDeviceRequest.DeviceInstance
	deviceInstances := instancepool.DeviceInstancesNameFrom(dic.Get)
	protocols := instancepool.ProtocolNameFrom(dic.Get)
	// Check if the device ID added by the user is duplicated
	for k := range deviceInstances {
		if k == instanceID {
			klog.Error("The deviceInstance name used in the uploaded file already exists in mapperService")
			return common.KindDuplicateName
		}
	}
	// Check whether mapperService already has a protocol
	if _, ok := protocols[deviceInstance.ProtocolName]; !ok {
		if addDeviceRequest.Protocol != nil {
			i := 0
			for i = 0; i < len(addDeviceRequest.Protocol); i++ {
				if deviceInstance.ProtocolName == addDeviceRequest.Protocol[i].Name {
					break
				}
			}
			if i == len(addDeviceRequest.Protocol) {
				klog.Error("http callback error : protocol mismatch , there is no protocol matching the uploaded file")
				return common.KindEntityDoesNotExist
			}
		}
	}
	// Check whether mapperService already has a deviceModel
	deviceModels := instancepool.DeviceModelsNameFrom(dic.Get)
	if _, ok := deviceModels[deviceInstance.Model]; !ok {
		if addDeviceRequest.DeviceModels != nil {
			i := 0
			for i = 0; i < len(addDeviceRequest.DeviceModels); i++ {
				if deviceInstance.Model == addDeviceRequest.DeviceModels[i].Name {
					break
				}
			}
			if i == len(addDeviceRequest.DeviceModels) {
				klog.Error("http callback error : deviceModel mismatch , there is no deviceModel matching the uploaded file")
				return common.KindEntityDoesNotExist
			}
		}
	}
	// Iterate through the json file uploaded by the user to find
	isFind := false
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
					klog.Error("Property not found in the uploaded file")
					return common.KindServerError
				}
				break
			}
		}
		if l == len(addDeviceRequest.DeviceModels) {
			klog.Error("Device model not found in the uploaded file")
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
						klog.Error("Property not exists")
						return common.KindEntityDoesNotExist
					}
					break
				}
			}
			if l == len(deviceModels) {
				klog.Error("Device model not exists")
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
			klog.Error("PropertyVisitor not found")
			return common.KindEntityDoesNotExist
		}
	}

	mqttClient := instancepool.MqttClientNameFrom(dic.Get)
	driver := instancepool.ProtocolDriverNameFrom(dic.Get)
	wg := instancepool.WgNameFrom(dic.Get)
	stopFunctions := instancepool.StopFunctionsNameFrom(dic.Get)
	connectInfo := instancepool.ConnectInfoNameFrom(dic.Get)
	mutex := instancepool.MutexNameFrom(dic.Get)
	mapMutex := instancepool.DeviceLockNameFrom(dic.Get)
	for _, visitorV := range deviceInstance.PropertyVisitors {
		tempVisitorV := visitorV
		driverName := common.DriverPrefix + deviceInstance.ID + visitorV.PropertyName
		connectInfo[driverName] = &configmap.ConnectInfo{
			ProtocolCommonConfig: deviceInstance.PProtocol.ProtocolCommonConfig,
			VisitorConfig:        tempVisitorV.VisitorConfig,
			ProtocolConfig:       deviceInstance.PProtocol.ProtocolConfigs,
		}
	}
	var waitInit *sync.WaitGroup
	waitInit = new(sync.WaitGroup)
	waitInit.Add(1)
	go func() {
		ctx, cancelFunc := context.WithCancel(context.Background())
		mqttadapter.SendTwin(ctx, instanceID, deviceInstance, driver, mqttClient, wg, dic, mapMutex[instanceID])
		mqttadapter.SendData(ctx, instanceID, deviceInstance, driver, mqttClient, wg, dic, mapMutex[instanceID])
		mqttadapter.SendDeviceState(ctx, instanceID, deviceInstance, driver, mqttClient, wg, dic, mapMutex[instanceID])
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
	return kind
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
		cancelFunc()
		modelNameDeleted := deviceInstances[instanceID].Model
		protocolNameDeleted := deviceInstances[instanceID].ProtocolName
		if _, ok := deviceInstances[instanceID]; ok {
			delete(deviceInstances, instanceID)
		}
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
		klog.V(1).Infof("Remove %s failed,there is no such instanceId\n", instanceID)
		return common.KindEntityDoesNotExist
	}
	return ""
}

// ReadDeviceData internal callback function
func ReadDeviceData(deviceID string, propertyName string, dic *di.Container) (response string, kind common.ErrKind) {
	deviceInstance := instancepool.DeviceInstancesNameFrom(dic.Get)
	if _, ok := deviceInstance[deviceID]; !ok {
		kind := common.KindEntityDoesNotExist
		return "", kind
	}
	index := -1
	for i, twin := range deviceInstance[deviceID].Twins {
		if twin.PropertyName == propertyName {
			index = i
			break
		}
	}
	if index == -1 {
		kind := common.KindEntityDoesNotExist
		return "", kind
	}
	mapMutex := instancepool.DeviceLockNameFrom(dic.Get)
	protocolDriver := instancepool.ProtocolDriverNameFrom(dic.Get)
	response, err := controller.GetDeviceData(deviceID, deviceInstance[deviceID].Twins[index], protocolDriver, mapMutex[deviceID], dic)
	if err != nil {
		kind = common.KindServerError
		return response, kind
	}
	return response, ""
}

// WriteDeviceData internal callback function
func WriteDeviceData(deviceID string, values url.Values, dic *di.Container) (kind common.ErrKind) {
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
