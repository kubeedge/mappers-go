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

package main

import (
	"os"

	"k8s.io/klog/v2"

	mappercommon "github.com/kubeedge/mappers-go/mappers/common"
	"github.com/kubeedge/mappers-go/mappers/opcua/config"
	"github.com/kubeedge/mappers-go/mappers/opcua/device"
	"github.com/kubeedge/mappers-go/mappers/opcua/globals"
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
	klog.V(4).Info(c.Configmap)

	globals.MqttClient = mappercommon.MqttClient{IP: c.Mqtt.ServerAddress,
		User:       c.Mqtt.UserName,
		Passwd:     c.Mqtt.Password,
		Cert:       c.Mqtt.CertFile,
		PrivateKey: c.Mqtt.PrivateKeyFile}
	if err = globals.MqttClient.Connect(); err != nil {
		klog.Fatal(err)
		os.Exit(1)
	}

	if err = device.DevInit(c.Configmap); err != nil {
		klog.Fatal(err)
		os.Exit(1)
	}
	device.DevStart()
}
