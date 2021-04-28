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

// TemplateVisitorConfig is the Template register configuration.
type TemplateVisitorConfig struct {
}

// TemplateProtocolConfig is the protocol configuration.
type TemplateProtocolConfig struct {
}

// TemplateProtocolCommonConfig is the Template protocol configuration.
type TemplateProtocolCommonConfig struct {
}

// CustomizedValue is the customized part for Template protocol.
type CustomizedValue map[string]interface{}
