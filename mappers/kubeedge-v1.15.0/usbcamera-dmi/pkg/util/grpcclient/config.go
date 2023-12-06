package grpcclient

import (
	"github.com/kubeedge/usb/pkg/config"
)

var cfg *config.Config

func Init(c *config.Config) {
	cfg = &config.Config{}
	cfg = c
}
