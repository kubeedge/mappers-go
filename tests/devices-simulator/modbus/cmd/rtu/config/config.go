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

package config

import (
	"github.com/spf13/pflag"
	"strings"
)

type Config struct {
	SlaveID      uint8
	ClientAddr   string
	ServerAddr   string
	BaudRate     int
	DataBits     int
	StopBits     int
	Interval     int
	Parity       string
	RS485Enabled bool
}

func (cfg *Config) Flags(fs *pflag.FlagSet) {
	fs.Uint8VarP(&cfg.SlaveID, "id", "", cfg.SlaveID, "")
	fs.StringVarP(&cfg.ClientAddr, "client-address", "", cfg.ClientAddr, "")
	fs.StringVarP(&cfg.ServerAddr, "server-address", "", cfg.ServerAddr, "")
	fs.StringVarP(&cfg.Parity, "parity", "p", cfg.Parity, "")
	fs.IntVarP(&cfg.BaudRate, "baud-rate", "b", cfg.BaudRate, "")
	fs.IntVarP(&cfg.DataBits, "data-bits", "d", cfg.DataBits, "")
	fs.IntVarP(&cfg.StopBits, "stop-bits", "s", cfg.StopBits, "")
	fs.IntVarP(&cfg.Interval, "interval", "i", cfg.Interval, "")
	fs.BoolVar(&cfg.RS485Enabled, "rs485enable", cfg.RS485Enabled, "")
}

func (cfg *Config) Normalize() *Config {
	var parity = strings.ToUpper(cfg.Parity)
	if parity == "N" || parity == "NONE" {
		cfg.StopBits = 2
	}
	return cfg
}

func NewConfig() *Config {
	return &Config{
		SlaveID:      1,
		ClientAddr:   "/dev/ttyS001",
		ServerAddr:   "/dev/ttyS002",
		Parity:       "E",
		BaudRate:     19200,
		DataBits:     8,
		StopBits:     1,
		Interval:     60,
		RS485Enabled: false,
	}
}
