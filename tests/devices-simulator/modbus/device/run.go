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

package device

import (
	"fmt"
	"time"

	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
	"github.com/pkg/errors"
	"github.com/tbrandon/mbserver"

	rtu "github.com/kubeedge/mappers-go/tests/devices-simulator/modbus/cmd/rtu/config"
	tcp "github.com/kubeedge/mappers-go/tests/devices-simulator/modbus/cmd/tcp/config"
	log "k8s.io/klog"
)

func RunAsRTU(s *mbserver.Server, cfg *rtu.Config) error {
	fmt.Printf(cfg.ServerAddr)
	if err := s.ListenRTU(&serial.Config{
		Address:  cfg.ServerAddr,
		BaudRate: cfg.BaudRate,
		DataBits: cfg.DataBits,
		StopBits: cfg.StopBits,
		Parity:   cfg.Parity,
		RS485: serial.RS485Config{
			Enabled: cfg.RS485Enabled,
		},
	}); err != nil {
		return err
	}
	defer s.Close()
	log.Info("Listening on " + cfg.ServerAddr)

	var handler = modbus.NewRTUClientHandler(cfg.ClientAddr)
	handler.BaudRate = cfg.BaudRate
	handler.DataBits = cfg.DataBits
	handler.StopBits = cfg.StopBits
	handler.Parity = cfg.Parity
	handler.SlaveId = cfg.SlaveID
	handler.RS485.Enabled = cfg.RS485Enabled
	_ = handler.Connect()

	t := newThermometer(handler)
	defer t.close()
	return t.setValue(s, time.Duration(cfg.Interval)*time.Second)
}

func RunAsTCP(s *mbserver.Server, cfg *tcp.Config) error {
	var sAddress = "0.0.0.0:" + fmt.Sprintf("%d", cfg.Port)
	if err := s.ListenTCP(sAddress); err != nil {
		return err
	}
	defer s.Close()
	log.Info("Listening on " + sAddress)

	var handler = modbus.NewTCPClientHandler(sAddress)
	handler.SlaveId = cfg.SlaveID
	_ = handler.Connect()

	t := newThermometer(handler)
	defer t.close()

	return t.setValue(s, time.Duration(cfg.Interval)*time.Second)
}

func RunAsRTUNormal(cfg *rtu.Config) error {
	var s = mbserver.NewServer()
	err := RunAsRTU(s, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to start Modbus RTU server")
	}
	return err
}

func RunAsTCPNormal(cfg *tcp.Config) error {
	var s = mbserver.NewServer()
	err := RunAsTCP(s, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to start Modbus TCP server")
	}
	return err
}

func RunAsRTUError(cfg *rtu.Config) error {
	var s = mbserver.NewServer()
	InsertServerError(s)
	err := RunAsRTU(s, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to start Modbus RTU Error server")
	}
	return err
}

func RunAsTCPError(cfg *tcp.Config) error {
	var s = mbserver.NewServer()
	InsertServerError(s)
	err := RunAsTCP(s, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to start Modbus TCP Error server")
	}
	return err
}
