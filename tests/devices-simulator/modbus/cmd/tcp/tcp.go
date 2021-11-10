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

package tcp

import (
	"github.com/kubeedge/mappers-go/tests/devices-simulator/modbus/cmd/tcp/config"
	"github.com/kubeedge/mappers-go/tests/devices-simulator/modbus/device"
	"github.com/spf13/cobra"
)

func CommandTCP() *cobra.Command {
	var cfg = config.NewConfig()

	var c = &cobra.Command{
		Use:  "tcp",
		Long: "modbus tcp simulator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return device.RunAsTCPNormal(cfg)
		},
	}

	cfg.Flags(c.Flags())
	return c
}
