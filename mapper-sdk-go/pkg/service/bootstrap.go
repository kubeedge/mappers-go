// Package service responsible for interacting with developers
package service

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/config"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/mqttadapter"
	"k8s.io/klog/v2"
	"os"
)

// Bootstrap the entrance to mapper
func Bootstrap(protocolName string, deviceInterface interface{}) {
	var err error
	var c config.Config
	klog.InitFlags(nil)
	defer klog.Flush()
	ms = &MapperService{}
	ms.InitMapperService(protocolName, c, deviceInterface)
	klog.V(1).Info("MapperService Init Successful......")

	for id, instance := range ms.deviceInstances {
		ms.wg.Add(1)
		go publishMqtt(id, instance)
	}
	err = initSubscribeMqtt()
	if err != nil {
		klog.Errorf("Failed to subscribe mqtt topic : %v\n", err)
		os.Exit(1)
	}
	ms.wg.Wait()
	klog.V(1).Info("All devices have been deleted.Mapper exit")
}

// publishMqtt push device messages to mqtt by timer
func publishMqtt(id string, instance *configmap.DeviceInstance) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	mqttadapter.SendTwin(ctx, id, instance, ms.driver, ms.mqttClient, ms.wg, ms.dic, ms.deviceMutex[id])
	mqttadapter.SendData(ctx, id, instance, ms.driver, ms.mqttClient, ms.wg, ms.dic, ms.deviceMutex[id])
	mqttadapter.SendDeviceState(ctx, id, instance, ms.driver, ms.mqttClient, ms.wg, ms.dic, ms.deviceMutex[id])
	ms.stopFunctions[id] = cancelFunc
	ms.wg.Done()
}

// initSubscribeMqtt subscribe topics and set callback methods
func initSubscribeMqtt() error {
	for k := range ms.deviceInstances {
		topic := fmt.Sprintf(common.TopicTwinUpdateDelta, k)
		onMessage := func(client mqtt.Client, message mqtt.Message) {
			mqttadapter.SyncInfo(ms.dic, message)
		}
		err := ms.mqttClient.Subscribe(topic, onMessage)
		if err != nil {
			return err
		}
		updateDevice := func(client mqtt.Client, message mqtt.Message) {
			mqttadapter.UpdateDevice(ms.dic, message)
		}
		err = ms.mqttClient.Subscribe(common.TopicDeviceUpdate, updateDevice)
		if err != nil {
			return err
		}
		klog.V(1).Infof("Event %s is Listening....\n", k)
	}
	return nil
}
