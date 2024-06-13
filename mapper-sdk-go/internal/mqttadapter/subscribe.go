package mqttadapter

import (
	"context"
	"encoding/json"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"regexp"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/controller"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/instancepool"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
)

// SyncInfo callback function of Mqtt subscribe message.
// The function will update device's value according to the message sent from the cloud
func SyncInfo(dic *di.Container, message mqtt.Message) {
	re := regexp.MustCompile(`hw/events/device/(.+)/twin/update/delta`)
	instanceID := re.FindStringSubmatch(message.Topic())[1]
	deviceInstances := instancepool.DeviceInstancesNameFrom(dic.Get)
	driver := instancepool.ProtocolDriverNameFrom(dic.Get)
	mapMutex := instancepool.DeviceLockNameFrom(dic.Get)
	if _, ok := deviceInstances[instanceID]; !ok {
		klog.Errorf("Instance :%s does not exist", instanceID)
		return
	}
	var delta DeviceTwinDelta
	if err := json.Unmarshal(message.Payload(), &delta); err != nil {
		klog.Errorf("Unmarshal %s message failed: %v", instanceID, err)
		return
	}
	for twinName, twinValue := range delta.Delta {
		i := 0
		for i = 0; i < len(deviceInstances[instanceID].Twins); i++ {
			if twinName == deviceInstances[instanceID].Twins[i].PropertyName {
				break
			}
		}
		if i == len(deviceInstances[instanceID].Twins) {
			continue
		}
		if len(twinValue) > 30{
			klog.V(4).Infof("Set %s:%s value to %s......", instanceID, twinName, twinValue[:30])
		}else{
			klog.V(4).Infof("Set %s:%s value to %s", instanceID, twinName, twinValue)
		}
		deviceInstances[instanceID].Twins[i].Desired.Value = twinValue
		err := controller.SetVisitor(instanceID, deviceInstances[instanceID].Twins[i], driver, mapMutex[instanceID], dic)
		if err != nil {
			klog.Error(err)
			return
		}
	}
}

// UpdateDevice callback function of Mqtt subscribe message.
// The function support for dynamically adding/removing devices
func UpdateDevice(dic *di.Container, message mqtt.Message) {
	choices := make(map[string]string)
	if err := json.Unmarshal(message.Payload(), &choices); err != nil {
		klog.Errorf("Unmarshal UpdateDevice message failed: %v", err)
		return
	}
	if choices["option"] == "add" {
		addDevice(dic, message)
	} else {
		removeDevice(dic, message)
	}

}

// removeDevice support for dynamically removing devices, delete only local memory data
func removeDevice(dic *di.Container, message mqtt.Message) {
	re := regexp.MustCompile(`hw/events/node/(.+)/membership/updated`)
	stopFunctions := instancepool.StopFunctionsNameFrom(dic.Get)
	deviceInstances := instancepool.DeviceInstancesNameFrom(dic.Get)
	deviceModels := instancepool.DeviceModelsNameFrom(dic.Get)
	protocol := instancepool.ProtocolNameFrom(dic.Get)
	instanceID := re.FindStringSubmatch(message.Topic())[1]
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
	}
}

// addDevice support for dynamically adding devices , delete only local memory data
func addDevice(dic *di.Container, message mqtt.Message) {
	re := regexp.MustCompile(`hw/events/node/(.+)/membership/updated`)
	instanceID := re.FindStringSubmatch(message.Topic())[1]
	configMap := instancepool.ConfigMapNameFrom(dic.Get)
	deviceInstances := instancepool.DeviceInstancesNameFrom(dic.Get)
	deviceModels := instancepool.DeviceModelsNameFrom(dic.Get)
	protocol := instancepool.ProtocolNameFrom(dic.Get)
	connectInfo := instancepool.ConnectInfoNameFrom(dic.Get)
	driver := instancepool.ProtocolDriverNameFrom(dic.Get)
	mqttClient := instancepool.MqttClientNameFrom(dic.Get)
	wg := instancepool.WgNameFrom(dic.Get)
	deviceMutex := instancepool.DeviceLockNameFrom(dic.Get)
	stopFunctions := instancepool.StopFunctionsNameFrom(dic.Get)
	defaultConfigFile := configMap
	mutex := instancepool.MutexNameFrom(dic.Get)
	mutex.Lock()
	// parseConfigmap
	if err := configmap.ParseOdd(defaultConfigFile, deviceInstances, deviceModels, protocol, instanceID); err != nil {
		klog.Errorf("Please check you config-file %s,", err.Error())
		return
	}
	mutex.Unlock()
	configmap.GetConnectInfo(deviceInstances, connectInfo)
	deviceMutex[instanceID] = new(common.Lock)
	deviceMutex[instanceID].DeviceLock = new(sync.Mutex)
	go func() {
		ctx, cancelFunc := context.WithCancel(context.Background())
		SendTwin(ctx, instanceID, deviceInstances[instanceID], driver, mqttClient, wg, dic, deviceMutex[instanceID])
		SendData(ctx, instanceID, deviceInstances[instanceID], driver, mqttClient, wg, dic, deviceMutex[instanceID])
		SendDeviceState(ctx, instanceID, deviceInstances[instanceID], driver, mqttClient, wg, dic, deviceMutex[instanceID])
		stopFunctions[instanceID] = cancelFunc
		klog.V(1).Infof("Add %s successful\n", instanceID)
	}()
}
