package main

import (
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/service"
	"github.com/kubeedge/mappers-go/mappers/idmvs/driver"
)

// main IDMVS device program entry
func main() {
	gd := &driver.IDMVS{}
	service.Bootstrap("IDMVS", gd)
}
