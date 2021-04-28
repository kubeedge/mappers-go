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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	config := Config{}
	if err := config.Parse(); err != nil {
		t.Log(err)
		t.FailNow()
	}

	assert.Equal(t, "tcp://127.0.0.1:1883", config.Mqtt.ServerAddress)
	assert.Equal(t, "/opt/kubeedge/deviceProfile.json", config.Configmap)
}
