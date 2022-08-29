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

package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/onvif"
	"github.com/kubeedge/mappers-go/pkg/global"
	"github.com/kubeedge/mappers-go/pkg/util/parse"

	"k8s.io/klog/v2"
)

var deviceMuxs map[string]context.CancelFunc
var devices map[string]*onvif.OnvifDev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitor property is readonly, if not then set it.
func setVisitor(visitorConfig *onvif.OnvifVisitorConfig, twin *common.Twin, client *onvif.OnvifClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		return
	}

	fmt.Printf("%v", visitorConfig)
	method, ok := visitorConfig.ConfigData["method"]
	if !ok {
		klog.Error("set visitor error: no method")
		return
	}
	err := client.Set(method.(string), twin.Desired.Value)
	if err != nil {
		klog.Errorf("set visitor error: %v, %v", err, visitorConfig)
		return
	}
}

// getDeviceID extract the device ID from Mqtt topic.
func getDeviceID(topic string) (id string) {
	re := regexp.MustCompile(`hw/events/device/(.+)/twin/update/delta`)
	return re.FindStringSubmatch(topic)[1]
}

// initOnvif initialize Onvif client.
func initOnvif(name string, protocolConfig onvif.OnvifProtocolConfig) (client *onvif.OnvifClient, err error) {
	OnvifConfig := onvif.OnvifConfig{Name: name}

	var ok bool
	var tmp interface{}
	if tmp, ok = protocolConfig.ConfigData["url"]; !ok {
		return nil, errors.New("Protocol config has not url")
	}
	OnvifConfig.URL = tmp.(string)
	if tmp, ok = protocolConfig.ConfigData["userName"]; !ok {
		klog.V(2).Info("protocol config has not userName")
	} else {
		OnvifConfig.User = tmp.(string)
	}
	if tmp, ok = protocolConfig.ConfigData["password"]; !ok {
		klog.V(2).Info("protocol config has not passwordfile")
	} else {
		OnvifConfig.Passwordfile = tmp.(string)
	}
	if tmp, ok = protocolConfig.ConfigData["cert"]; !ok {
		klog.V(2).Info("Protocol config has not certfile")
	} else {
		OnvifConfig.Certfile = tmp.(string)
	}
	if tmp, ok = protocolConfig.ConfigData["remoteCert"]; !ok {
		klog.V(2).Info("protocol config has not remoteCertfile")
	} else {
		OnvifConfig.RemoteCertfile = tmp.(string)
	}
	if tmp, ok = protocolConfig.ConfigData["key"]; !ok {
		klog.V(2).Info("protocol config has not keyfile")
	} else {
		OnvifConfig.Keyfile = tmp.(string)
	}
	klog.V(2).Info("onvif configuration: ", OnvifConfig)
	return onvif.NewClient(OnvifConfig)
}

