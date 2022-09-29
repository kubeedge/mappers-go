/*
Copyright 2022 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package parse

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/util/grpcclient"
)

var ErrEmptyData error = errors.New("device or device model list is empty")

// Parse the configmap.
func Parse(path string,
	devices map[string]*common.DeviceInstance,
	dms map[string]common.DeviceModel,
	protocols map[string]common.Protocol) error {
	var deviceProfile common.DeviceProfile

	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsonFile, &deviceProfile); err != nil {
		return err
	}

	for i := 0; i < len(deviceProfile.DeviceInstances); i++ {
		instance := deviceProfile.DeviceInstances[i]
		j := 0
		for j = 0; j < len(deviceProfile.Protocols); j++ {
			if instance.ProtocolName == deviceProfile.Protocols[j].Name {
				instance.PProtocol = deviceProfile.Protocols[j]
				break
			}
		}
		if j == len(deviceProfile.Protocols) {
			return errors.New("protocol not found")
		}

		if instance.PProtocol.Protocol != "bluetooth" {
			continue
		}

		for k := 0; k < len(instance.PropertyVisitors); k++ {
			modelName := instance.PropertyVisitors[k].ModelName
			propertyName := instance.PropertyVisitors[k].PropertyName
			l := 0
			for l = 0; l < len(deviceProfile.DeviceModels); l++ {
				if modelName == deviceProfile.DeviceModels[l].Name {
					m := 0
					for m = 0; m < len(deviceProfile.DeviceModels[l].Properties); m++ {
						if propertyName == deviceProfile.DeviceModels[l].Properties[m].Name {
							instance.PropertyVisitors[k].PProperty = deviceProfile.DeviceModels[l].Properties[m]
							break
						}
					}

					if m == len(deviceProfile.DeviceModels[l].Properties) {
						return errors.New("property not found")
					}
					break
				}
			}
			if l == len(deviceProfile.DeviceModels) {
				return errors.New("device model not found")
			}
		}

		for k := 0; k < len(instance.Twins); k++ {
			name := instance.Twins[k].PropertyName
			l := 0
			for l = 0; l < len(instance.PropertyVisitors); l++ {
				if name == instance.PropertyVisitors[l].PropertyName {
					instance.Twins[k].PVisitor = &instance.PropertyVisitors[l]
					break
				}
			}
			if l == len(instance.PropertyVisitors) {
				return errors.New("propertyVisitor not found")
			}
		}

		for k := 0; k < len(instance.Datas.Properties); k++ {
			name := instance.Datas.Properties[k].PropertyName
			l := 0
			for l = 0; l < len(instance.PropertyVisitors); l++ {
				if name == instance.PropertyVisitors[l].PropertyName {
					instance.Datas.Properties[k].PVisitor = &instance.PropertyVisitors[l]
					break
				}
			}
			if l == len(instance.PropertyVisitors) {
				return errors.New("propertyVisitor not found")
			}
		}

		devices[instance.ID] = new(common.DeviceInstance)
		devices[instance.ID] = &instance
		klog.V(4).Info("Instance: ", instance.ID, instance)
	}

	for i := 0; i < len(deviceProfile.DeviceModels); i++ {
		dms[deviceProfile.DeviceModels[i].Name] = deviceProfile.DeviceModels[i]
	}

	for i := 0; i < len(deviceProfile.Protocols); i++ {
		protocols[deviceProfile.Protocols[i].Name] = deviceProfile.Protocols[i]
	}
	return nil
}

func ParseByUsingRegister(cfg *config.Config,
	devices map[string]*common.DeviceInstance,
	dms map[string]common.DeviceModel,
	protocols map[string]common.Protocol) error {
	deviceList, deviceModelList, err := grpcclient.RegisterMapper(cfg, true)
	if err != nil {
		return err
	}

	if len(deviceList) == 0 || len(deviceModelList) == 0 {
		return ErrEmptyData
	}
	modelMap := make(map[string]common.DeviceModel)
	for _, model := range deviceModelList {
		cur := ParseDeviceModelFromGrpc(model)
		modelMap[model.Name] = cur
	}

	for _, device := range deviceList {
		commonModel := modelMap[device.Spec.DeviceModelReference]
		protocol, err := BuildProtocolFromGrpc(device)
		if err != nil {
			return err
		}
		instance, err := ParseDeviceFromGrpc(device, &commonModel)
		if err != nil {
			return err
		}
		instance.PProtocol = protocol
		devices[instance.ID] = new(common.DeviceInstance)
		devices[instance.ID] = instance
		klog.V(4).Info("Instance: ", instance.ID)
		dms[instance.Model] = modelMap[instance.Model]
		protocols[instance.ProtocolName] = protocol
	}

	return nil
}
