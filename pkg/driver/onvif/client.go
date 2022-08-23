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
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/kubeedge/mappers-go/pkg/common"

	"github.com/use-go/onvif"
	"k8s.io/klog/v2"
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
	StreamURI      string
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

	client.Config.StreamURI, err = OnvifFunc(client.Client, "GetStreamUri", "")
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (c *OnvifClient) GetStream() string {
	tmp := strings.Split(c.Config.StreamURI, "://")
	protocol := tmp[0]
	tmp = strings.Split(tmp[1], "/")
	ip := tmp[0]
	password, _ := readPassword(c.Config.Passwordfile)

	streamURI := fmt.Sprintf("%s://%s:%s@%s", protocol, c.Config.User, password, ip)
	fmt.Println("stream: ", streamURI)
	return streamURI
}

// GetStatus get device status.
// For the package onvif doesn't expose any http/connection function,
// we call the GetSystemDataAndTime function to get the connection status.
func (c *OnvifClient) GetStatus() string {
	_, err := OnvifFunc(c.Client, "GetSystemDateAndTime", "")
	if err == nil {
		return common.DEVSTOK
	}
	return common.DEVSTDISCONN
}

// Get get register.
func (c *OnvifClient) Get(method, value string) (results string, err error) {
	switch method {
	case "SaveFrame":
		return strconv.FormatBool(IfSaveFrame), nil
	case "SaveVideo":
		return strconv.FormatBool(IfSaveVideo), nil
	default:
		results, err = OnvifFunc(c.Client, method, value)
	}
	klog.V(2).Info("Get result: ", results)
	return results, err
}

// Set set register.
func (c *OnvifClient) Set(method, value string) (err error) {
	var tmp interface{}

	switch method {
	case "SaveFrame":
		tmp, err = common.Convert("boolean", value)
		IfSaveFrame = tmp.(bool)
	case "SaveVideo":
		tmp, err = common.Convert("boolean", value)
		IfSaveVideo = tmp.(bool)
	default:
		_, err = OnvifFunc(c.Client, method, value)
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
