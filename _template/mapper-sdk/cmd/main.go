package main

import (
	"github.com/kubeedge/mappers-go/_template/mapper-sdk/driver"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/service"
)

// main Template device program entry
func main() {
	d := &driver.Template{}
	// TODO: Modify your protocol name to be consistent with the CRDs definition
	service.Bootstrap("Template", d)
}
