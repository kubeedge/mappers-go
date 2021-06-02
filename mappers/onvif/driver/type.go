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

package driver

type OnvifResources struct {
	Resources map[string]*Resource `json:"resources"`
}

type Resource struct {
	URL            string `json:"url"`
	UserName       string `json:"userName,omitempty"`
	Password       string `json:"password,omitempty"`
	Certfile       string `json:"certfile,omitempty"`
	RemoteCertfile string `json:"remoteCertfile,omitempty"`
	Keyfile        string `json:"keyfile,omitempty"`
}
