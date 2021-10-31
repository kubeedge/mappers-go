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
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

	"github.com/kubeedge/mappers-go/tests/devices-simulator/modbus/cmd/rtu"
	"github.com/kubeedge/mappers-go/tests/devices-simulator/modbus/cmd/tcp"
)

var allCommands = []*cobra.Command{
	tcp.CommandTCP(),
	rtu.CommandRTU(),
	tcp.CommandTCPError(),
	rtu.CommandRTUEroor(),
}

func main() {
	var c = &cobra.Command{
		Use: "modbus",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				basename  = filepath.Base(os.Args[1])
				targetCmd *cobra.Command
			)
			for _, cmd := range allCommands {
				if cmd.Name() == basename {
					targetCmd = cmd
					break
				}
			}
			if targetCmd != nil {
				return targetCmd.Execute()
			}
			return cmd.Help()
		},
	}
	c.AddCommand(allCommands...)

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
