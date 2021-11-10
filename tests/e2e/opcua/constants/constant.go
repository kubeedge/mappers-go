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

package constants

const (
	MakeOpcuaDevice           = "cd $GOPATH/src/github.com/kubeedge/mappers-go/tests/devices-simulator/opcua;docker build -t opcua-simulator:v1.0-linux-amd64 ."
	CheckOpcuaDeviceImage     = "docker images | grep opcua-simulator"
	RunOpcuaDevice            = "docker run -d --network host opcua-simulator:v1.0-linux-amd64"
	GetOpcuaDeviceContainerID = "docker ps | grep opcua-simulator"
	DeleteOpcuaDevice         = "docker rmi opcua-simulator:v1.0-linux-amd64"

	MakeOpcuaMapper           = "cd $GOPATH/src/github.com/kubeedge/mappers-go;make mapper opcua package"
	CheckOpcuaMapperImage     = "docker images | grep opcua-mapper"
	DeleteOpcuaMapperImage    = "docker rmi opcua-mapper:v1.0-linux-amd64"
	RunOpcuaMapper            = "docker run -d --network host -v $GOPATH/src/github.com/kubeedge/mappers-go/tests/e2e/opcua/devicesprofiles/%s.json:/opt/kubeedge/deviceProfile.json -v /etc/kubeedge/ca:/ca opcua-mapper:v1.0-linux-amd64"
	GetOpcuaMapperContainerID = "docker ps | grep opcua-mapper"

	SourceCaPath = "$GOPATH/src/github.com/kubeedge/mappers-go/tests/e2e/opcua/ca"
	DstCaPath    = "/etc/kubeedge/ca/testca"
)