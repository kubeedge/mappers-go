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
	"context"
	"fmt"
	"net"
	"os"
	"time"

	dmiapi "github.com/kubeedge/mappers-go/pkg/apis/upstream/v1"
	"google.golang.org/grpc"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/mappers/modbus/device"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/global"
	"github.com/kubeedge/mappers-go/pkg/grpcserver"
	"github.com/kubeedge/mappers-go/pkg/httpserver"
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

	global.MqttClient = common.MqttClient{
		IP:         c.Mqtt.ServerAddress,
		User:       c.Mqtt.Username,
		Passwd:     c.Mqtt.Password,
		Cert:       c.Mqtt.Cert,
		PrivateKey: c.Mqtt.PrivateKey,
	}
	if err = global.MqttClient.Connect(); err != nil {
		klog.Fatal(err)
		os.Exit(1)
	}

	panel := device.NewDevPanel()
	if err = panel.DevInit(&c); err != nil {
		klog.Fatal(err)
	}
	klog.Infoln("devInit finished")

	// register to edgecore
	// TODO health check
	if err = registerMapper(c.Common); err != nil {
		klog.Fatal(err)
	}
	klog.Infoln("registerMapper finished")

	// start grpc server
	grpcServer := grpcserver.NewServer(
		grpcserver.Config{
			SockPath: c.GrpcServer.SocketPath,
			Protocol: common.ProtocolModbus,
		},
	)
	go grpcServer.Start()
	klog.Infoln("grpc server start finished")
	go httpserver.StartHttpServer(c.HttpServer.Host)
	klog.Infoln("http server start finished")
	panel.DevStart()
}

func registerMapper(cfg config.Common) error {
	// 连接grpc服务器
	//conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	conn, err := grpc.Dial(cfg.EdgeCoreSock,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				unixAddress, err := net.ResolveUnixAddr("unix", cfg.EdgeCoreSock)
				if err != nil {
					return nil, err
				}
				return net.DialUnix("unix", nil, unixAddress)
			},
		),
	)
	if err != nil {
		return fmt.Errorf("did not connect: %v", err)
	}
	// 延迟关闭连接
	defer conn.Close()

	// 初始化Greeter服务客户端
	c := dmiapi.NewMapperClient(conn)

	// 初始化上下文，设置请求超时时间为1秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 延迟关闭请求会话
	defer cancel()

	// 调用SayHello接口，发送一条消息
	_, err = c.MapperRegister(ctx, &dmiapi.MapperRegisterRequest{
		Mapper: &dmiapi.MapperInfo{
			Name:       cfg.Name,
			Version:    cfg.Version,
			ApiVersion: cfg.APIVersion,
			Protocol:   cfg.Protocol,
			Address:    []byte(cfg.Address),
			State:      common.DEVSTOK,
		},
	})
	return err
}
