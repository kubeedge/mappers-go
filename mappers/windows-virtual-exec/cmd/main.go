/*
Copyright 2024 The KubeEdge Authors.
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
	"os/signal"

	klog "k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/config"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/core/model"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/core/mqtt"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/core/store"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/missions"
)

func main() {
	var err error
	var c config.Config

	klog.InitFlags(nil)
	defer klog.Flush()

	if err = c.Parse(); err != nil {
		klog.Fatal(err)
	}

	store.InitDB("internal.db")
	if err := store.DB.AutoMigrate(&model.Mission{}); err != nil {
		klog.Errorf("Failed to init db: %v", err)
	}

	if err := mqtt.InitClient(
		c.Mqtt.ServerAddress,
		c.Mqtt.Username,
		c.Mqtt.Password,
		c.Mqtt.Cert,
		c.Mqtt.PrivateKey,
	); err != nil {
		klog.Fatal(err)
	}

	missions.InitCallback(c.NodeName)
	klog.Info("Start to subscribe")
	missions.InitMissions(c.NodeName)

	// waiting kill signal
	var ch = make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	klog.Info("Exit")
}
