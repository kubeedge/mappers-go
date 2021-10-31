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
)

type Config struct {
	SlaveID  uint8
	Port     int
	Interval int
}

func (cfg *Config) Flags(fs *pflag.FlagSet) {
	fs.Uint8VarP(&cfg.SlaveID, "id", "", cfg.SlaveID, "")
	fs.IntVarP(&cfg.Port, "port", "p", cfg.Port, "")
	fs.IntVarP(&cfg.Interval, "interval", "i", cfg.Interval, "")
}

func NewConfig() *Config {
	return &Config{
		SlaveID:  1,
		Port:     5020,
		Interval: 60,
	}
}
