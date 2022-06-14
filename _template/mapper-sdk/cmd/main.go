package main

import (
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/service"
	"github.com/kubeedge/mappers-go/mappers/Template/driver"
)

// main Template device program entry
func main() {
	d := &driver.Template{}
	// TODO: Modify your protocol name to be consistent with the CRDs definition
	service.Bootstrap("Template", d)
}