// initTwin initialize the timer to get twin value.
func initTwin(ctx context.Context, dev *onvif.OnvifDev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig onvif.OnvifVisitorConfig
		if err := json.Unmarshal(dev.Instance.Twins[i].PVisitor.VisitorConfig, &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		method, ok := visitorConfig.ConfigData["method"]
		if !ok {
			klog.Error("init twin error: no method")
			continue
		}

		if method.(string) == "SaveFrame" {
			outDir, ok1 := visitorConfig.ConfigData["outputDir"]
			format, ok2 := visitorConfig.ConfigData["format"]
			frameCount, ok3 := visitorConfig.ConfigData["frameCount"]
			frameInterval, ok4 := visitorConfig.ConfigData["frameInterval"]
			if !ok1 || !ok2 {
				klog.Error("init twin error: no outputDir or format")
				continue
			}
			if !ok3 || !ok4 {
				klog.Error("init twin error: no frameCount or frameInterval")
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				streamURI := dev.OnvifClient.GetStream()
				err := onvif.SaveFrame(streamURI, outDir.(string), format.(string), int(frameCount.(float64)), int(frameInterval.(float64)))
				if err != nil {
					klog.Errorf("init twin error: %v", err)
					return
				}
			}()

			twinData := TwinData{Client: dev.OnvifClient,
				Name:   dev.Instance.Twins[i].PropertyName,
				Method: method.(string),
				Topic:  fmt.Sprintf(common.TopicTwinUpdate, dev.Instance.ID)}
			collectCycle := time.Duration(dev.Instance.Twins[i].PVisitor.CollectCycle)
			// If the collect cycle is not set, set it to 1 second.
			if collectCycle == 0 {
				collectCycle = 1 * time.Second
			}
			ticker := time.NewTicker(collectCycle)
			go func() {
				for {
					select {
					case <-ticker.C:
						twinData.Run()
					case <-ctx.Done():
						return
					}
				}
			}()
		} else if method.(string) == "SaveVideo" {
			outDir, ok1 := visitorConfig.ConfigData["outputDir"]
			format, ok2 := visitorConfig.ConfigData["format"]
			frameCount, ok3 := visitorConfig.ConfigData["frameCount"]
			if !ok1 || !ok2 || !ok3 {
				klog.Error("init twin error: no outputDir, format, or frameCount")
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				streamURI := dev.OnvifClient.GetStream()
				err := onvif.SaveVideo(streamURI, outDir.(string), format.(string), int(frameCount.(float64)))
				if err != nil {
					klog.Errorf("init twin error: %v", err)
					return
				}
			}()

			twinData := TwinData{Client: dev.OnvifClient,
				Name:   dev.Instance.Twins[i].PropertyName,
				Method: method.(string),
				Topic:  fmt.Sprintf(common.TopicTwinUpdate, dev.Instance.ID)}
			collectCycle := time.Duration(dev.Instance.Twins[i].PVisitor.CollectCycle)
			// If the collect cycle is not set, set it to 1 second.
			if collectCycle == 0 {
				collectCycle = 1 * time.Second
			}
			ticker := time.NewTicker(collectCycle)
			go func() {
				for {
					select {
					case <-ticker.C:
						twinData.Run()
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.OnvifClient)
	}
}

// initSubscribeMqtt subscribe Mqtt topics.
func initSubscribeMqtt(instanceID string) error {
	topic := fmt.Sprintf(common.TopicTwinUpdateDelta, instanceID)
	klog.V(1).Info("Subscribe topic: ", topic)
	err := global.MqttClient.Subscribe(topic, OnEventBus)
	if err != nil {
		return err
	}
	klog.V(1).Info("Subscribe topic: ", TopicOnvifGetResource)
	return global.MqttClient.Subscribe(TopicOnvifGetResource, On3rdParty)
}

// initGetStatus start timer to get device status and send to eventbus.
func initGetStatus(ctx context.Context, dev *onvif.OnvifDev) {
	getStatus := GetStatus{Client: dev.OnvifClient,
		topic: fmt.Sprintf(common.TopicStateUpdate, dev.Instance.ID)}
	ticker := time.NewTicker(time.Second)
	select {
	case <-ticker.C:
		getStatus.Run()
	case <-ctx.Done():
		return
	}
}

// start the device.
func start(ctx context.Context, dev *onvif.OnvifDev) {
	var protocolConfig onvif.OnvifProtocolConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolConfigs, &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfig error: %v", err)
		wg.Done()
		return
	}

	client, err := initOnvif(dev.Instance.Name, protocolConfig)
	if err != nil {
		klog.Errorf("Init error: %v", err)
		wg.Done()
		return
	}
	dev.OnvifClient = client
	go initTwin(ctx, dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Errorf("Init subscribe mqtt error: %v", err)
		wg.Done()
		return
	}

	go initGetStatus(ctx, dev)
	<-ctx.Done()
	wg.Done()
}

// DevInit initialize the device data.
func DevInit(cfg *config.Config) error {
	devices = make(map[string]*onvif.OnvifDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)
	deviceMuxs = make(map[string]context.CancelFunc)
	devs := make(map[string]*common.DeviceInstance)

	switch cfg.DevInit.Mode {
	case common.DevInitModeRegister:
		if err := parse.ParseByUsingRegister(cfg, devs, models, protocols); err != nil {
			return err
		}
	case common.DevInitModeConfigmap:
		if err := parse.Parse(cfg.DevInit.Configmap, devs, models, protocols); err != nil {
			return err
		}
	case common.DevInitModeMetaServer:
		if err := parse.ParseByUsingMetaServer(cfg, devs, models, protocols); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported dev init mode %s", cfg.DevInit.Mode)
	}

	for key, deviceInstance := range devs {
		cur := new(onvif.OnvifDev)
		cur.Instance = *deviceInstance
		devices[key] = cur
	}
	return nil
}

// DevStart start all devices.
func DevStart() {
	for id, dev := range devices {
		klog.V(4).Info("Dev: ", id, dev)
		ctx, cancel := context.WithCancel(context.Background())
		deviceMuxs[id] = cancel
		wg.Add(1)
		go start(ctx, dev)
	}
	wg.Wait()
}

func UpdateDev(model *common.DeviceModel, device *common.DeviceInstance, protocol *common.Protocol) {
	devices[device.ID] = new(onvif.OnvifDev)
	devices[device.ID].Instance = *device
	models[device.Model] = *model
	protocols[device.ProtocolName] = *protocol

	// stop old
	if err := stopDev(device.ID); err != nil {
		klog.Error(err)
	}
	// start new
	ctx, cancelFunc := context.WithCancel(context.Background())
	deviceMuxs[device.ID] = cancelFunc
	wg.Add(1)
	go start(ctx, devices[device.ID])
}

func stopDev(id string) error {
	cancelFunc, ok := deviceMuxs[id]
	if !ok {
		return fmt.Errorf("can not find device %s from device muxs", id)
	}
	cancelFunc()
	return nil
}

func DealDeviceTwinGet(deviceID string, twinName string) (interface{}, error) {
	srcDev, ok := devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("not found device %s", deviceID)
	}

	res := make([]parse.TwinResultResponse, 0)
	for _, twin := range srcDev.Instance.Twins {
		if twinName != "" && twin.PropertyName != twinName {
			continue
		}
		payload, err := getTwinData(deviceID, twin, srcDev.OnvifClient)
		if err != nil {
			return nil, err
		}
		cur := parse.TwinResultResponse{
			PropertyName: twin.PropertyName,
			Payload:      payload,
		}
		res = append(res, cur)
	}
	return json.Marshal(res)
}

func getTwinData(deviceID string, twin common.Twin, client *onvif.OnvifClient) ([]byte, error) {
	var visitorConfig onvif.OnvifVisitorConfig
	if err := json.Unmarshal(twin.PVisitor.VisitorConfig, &visitorConfig); err != nil {
		return nil, fmt.Errorf("unmarshal VisitorConfig error: %v", err)
	}
	setVisitor(&visitorConfig, &twin, client)
	method, ok := visitorConfig.ConfigData["method"]
	if !ok {
		return nil, fmt.Errorf("get twin data failed: no method")
	}
	td := TwinData{
		Client: client,
		Name:   twin.PropertyName,
		Method: method.(string),
		Topic:  fmt.Sprintf(common.TopicTwinUpdate, deviceID),
	}
	return td.GetPayload()
}
