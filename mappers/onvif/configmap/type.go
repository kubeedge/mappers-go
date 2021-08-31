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

// OnvifVisitorConfig is the Onvif register configuration.
type OnvifVisitorConfig struct {
	ProtocolName string          `json:"protocolName"`
	ConfigData   CustomizedValue `json:"configData"`
}

// OnvifProtocolConfig is the protocol configuration.
type OnvifProtocolConfig struct {
	ProtocolName string          `json:"protocolName"`
	ConfigData   CustomizedValue `json:"configData"`
}

// OnvifProtocolCommonConfig is the Onvif protocol configuration.
type OnvifProtocolCommonConfig struct {
}

// CustomizedValue is the customized part for Onvif protocol.
type CustomizedValue map[string]interface{}
