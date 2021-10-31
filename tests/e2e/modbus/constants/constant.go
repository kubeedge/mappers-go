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
	MakeModbusDevice           = "cd $GOPATH/src/github.com/kubeedge/mappers-go;make device modbus package"
	CheckModbusDeviceImage     = "docker images | grep modbus-simulator"
	RunModbusTCPDevice         = "docker run -d --network host modbus-simulator:v1.0-linux-amd64 tcp"
	RunModbusTCPErrorDevice    = "docker run -d --network host modbus-simulator:v1.0-linux-amd64 tcperror"
	RunModbusRTUDevice         = "docker run -d -v /dev:/dev modbus-simulator:v1.0-linux-amd64 rtu"
	RunModbusRTUErrorDevice    = "docker run -d -v /dev:/dev modbus-simulator:v1.0-linux-amd64 rtuerror"
	RunSocat                   = "docker run -d -v /dev:/dev alpine/socat:1.7.3.4-r0 -d -d pty,raw,echo=0,link=/dev/ttyS001 pty,raw,echo=0,link=/dev/ttyS002"
	GetSocat                   = "docker ps | grep alpine/socat"
	GetModbusDeviceContainerID = "docker ps | grep modbus-simulator"
	DeleteModbusDevice         = "docker rmi modbus-simulator:v1.0-linux-amd64"

	MakeModbusMapper        = "cd $GOPATH/src/github.com/kubeedge/mappers-go;make mapper modbus package"
	CheckModbusMapperImage  = "docker images | grep modbus-mapper"
	DeleteModbusMapperImage = "docker rmi modbus-mapper:v1.0-linux-amd64"

	RunTCPModbusMapper         = "docker run -d --network host -v $GOPATH/src/github.com/kubeedge/mappers-go/tests/e2e/modbus/devicesprofiles/%s.json:/opt/kubeedge/deviceProfile.json modbus-mapper:v1.0-linux-amd64"
	RunRTUModbusMapper         = "docker run -d --network host -v /dev:/dev -v $GOPATH/src/github.com/kubeedge/mappers-go/tests/e2e/modbus/devicesprofiles/%s.json:/opt/kubeedge/deviceProfile.json modbus-mapper:v1.0-linux-amd64"
	GetModbusMapperContainerID = "docker ps | grep modbus-mapper"
)
