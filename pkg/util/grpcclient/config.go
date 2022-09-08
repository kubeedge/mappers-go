package grpcclient

import (
	"github.com/kubeedge/mappers-go/config"
)

var cfg *config.Config

func Init(c *config.Config) {
	cfg = &config.Config{}
	cfg = c
}
