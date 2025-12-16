/*
Copyright 2025 The KubeEdge Authors.

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

package main

import (
	"errors"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mapper-framework/pkg/common"
	"github.com/kubeedge/mapper-framework/pkg/config"
	"github.com/kubeedge/mapper-framework/pkg/grpcclient"
	"github.com/kubeedge/mapper-framework/pkg/grpcserver"
	"github.com/kubeedge/mapper-framework/pkg/httpserver"
	"github.com/kubeedge/modbus/device"
)

func main() {
	var err error
	var c *config.Config

	klog.InitFlags(nil)
	defer klog.Flush()

	if c, err = config.Parse(); err != nil {
		klog.Fatal(err)
	}
	klog.Infof("config: %+v", c)

	klog.Infoln("Mapper will register to edgecore")
	deviceList, deviceModelList, err := grpcclient.RegisterMapper(true)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infoln("Mapper register finished")

	panel := device.NewDevPanel()
	err = panel.DevInit(deviceList, deviceModelList)
	if err != nil && !errors.Is(err, device.ErrEmptyData) {
		klog.Fatal(err)
	}
	klog.Infoln("devInit finished")
	go panel.DevStart()

	// start http server
	httpServer := httpserver.NewRestServer(panel, c.Common.HTTPPort)
	go httpServer.StartServer()

	// start grpc server
	grpcServer := grpcserver.NewServer(
		grpcserver.Config{
			SockPath: c.GrpcServer.SocketPath,
			Protocol: common.ProtocolCustomized,
		},
		panel,
	)
	defer grpcServer.Stop()
	if err = grpcServer.Start(); err != nil {
		klog.Fatal(err)
	}

}
