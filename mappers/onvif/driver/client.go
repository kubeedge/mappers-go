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
	"errors"
	"io/ioutil"

	"github.com/kubeedge/mappers-go/mappers/common"
	"k8s.io/klog/v2"

	"github.com/use-go/onvif"
)

// OnvifConfig is the structure for client configuration.
type OnvifConfig struct {
	Name           string
	URL            string
	User           string
	Passwordfile   string
	Certfile       string
	RemoteCertfile string
	Keyfile        string
}

// OnvifClient is the structure for Onvif client.
type OnvifClient struct {
	Client  *onvif.Device
	Handler interface{}
	Config  OnvifConfig
}

var clients map[string]*OnvifClient

func newDevice(config OnvifConfig) (dev *onvif.Device, err error) {
	var password string
	
	dev, err = onvif.NewDevice(config.URL)
	if err != nil {
		return nil, err
	}
	password, err = readPassword(config.Passwordfile)
	if err != nil {
		return nil, err
	}
	dev.Authenticate(config.User, password)
	return dev, nil
}

// NewClient allocate and return a Onvif client.
func NewClient(config OnvifConfig) (*OnvifClient, error) {
	if client, ok := clients[config.URL]; ok {
		return client, nil
	}

	if clients == nil {
		clients = make(map[string]*OnvifClient)
	}
	dev, err := newDevice(config)
	if err != nil {
		return nil, err
	}
	client := OnvifClient{Client: dev, Config: config}
	clients[config.URL] = &client

	return &client, nil
}

// GetStatus get device status.
func (c *OnvifClient) GetStatus() string {
	return common.DEVSTOK
}

// Get get register.
func (c *OnvifClient) Get() (results []byte, err error) {
	klog.V(2).Info("Get result: ", results)
	return results, err
}

// Set set register.
func (c *OnvifClient) Set(method string, value string) (err error) {
	err = nil

	switch method{
	case "SaveFrame":
		IfSaveFrame = common.Convert("boolean", value)
	case "SaveVideo":
		IfSaveVideo = common.Convert("boolean", value)
	default:
		_, err = OnvifFunc(method, value)
	}
	klog.V(1).Info("Set result:", err)
	return err
}

func readPassword(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.New("Failed to load certificate")
	}
	// Remove the last character '\n'
	return string(b[:len(b)-1]), nil
}

func GetOnvifResources() OnvifResources {
	var r OnvifResources

	r.Resources = make(map[string]Resource)
	for _, client := range clients {
		index := client.Config.Name
		r.Resources[index].URL, err = OnvifFunc("GetStreamUri")
		if err != nil {
			klog.Errorf("Call Onvif function error: %v", err)
		}
		r.Resource[index].UserName = client.Config.UserName
		r.Resource[index].Password = ioutil.ReadFile(client.Config.Passwordfile)
		r.Resource[index].Certfile = ioutil.ReadFile(client.Config.Certfile)
		r.Resource[index].RemoteCertfile = ioutil.ReadFile(client.Config.RemoteCertfile)
		r.Resource[index].Keyfile = ioutil.ReadFile(client.Config.Keyfile)
	}
	return r
}
