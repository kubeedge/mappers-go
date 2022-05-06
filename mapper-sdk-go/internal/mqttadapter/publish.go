package mqttadapter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/clients/mqttclient"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/common"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/configmap"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/controller"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/models"
)

// SendTwin send twin to EdgeCore according to timer
func SendTwin(ctx context.Context, id string, instance *configmap.DeviceInstance, drivers models.ProtocolDriver, mqttClient mqttclient.MqttClient, wg *sync.WaitGroup, dic *di.Container, mutex *common.Lock) {
	for _, twinV := range instance.Twins {
		// ---------------setVisitor---------------
		err := controller.SetVisitor(id, twinV, drivers, mutex, dic)
		if err != nil {
			err = errors.New("Set device config error:" + err.Error())
		}
		// ---------------setVisitor---------------
		// ---------------Send Data by MQTT---------------
		collectCycle := time.Duration(twinV.PVisitor.CollectCycle)
		wg.Add(1)
		if collectCycle == -1 {
			go func() {
				defer wg.Done()
				<-ctx.Done()
				return
			}()
		} else {
			// If the collect cycle is not set, set it to 1 second.
			if collectCycle == 0 {
				collectCycle = 1 * time.Second
			}
			twinData := TwinData{
				Name:       twinV.PropertyName,
				Type:       twinV.Desired.Metadatas.Type,
				Topic:      fmt.Sprintf(common.TopicTwinUpdate, id),
				MqttClient: mqttClient,
				driverUnit: DriverUnit{
					instanceID: id,
					twin:       twinV,
					drivers:    drivers,
					mutex:      mutex,
					dic:        dic,
				},
			}
			timer := common.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
			go func() {
				timer.Start()
			}()
			go func() {
				defer wg.Done()
				<-ctx.Done()
				timer.Stop()
				return
			}()
		}

		// ---------------Send Data by MQTT---------------
	}
}

// SendData send twin to third-part application according to timer
func SendData(ctx context.Context, id string, instance *configmap.DeviceInstance, drivers models.ProtocolDriver, mqttClient mqttclient.MqttClient, wg *sync.WaitGroup, dic *di.Container, mutex *common.Lock) {
	for _, twinV := range instance.Twins {
		// ---------------Send Data by MQTT---------------
		collectCycle := time.Duration(twinV.PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == -1 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ctx.Done()
				return
			}()
		} else {
			if collectCycle == 0 {
				collectCycle = 1 * time.Second
			}
			twinData := TwinData{
				Name:       twinV.PropertyName,
				Type:       twinV.Desired.Metadatas.Type,
				Topic:      fmt.Sprintf(common.TopicDataUpdate, id),
				MqttClient: mqttClient,
				driverUnit: DriverUnit{
					instanceID: id,
					twin:       twinV,
					drivers:    drivers,
					mutex:      mutex,
					dic:        dic,
				},
			}
			timer := common.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
			wg.Add(1)
			go func() {

				timer.Start()
			}()
			go func() {
				defer wg.Done()
				<-ctx.Done()
				timer.Stop()
				return
			}()
		}
		// ---------------Send Data by MQTT---------------
	}
}

// SendDeviceState send device's state to EdgeCore according to timer
func SendDeviceState(ctx context.Context, id string, instance *configmap.DeviceInstance, drivers models.ProtocolDriver, mqttClient mqttclient.MqttClient, wg *sync.WaitGroup, dic *di.Container, mutex *common.Lock) {
	var statusData StatusData
	var collectCycle time.Duration
	for _, twinV := range instance.Twins {
		// ---------------Send Data by MQTT---------------
		collectCycle = time.Duration(twinV.PVisitor.CollectCycle)
		if collectCycle == -1 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ctx.Done()
				return
			}()
		} else {
			// If the collect cycle is not set, set it to 1 second.
			if collectCycle == 0 {
				collectCycle = 1 * time.Second
			}
			statusData = StatusData{
				topic:      fmt.Sprintf(common.TopicStateUpdate, id),
				MqttClient: mqttClient,
				driverUnit: DriverUnit{
					instanceID: id,
					twin:       twinV,
					drivers:    drivers,
					mutex:      mutex,
					dic:        dic,
				},
			}
			timer := common.Timer{Function: statusData.Run, Duration: collectCycle, Times: 0}
			wg.Add(1)
			go func() {
				timer.Start()
			}()
			go func() {
				defer wg.Done()
				<-ctx.Done()
				timer.Stop()
				return
			}()
		}
	}
}
