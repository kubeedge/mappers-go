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
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/opcua"
	"github.com/kubeedge/mappers-go/pkg/global"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

var deviceMuxs map[string]context.CancelFunc
var devices map[string]*opcua.OPCUADev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitory property is readonly, if not then set it.
func setVisitor(visitorConfig *opcua.VisitorConfigOPCUA, twin *common.Twin, client *opcua.OPCUAClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(1).Info("Visit readonly register: ", visitorConfig.NodeID)
		return
	}

	results, err := client.Set(visitorConfig.NodeID, twin.Desired.Value)
	if err != nil || results != "OK" {
		klog.Errorf("Set error: %v, %v", err, visitorConfig)
		return
	}
}

// onMessage callback function of Mqtt subscribe message.
func onMessage(client mqtt.Client, message mqtt.Message) {
	klog.V(2).Info("Receive message", message.Topic())
	// Get device ID and get device instance
	id := common.GetDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Device id: ", id)

	var dev *opcua.OPCUADev
	var ok bool
	if dev, ok = devices[id]; !ok {
		klog.Error("Device not exist")
		return
	}

	// Get twin map key as the propertyName
	var delta common.DeviceTwinDelta
	if err := json.Unmarshal(message.Payload(), &delta); err != nil {
		klog.Errorf("Unmarshal message failed: %v", err)
		return
	}
	for twinName, twinValue := range delta.Delta {
		i := 0
		for i = 0; i < len(dev.Instance.Twins); i++ {
			if twinName == dev.Instance.Twins[i].PropertyName {
				break
			}
		}
		if i == len(dev.Instance.Twins) {
			klog.Error("Twin not found: ", twinName)
			continue
		}
		// Desired value is not changed.
		if dev.Instance.Twins[i].Desired.Value == twinValue {
			continue
		}
		dev.Instance.Twins[i].Desired.Value = twinValue
		var visitorConfig opcua.VisitorConfigOPCUA
		if err := json.Unmarshal(dev.Instance.Twins[i].PVisitor.VisitorConfig, &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig failed: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.OPCUAClient)
	}
}

// getRemoteCertfile get configuration of remote certification file.
func getRemoteCertfile(customizedValue opcua.CustomizedValue) string {
	if len(customizedValue) != 0 {
		if value, ok := customizedValue["remoteCertificate"]; ok {
			klog.V(4).Info("getRemoteCertfile: ", value)
			return fmt.Sprintf("%v", value)
		}
	}
	klog.Error("getRemoteCertfile null")
	return ""
}

// initOPCUA initialize OPCUA client
func initOPCUA(protocolConfig opcua.ProtocolConfigOPCUA, protocolCommConfig opcua.ProtocolCommonConfigOPCUA) (client *opcua.OPCUAClient, err error) {
	if protocolConfig.SecurityPolicy == "" {
		protocolConfig.SecurityPolicy = "None"
	}
	if protocolConfig.SecurityMode == "" {
		protocolConfig.SecurityMode = "None"
	}
	config := opcua.OPCUAConfig{
		URL:            protocolConfig.URL,
		User:           protocolConfig.UserName,
		Passwordfile:   protocolConfig.Password,
		SecurityPolicy: protocolConfig.SecurityPolicy,
		SecurityMode:   protocolConfig.SecurityMode,
		Certfile:       protocolConfig.Certificate,
		RemoteCertfile: getRemoteCertfile(protocolCommConfig.CustomizedValues),
		Keyfile:        protocolConfig.PrivateKey,
	}

	return opcua.NewClient(config)
}

// initTwin initialize the timer to get twin value.
func initTwin(ctx context.Context, dev *opcua.OPCUADev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig opcua.VisitorConfigOPCUA
		if err := json.Unmarshal(dev.Instance.Twins[i].PVisitor.VisitorConfig, &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.OPCUAClient)

		twinData := TwinData{Client: dev.OPCUAClient,
			Name:   dev.Instance.Twins[i].PropertyName,
			Type:   dev.Instance.Twins[i].Desired.Metadatas.Type,
			NodeID: visitorConfig.NodeID,
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
}

// initData initialize the timer to get data.
func initData(ctx context.Context, dev *opcua.OPCUADev) {
	for i := 0; i < len(dev.Instance.Datas.Properties); i++ {
		var visitorConfig opcua.VisitorConfigOPCUA
		if err := json.Unmarshal(dev.Instance.Twins[i].PVisitor.VisitorConfig, &visitorConfig); err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}

		twinData := TwinData{Client: dev.OPCUAClient,
			Name:   dev.Instance.Datas.Properties[i].PropertyName,
			Type:   dev.Instance.Datas.Properties[i].Metadatas.Type,
			NodeID: visitorConfig.NodeID,
			Topic:  fmt.Sprintf(common.TopicDataUpdate, dev.Instance.ID)}
		collectCycle := time.Duration(dev.Instance.Datas.Properties[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		ticker := time.NewTicker(collectCycle)
		select {
		case <-ticker.C:
			twinData.Run()
		case <-ctx.Done():
			return
		}
	}
}

// initSubscribeMqtt subscribe Mqtt topics.
func initSubscribeMqtt(instanceID string) error {
	topic := fmt.Sprintf(common.TopicTwinUpdateDelta, instanceID)
	klog.V(1).Info("Subscribe topic: ", topic)
	return global.MqttClient.Subscribe(topic, onMessage)
}

// initGetStatus start timer to get device status and send to eventbus.
func initGetStatus(ctx context.Context, dev *opcua.OPCUADev) {
	getStatus := GetStatus{Client: dev.OPCUAClient,
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
func start(ctx context.Context, dev *opcua.OPCUADev) {
	var protocolConfig opcua.ProtocolConfigOPCUA
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolConfigs, &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfig error: %v", err)
		wg.Done()
		return
	}

	var protocolCommConfig opcua.ProtocolCommonConfigOPCUA
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolCommonConfig, &protocolCommConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolCommonConfig error: %v", err)
		wg.Done()
		return
	}

	klog.V(4).Info("Protocolconfig: ", protocolConfig, protocolCommConfig)
	client, err := initOPCUA(protocolConfig, protocolCommConfig)
	if err != nil {
		klog.Errorf("Init error: %v", err)
		wg.Done()
		return
	}
	dev.OPCUAClient = client

	go initTwin(ctx, dev)
	go initData(ctx, dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Errorf("Init subscribe mqtt error: %v", err)
		wg.Done()
		return
	}

	go initGetStatus(ctx, dev)
	select {
	case <-ctx.Done():
		wg.Done()
	}
}

// DevInit initialize the device data.
func DevInit(cfg *config.Config) error {
	devices = make(map[string]*opcua.OPCUADev)
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
		cur := new(opcua.OPCUADev)
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
	devices[device.ID] = new(opcua.OPCUADev)
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
		payload, err := getTwinData(deviceID, twin, srcDev.OPCUAClient)
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

func getTwinData(deviceID string, twin common.Twin, client *opcua.OPCUAClient) ([]byte, error) {
	var visitorConfig opcua.VisitorConfigOPCUA
	if err := json.Unmarshal(twin.PVisitor.VisitorConfig, &visitorConfig); err != nil {
		return nil, fmt.Errorf("unmarshal VisitorConfig error: %v", err)
	}
	setVisitor(&visitorConfig, &twin, client)

	td := TwinData{
		Client: client,
		Name:   twin.PropertyName,
		Type:   twin.Desired.Metadatas.Type,
		NodeID: visitorConfig.NodeID,
		Topic:  fmt.Sprintf(common.TopicTwinUpdate, deviceID),
	}
	return td.GetPayload()
}
