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

// This application needs OPCUA server and device.
// Please edit by demand for testing.

package driver

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/use-go/onvif/media"
)

func TestReadWithoutAuth(t *testing.T) {
	c := ONVIFConfig{URL: "192.168.1.64:80",
		User:         "admin",
		Passwordfile: "/home/wei/ca/pass"}

	dev, err := NewDevice(c)
	assert.Nil(t, err)
	log.Printf("output %+v", dev.GetServices())
	/*
		res, err := dev.CallMethod(device.GetUsers{})
		bs, _ := ioutil.ReadAll(res.Body)
		log.Printf("output %+v %s", res.StatusCode, bs)

		capabilities := device.GetCapabilities{Category: "All"}
		fmt.Print(capabilities)
	*/
	res, err := dev.CallMethod(media.GetStreamUri{})
	bs, _ := ioutil.ReadAll(res.Body)
	log.Printf("output %+v %s", res.StatusCode, bs)
}
