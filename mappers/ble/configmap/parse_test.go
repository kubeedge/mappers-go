/*
Copyright 2021 The KubeEdge Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kubeedge/mappers-go/mappers/ble/globals"
	"github.com/kubeedge/mappers-go/mappers/common"
)

func TestParse(t *testing.T) {
	var devices map[string]*globals.BleDev
	var models map[string]common.DeviceModel
	var protocols map[string]common.Protocol

	devices = make(map[string]*globals.BleDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)

	assert.Nil(t, Parse("./configmap_test.json", devices, models, protocols))
	for _, device := range devices {
		var bpc BleProtocolConfig
		assert.Nil(t, json.Unmarshal(device.Instance.PProtocol.ProtocolConfigs, &bpc))
		assert.Equal(t, "A4:C1:38:1A:49:90", bpc.MacAddress)
	}
}
