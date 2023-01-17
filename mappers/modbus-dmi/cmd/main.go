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

package main

import (
	"errors"
	"os"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/mappers/modbus-dmi/device"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/grpcserver"
	"github.com/kubeedge/mappers-go/pkg/util/grpcclient"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

func main() {
	var err error
	var c config.Config

	klog.InitFlags(nil)
	defer klog.Flush()

	if err = c.Parse(); err != nil {
		klog.Fatal(err)
		os.Exit(1)
	}
	klog.Infof("config: %+v", c)

	grpcclient.Init(&c)

	// start grpc server
	grpcServer := grpcserver.NewServer(
		grpcserver.Config{
			SockPath: c.GrpcServer.SocketPath,
			Protocol: common.ProtocolModbus,
		},
	)

	panel := device.NewDevPanel()
	err = panel.DevInit(&c)
	if err != nil && !errors.Is(err, parse.ErrEmptyData) {
		klog.Fatal(err)
	}
	klog.Infoln("devInit finished")

	// register to edgecore
	// if dev init mode is register, mapper's dev will init when registry to edgecore
	if c.DevInit.Mode != common.DevInitModeRegister {
		klog.Infoln("======dev init mode is not register, will register to edgecore")
		// TODO health check
		if _, _, err = grpcclient.RegisterMapper(&c, false); err != nil {
			klog.Fatal(err)
		}
		klog.Infoln("registerMapper finished")
	}

	panel.DevStart()

	if err = grpcServer.Start(); err != nil {
		klog.Fatal(err)
	}
	klog.Infoln("grpc server start finished")
}
