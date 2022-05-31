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

package onvif

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/beevik/etree"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/media"
	"k8s.io/klog/v2"
)

func getXMLValue(source string, path string) (result string, err error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(source); err != nil {
		return "", fmt.Errorf("Parser error: %v. Source %s, path %s", err, source, path)
	}
	ret := doc.Root().FindElement(path)
	if ret == nil {
		return "", fmt.Errorf("Parser error: element : %s not found in %s", path, source)
	}
	return strings.Split(ret.Text(), " ")[0], nil
}

func getResponseXML(dev *onvif.Device, method interface{}, path string) (string, error) {
	res, err := dev.CallMethod(method)
	if err != nil {
		return "", err
	}
	bs, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return "", fmt.Errorf("get error: %v", res.StatusCode)
	}
	return getXMLValue(string(bs), path)
}

func getStreamURI(dev *onvif.Device) (string, error) {
	return getResponseXML(dev, media.GetStreamUri{}, "./Body/GetStreamUriResponse/MediaUri/Uri")
}

func systemReboot(dev *onvif.Device) {
	res, err := getResponseXML(dev, device.SystemReboot{}, "")
	klog.V(5).Info("systemReboot res", res, err)
}

func getSystemDateAndTime(dev *onvif.Device) (string, error) {
	return getResponseXML(dev, device.GetSystemDateAndTime{}, "")
}

// OnvifFunc standard ONVIF functions. Users could expand this function as required.
func OnvifFunc(dev *onvif.Device, funcName string, value string) (ret string, err error) {
	switch funcName {
	case "GetStreamUri":
		return getStreamURI(dev)
	case "GetSystemDateAndTime":
		return getSystemDateAndTime(dev)
	case "SystemReboot":
		systemReboot(dev)
		return "", nil
	default:
		return "", fmt.Errorf("Wrong Onvif function: %s", funcName)
	}
}

func readFile(file string) []byte {
	ret, err := ioutil.ReadFile(file)
	if err != nil {
		klog.Errorf("get onvif resource error. Failed to read file: %s", file)
		return []byte{}
	}
	return ret
}

// GetOnvifResources get stream resources.
func GetOnvifResources() OnvifResources {
	var r OnvifResources

	r.Resources = make(map[string]*Resource, len(clients))
	for _, client := range clients {
		index := client.Config.Name
		r.Resources[index] = new(Resource)
		r.Resources[index].UserName = client.Config.User
		r.Resources[index].Password = readFile(client.Config.Passwordfile)
		r.Resources[index].Certfile = readFile(client.Config.Certfile)
		r.Resources[index].RemoteCertfile = readFile(client.Config.RemoteCertfile)
		r.Resources[index].Keyfile = readFile(client.Config.Keyfile)
		r.Resources[index].URL = client.Config.StreamURI
	}
	return r
}
