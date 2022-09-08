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

	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/kubeedge/kubeedge/cloud/pkg/apis/devices/v1alpha2"
	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/util/grpcclient"
	"github.com/kubeedge/mappers-go/pkg/util/httpclient"
	"k8s.io/klog/v2"
)

var tmpDevices = []v1alpha2.Device{
	{
		TypeMeta: v1.TypeMeta{
			Kind:       "Device",
			APIVersion: "devices.kubeedge.io/v1alpha2",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "modbustcp-device",
			Namespace: "dec-public",
			Labels: map[string]string{
				"description": "counter",
				"model":       "simulation",
			},
		},
		Spec: v1alpha2.DeviceSpec{
			DeviceModelRef: &v12.LocalObjectReference{
				Name: "modbus-sample-model",
			},
			Protocol: v1alpha2.ProtocolConfig{
				Modbus: &v1alpha2.ProtocolConfigModbus{SlaveID: pointer.Int64Ptr(1)},
				Common: &v1alpha2.ProtocolConfigCommon{
					TCP: &v1alpha2.ProtocolConfigTCP{
						IP:   "10.222.1.1",
						Port: 502,
					},
				},
			},
			PropertyVisitors: []v1alpha2.DevicePropertyVisitor{
				{
					PropertyName: "property0",
					ReportCycle:  5000,
					CollectCycle: 5000,
					VisitorConfig: v1alpha2.VisitorConfig{
						Modbus: &v1alpha2.VisitorConfigModbus{
							Register:       "HoldingRegister",
							Offset:         pointer.Int64Ptr(0),
							Limit:          pointer.Int64Ptr(1),
							Scale:          1,
							IsSwap:         true,
							IsRegisterSwap: true,
						},
					},
				},
				{
					PropertyName: "property1",
					ReportCycle:  5000,
					CollectCycle: 5000,
					VisitorConfig: v1alpha2.VisitorConfig{
						Modbus: &v1alpha2.VisitorConfigModbus{
							Register:       "HoldingRegister",
							Offset:         pointer.Int64Ptr(1),
							Limit:          pointer.Int64Ptr(1),
							Scale:          1,
							IsSwap:         true,
							IsRegisterSwap: true,
						},
					},
				},
				{
					PropertyName: "property2",
					ReportCycle:  5000,
					CollectCycle: 5000,
					VisitorConfig: v1alpha2.VisitorConfig{
						Modbus: &v1alpha2.VisitorConfigModbus{
							Register:       "HoldingRegister",
							Offset:         pointer.Int64Ptr(2),
							Limit:          pointer.Int64Ptr(1),
							Scale:          1,
							IsSwap:         true,
							IsRegisterSwap: true,
						},
					},
				},
			},
			Data: v1alpha2.DeviceData{},
			NodeSelector: &v12.NodeSelector{
				NodeSelectorTerms: []v12.NodeSelectorTerm{
					{
						MatchExpressions: []v12.NodeSelectorRequirement{
							{
								Key:      "",
								Operator: v12.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
						},
					},
				},
			},
		},
		Status: v1alpha2.DeviceStatus{
			Twins: []v1alpha2.Twin{
				{
					PropertyName: "property0",
					Desired: v1alpha2.TwinProperty{
						Value: "0",
						Metadata: map[string]string{
							"timestamp": "1550049403598",
							"type":      "int",
						},
					},
					Reported: v1alpha2.TwinProperty{},
				},
				{
					PropertyName: "property1",
					Desired: v1alpha2.TwinProperty{
						Value: "0",
						Metadata: map[string]string{
							"timestamp": "1550049403598",
							"type":      "int",
						},
					},
					Reported: v1alpha2.TwinProperty{},
				},
				{
					PropertyName: "property2",
					Desired: v1alpha2.TwinProperty{
						Value: "0",
						Metadata: map[string]string{
							"timestamp": "1550049403598",
							"type":      "int",
						},
					},
					Reported: v1alpha2.TwinProperty{},
				},
			},
		},
	},
}

var tmpDeviceModels = []v1alpha2.DeviceModel{
	{
		TypeMeta: v1.TypeMeta{
			Kind:       "DeviceModel",
			APIVersion: "devices.kubeedge.io/v1alpha2",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "modbus-sample-model",
			Namespace: "dec-public",
		},
		Spec: v1alpha2.DeviceModelSpec{
			Properties: []v1alpha2.DeviceProperty{
				{
					Name:        "property0",
					Description: "property0 description",
					Type: v1alpha2.PropertyType{
						Int: &v1alpha2.PropertyTypeInt64{
							AccessMode:   "ReadWrite",
							DefaultValue: 0,
						},
					},
				},
				{
					Name:        "property1",
					Description: "property1 description",
					Type: v1alpha2.PropertyType{
						Int: &v1alpha2.PropertyTypeInt64{
							AccessMode:   "ReadWrite",
							DefaultValue: 0,
						},
					},
				},
				{
					Name:        "property2",
					Description: "property2 description",
					Type: v1alpha2.PropertyType{
						Int: &v1alpha2.PropertyTypeInt64{
							AccessMode:   "ReadWrite",
							DefaultValue: 0,
						},
					},
				},
			},
		},
	},
}

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

func ParseByUsingMetaServer(cfg *config.Config,
	devices map[string]*common.DeviceInstance,
	dms map[string]common.DeviceModel,
	protocols map[string]common.Protocol) error {
	// TODO it may be get all device from namespace
	deviceList, err := httpclient.GetDeviceList(cfg.DevInit.MetaServer.Addr, cfg.DevInit.MetaServer.Namespace)
	if err != nil {
		return err
	}
	if len(deviceList) == 0 {
		return ErrEmptyData
	}
	deviceModelList, err := httpclient.GetDeviceModelList(cfg.DevInit.MetaServer.Addr, cfg.DevInit.MetaServer.Namespace)
	if err != nil {
		return err
	}
	if len(deviceModelList) == 0 {
		return ErrEmptyData
	}
	// TODO test only
	//deviceList := tmpDevices
	//deviceModelList := tmpDeviceModels
	modelMap := make(map[string]common.DeviceModel)
	for _, model := range deviceModelList {
		cur := ParseDeviceModel(&model)
		modelMap[model.Name] = cur
	}

	for _, device := range deviceList {
		commonModel := modelMap[device.Spec.DeviceModelRef.Name]
		protocol, err := BuildProtocol(&device)
		if err != nil {
			return err
		}
		instance, err := ParseDevice(&device, &commonModel)
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

func ParseByUsingRegister(cfg *config.Config,
	devices map[string]*common.DeviceInstance,
	dms map[string]common.DeviceModel,
	protocols map[string]common.Protocol) error {
	klog.Infoln("======parse device and model from register")
	deviceList, deviceModelList, err := grpcclient.RegisterMapper(cfg, true)
	if err != nil {
		return err
	}

	if len(deviceList) == 0 || len(deviceModelList) == 0 {
		return ErrEmptyData
	}
	modelMap := make(map[string]common.DeviceModel)
	for _, model := range deviceModelList {
		klog.Infof("======ParseByUsingRegister model %+v", model)
		cur := ParseDeviceModelFromGrpc(model)
		klog.Infof("======model map key: %s, value: %+v", model.Name, cur)
		modelMap[model.Name] = cur
	}

	for _, device := range deviceList {
		klog.Infof("======ParseByUsingRegister device %+v", device)
		commonModel := modelMap[device.Spec.DeviceModelReference]
		klog.Infof("======get model from map, ref key: %s, value: %+v", device.Spec.DeviceModelReference, commonModel)
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
