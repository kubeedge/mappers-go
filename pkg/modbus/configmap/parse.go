/*
Copyright 2020 The KubeEdge Authors.

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

package configmap

import (
	"encoding/json"
	"io/ioutil"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/modbus/globals"
)

// Parse parse the configmap.
func Parse(path string,
	devices map[string]*globals.ModbusDev,
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
		err := common.ValidateProfileDeviceInstance(&instance, &deviceProfile, "modbus")
		if err == common.ErrorProtocolNotExpected {
			continue
		} else if err != nil {
			klog.Errorf("validate device profile failed: %v", err)
			return err
		}

		devices[instance.ID] = new(globals.ModbusDev)
		devices[instance.ID].Instance = instance
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
