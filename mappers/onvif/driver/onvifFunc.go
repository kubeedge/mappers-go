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

package driver

import (
	"io/ioutil"
	"log"

	"github.com/use-go/onvif/media"
)

func OnvifFunc(funcName string) (ret interface{}, err error) {
	switch funcName {
	case "GetStreamUri":
		res, err := dev.CallMethod(media.GetStreamUri{})
		bs, _ := ioutil.ReadAll(res.Body)
		log.Printf("output %+v %s", res.StatusCode, bs)
	case "SystemReboot":
		res, err := dev.CallMethod(device.SystemReboot{})
		bs, _ := ioutil.ReadAll(res.Body)
		log.Printf("output %+v %s", res.StatusCode, bs)
	default:
		return nil, fmt.Errorf("Wrong Onvif function: %s", funcName)
	}
}
