package service

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/clients/httpclient"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/clients/mqttclient"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/config"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/controller"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/httpadapter"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/instancepool"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/models"
	"k8s.io/klog/v2"
)

var (
	ms *MapperService
)

// MapperService the structure of the variable required by the mapper
type MapperService struct {
	ProtocolName     string
	configMap       string
	deviceInstances map[string]*configmap.DeviceInstance
	deviceModels    map[string]*configmap.DeviceModel
	protocol        map[string]*configmap.Protocol
	Controller      *httpadapter.RestController
	driver          models.ProtocolDriver
	connectInfo     map[string]*configmap.ConnectInfo
	dic             *di.Container
	wg              *sync.WaitGroup
	mqttClient      mqttclient.MqttClient
	httpClient      *httpclient.HTTPClient
	mutex           *sync.Mutex
	quit            chan os.Signal
	stopFunctions   map[string]context.CancelFunc
	deviceMutex     map[string]*common.Lock
}

// InitMapperService initialize the mapperService config.
func (ms *MapperService) InitMapperService(protocolName string, c config.Config, deviceInterface interface{}) {
	if err := c.Parse(); err != nil {
		klog.Errorf("Failed to parse config.yaml file :%v", err)
		os.Exit(1)
	}
	if protocolName == "" {
		klog.Errorf("Please specify device protocol name")
		os.Exit(1)
	}
	ms.ProtocolName = protocolName
	ms.configMap = c.Configmap
	ms.deviceInstances = make(map[string]*configmap.DeviceInstance)
	ms.deviceModels = make(map[string]*configmap.DeviceModel)
	ms.protocol = make(map[string]*configmap.Protocol)
	ms.connectInfo = make(map[string]*configmap.ConnectInfo)
	ms.wg = new(sync.WaitGroup)
	ms.mutex = new(sync.Mutex)
	ms.quit = make(chan os.Signal)
	ms.stopFunctions = make(map[string]context.CancelFunc)
	ms.deviceMutex = make(map[string]*common.Lock)
	if driver, ok := deviceInterface.(models.ProtocolDriver); !ok {
		klog.Errorf("Please specify device interface")
		os.Exit(1)
	} else {
		ms.driver = driver
	}
	signal.Notify(ms.quit, os.Interrupt)
	ms.waitExit()
	ms.mqttClient = mqttclient.MqttClient{
		IP:         c.Mqtt.ServerAddress,
		ServerName: c.Mqtt.ServerName,
		User:       c.Mqtt.Username,
		Passwd:     c.Mqtt.Password,
		ClientID:   c.Mqtt.ClientID,
		Cert:       c.Mqtt.Cert,
		PrivateKey: c.Mqtt.PrivateKey,
		CaCert:     c.Mqtt.CaCert,
	}
	if err := ms.mqttClient.Connect(); err != nil {
		klog.Errorf("Failed to connect mqtt broker: %v", err)
		os.Exit(1)
	}
	err := configmap.Parse(c.Configmap, ms.deviceInstances, ms.deviceModels, ms.protocol, ms.ProtocolName)
	if err != nil {
		klog.Errorf("Failed to parse configmap file %s:%v", c.Configmap, err)
		os.Exit(1)
	}
	configmap.GetConnectInfo(ms.deviceInstances, ms.connectInfo)
	ms.initDeviceMutex()
	ms.dic = di.NewContainer(di.ServiceConstructorMap{
		instancepool.DeviceInstancesName: func(get di.Get) interface{} {
			return ms.deviceInstances
		},
		instancepool.DeviceModelsName: func(get di.Get) interface{} {
			return ms.deviceModels
		},
		instancepool.ProtocolName: func(get di.Get) interface{} {
			return ms.protocol
		},
		instancepool.ConfigMapName: func(get di.Get) interface{} {
			return ms.configMap
		},
		instancepool.ProtocolDriverName: func(get di.Get) interface{} {
			return ms.driver
		},
		instancepool.MqttClientName: func(get di.Get) interface{} {
			return ms.mqttClient
		},
		instancepool.WgName: func(get di.Get) interface{} {
			return ms.wg
		},
		instancepool.MutexName: func(get di.Get) interface{} {
			return ms.mutex
		},
		instancepool.StopFunctionsName: func(get di.Get) interface{} {
			return ms.stopFunctions
		},
		instancepool.ConnectInfoName: func(get di.Get) interface{} {
			return ms.connectInfo
		},
		instancepool.DeviceLockName: func(get di.Get) interface{} {
			return ms.deviceMutex
		},
	})
	controller.InitDeviceConfig(ms.driver, ms.dic)
	ms.httpClient = httpclient.NewHTTPClient(ms.dic)
	err = ms.httpClient.Init(c)
	if err != nil {
		klog.Errorf("Failed to start Http server:%v", err)
	}
}

// waitExit create a goroutine to monitor exit signal
func (ms *MapperService) waitExit() {
	go func() {
		<-ms.quit
		err := ms.driver.StopDevice()
		if err != nil {
			klog.Errorf("Service has stopped but failed to stop device:%v", err)
			os.Exit(1)
		}
		klog.V(1).Info("Exit mapper safely")
		os.Exit(1)
	}()
}

// initDeviceMutex init the mutex of device
func (ms *MapperService) initDeviceMutex() {
	for i := range ms.deviceInstances {
		ms.deviceMutex[i] = new(common.Lock)
		ms.deviceMutex[i].DeviceLock = new(sync.Mutex)
	}
}
