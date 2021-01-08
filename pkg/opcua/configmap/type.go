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

// Common visitor configurations for opc-ua protocol
type VisitorConfigOPCUA struct {
	// Required: The ID of opc-ua node, e.g. "ns=1,i=1005"
	NodeID string `json:"nodeID,omitempty"`
	// The name of opc-ua node
	BrowseName string `json:"browseName,omitempty"`
}

type ProtocolConfigOPCUA struct {
	// Required: The URL for opc server endpoint.
	URL string `json:"url,omitempty"`
	// Username for access opc server.
	// +optional
	UserName string `json:"userName,omitempty"`
	// Password file for access opc server.
	// +optional
	Password string `json:"password,omitempty"`
	// Defaults to "None". The value could be "None", "Basic128Rsa15", "Basic256",
	// "Basic256Sha256", "Aes128Sha256RsaOaep", "Aes256Sha256RsaPss".
	// +optional
	SecurityPolicy string `json:"securityPolicy,omitempty"`
	// Defaults to "None". The value could be "None", "Sign" and "SignAndEncrypt".
	// +optional
	SecurityMode string `json:"securityMode,omitempty"`
	// Certificate file for access opc server.
	// +optional
	Certificate string `json:"certificate,omitempty"`
	// PrivateKey file for access opc server.
	// +optional
	PrivateKey string `json:"privateKey,omitempty"`
	// Timeout seconds for the opc server connection.???
	// +optional
	Timeout int64 `json:"timeout,omitempty"`
}

// ProtocolCommonConfigOPCUA is the OPCUA configuration.
type ProtocolCommonConfigOPCUA struct {
	CustomizedValues CustomizedValue `json:"customizedValues,omitempty"`
}

// CustomizedValue is the customized part for modbus protocol.
type CustomizedValue map[string]interface{}
