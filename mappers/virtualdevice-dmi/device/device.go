package device

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/config"
	"github.com/kubeedge/mappers-go/mappers/virtualdevice-dmi/driver"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

type DevPanel struct {
	deviceMuxs   map[string]context.CancelFunc
	devices      map[string]*driver.CustomizedDev
	models       map[string]common.DeviceModel
	protocols    map[string]common.Protocol
	wg           sync.WaitGroup
	serviceMutex sync.Mutex
	quitChan     chan os.Signal
}

var (
	devPanel *DevPanel
	once     sync.Once
)

// NewDevPanel init and return devPanel
func NewDevPanel() *DevPanel {
	once.Do(func() {
		devPanel = &DevPanel{
			deviceMuxs:   make(map[string]context.CancelFunc),
			devices:      make(map[string]*driver.CustomizedDev),
			models:       make(map[string]common.DeviceModel),
			protocols:    make(map[string]common.Protocol),
			wg:           sync.WaitGroup{},
			serviceMutex: sync.Mutex{},
			quitChan:     make(chan os.Signal),
		}
	})
	return devPanel
}

// DevStart start all devices.
func (d *DevPanel) DevStart() {
	for id, dev := range d.devices {
		klog.V(4).Info("Dev: ", id, dev)
		ctx, cancel := context.WithCancel(context.Background())
		d.deviceMuxs[id] = cancel
		d.wg.Add(1)
		go d.start(ctx, dev)
	}
	signal.Notify(d.quitChan, os.Interrupt)
	go func() {
		<-d.quitChan
		for id, device := range d.devices {
			err := device.CustomizedClient.StopDevice()
			if err != nil {
				klog.Errorf("Service has stopped but failed to stop %s:%v", id, err)
			}
		}
		klog.V(1).Info("Exit mapper")
		os.Exit(1)
	}()
	d.wg.Wait()
}

// start the device
func (d *DevPanel) start(ctx context.Context, dev *driver.CustomizedDev) {
	defer d.wg.Done()

	var protocolConfig driver.CustomizedDeviceProtocolConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolConfigs, &protocolConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolConfigs error: %v", err)
		return
	}
	var protocolCommonConfig driver.CustomizedDeviceProtocolCommonConfig
	if err := json.Unmarshal(dev.Instance.PProtocol.ProtocolCommonConfig, &protocolCommonConfig); err != nil {
		klog.Errorf("Unmarshal ProtocolCommonConfig error: %v", err)
		return
	}

	client, err := driver.NewClient(protocolCommonConfig, protocolConfig)
	if err != nil {
		klog.Errorf("Init dev %s error: %v", dev.Instance.Name, err)
		return
	}
	dev.CustomizedClient = client
	err = dev.CustomizedClient.InitDevice()
	if err != nil {
		klog.Errorf("Init device %s error: %v", dev.Instance.ID, err)
		return
	}
	go initTwin(ctx, dev)
	<-ctx.Done()
}

// initTwin initialize the timer to get twin value.
func initTwin(ctx context.Context, dev *driver.CustomizedDev) {
	for _, twin := range dev.Instance.Twins {
		var visitorConfig driver.CustomizedDeviceVisitorConfig
		err := json.Unmarshal(twin.PVisitor.VisitorConfig, &visitorConfig)
		if err != nil {
			klog.Errorf("Unmarshal VisitorConfig error: %v", err)
			continue
		}
		err = setVisitor(&visitorConfig, &twin, dev)
		if err != nil {
			klog.Error(err)
			continue
		}
		twinData := &TwinData{
			DeviceName:    dev.Instance.Name,
			Client:        dev.CustomizedClient,
			Name:          twin.PropertyName,
			Type:          twin.Desired.Metadatas.Type,
			VisitorConfig: &visitorConfig,
			Topic:         fmt.Sprintf(common.TopicTwinUpdate, dev.Instance.ID),
		}

		collectCycle := time.Duration(twin.PVisitor.CollectCycle)
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

// setVisitor check if visitor property is readonly, if not then set it.
func setVisitor(visitorConfig *driver.CustomizedDeviceVisitorConfig, twin *common.Twin, dev *driver.CustomizedDev) error {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(1).Infof("%s twin readonly property: %s", dev.Instance.Name, twin.PropertyName)
		return nil
	}
	klog.V(2).Infof("Convert type: %s, value: %s ", twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	value, err := common.Convert(twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	if err != nil {
		klog.Errorf("Failed to convert value as %s : %v", twin.PVisitor.PProperty.DataType, err)
		return err
	}
	err = dev.CustomizedClient.SetDeviceData(value, visitorConfig)
	if err != nil {
		return fmt.Errorf("%s set device data error: %v", twin.PropertyName, err)
	}
	return nil
}

// DevInit initialize the device
func (d *DevPanel) DevInit(cfg *config.Config) error {
	devs := make(map[string]*common.DeviceInstance)

	switch cfg.DevInit.Mode {
	case common.DevInitModeConfigmap:
		if err := parse.Parse(cfg.DevInit.Configmap, devs, d.models, d.protocols); err != nil {
			return err
		}
	case common.DevInitModeRegister:
		if err := parse.ParseByUsingRegister(cfg, devs, d.models, d.protocols); err != nil {
			return err
		}
	}

	for key, deviceInstance := range devs {
		cur := new(driver.CustomizedDev)
		cur.Instance = *deviceInstance
		d.devices[key] = cur
	}
	return nil
}

// UpdateDev stop old device, then update and start new device
func (d *DevPanel) UpdateDev(model *common.DeviceModel, device *common.DeviceInstance, protocol *common.Protocol) {
	d.serviceMutex.Lock()
	defer d.serviceMutex.Unlock()

	if oldDevice, ok := d.devices[device.ID]; ok {
		err := d.stopDev(oldDevice, device.ID)
		if err != nil {
			klog.Error(err)
		}
	}
	// start new device
	d.devices[device.ID] = new(driver.CustomizedDev)
	d.devices[device.ID].Instance = *device
	d.models[device.ID] = *model
	d.protocols[device.ID] = *protocol

	ctx, cancelFunc := context.WithCancel(context.Background())
	d.deviceMuxs[device.ID] = cancelFunc
	d.wg.Add(1)
	go d.start(ctx, d.devices[device.ID])
}

// UpdateDevTwins update device's twins
func (d *DevPanel) UpdateDevTwins(deviceID string, twins []common.Twin) error {
	d.serviceMutex.Lock()
	defer d.serviceMutex.Unlock()
	dev, ok := d.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}
	dev.Instance.Twins = twins
	model := d.models[dev.Instance.Model]
	protocol := d.protocols[dev.Instance.ProtocolName]
	d.UpdateDev(&model, &dev.Instance, &protocol)
	return nil
}

// DealDeviceTwinGet get device's twin data
func (d *DevPanel) DealDeviceTwinGet(deviceID string, twinName string) (interface{}, error) {
	d.serviceMutex.Lock()
	defer d.serviceMutex.Unlock()
	dev, ok := d.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("not found device %s", deviceID)
	}
	var res []parse.TwinResultResponse
	for _, twin := range dev.Instance.Twins {
		if twinName != "" && twin.PropertyName != twinName {
			continue
		}
		payload, err := getTwinData(deviceID, twin, d.devices[deviceID])
		if err != nil {
			return nil, err
		}
		item := parse.TwinResultResponse{
			PropertyName: twinName,
			Payload:      payload,
		}
		res = append(res, item)
	}
	return json.Marshal(res)
}

// getTwinData get twin
func getTwinData(deviceID string, twin common.Twin, dev *driver.CustomizedDev) ([]byte, error) {
	var visitorConfig driver.CustomizedDeviceVisitorConfig
	err := json.Unmarshal(twin.PVisitor.VisitorConfig, &visitorConfig)
	if err != nil {
		return nil, err
	}
	err = setVisitor(&visitorConfig, &twin, dev)
	if err != nil {
		return nil, err
	}
	twinData := &TwinData{
		DeviceName:    deviceID,
		Client:        dev.CustomizedClient,
		Name:          twin.PropertyName,
		Type:          twin.Desired.Metadatas.Type,
		VisitorConfig: &visitorConfig,
		Topic:         fmt.Sprintf(common.TopicTwinUpdate, deviceID),
	}
	return twinData.GetPayLoad()
}

// GetDevice get device instance
func (d *DevPanel) GetDevice(deviceID string) (interface{}, error) {
	d.serviceMutex.Lock()
	defer d.serviceMutex.Unlock()
	found, ok := d.devices[deviceID]
	if !ok || found == nil {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	// get the latest reported twin value
	for i, twin := range found.Instance.Twins {
		payload, err := getTwinData(deviceID, twin, found)
		if err != nil {
			return nil, err
		}
		found.Instance.Twins[i].Reported.Value = string(payload)
	}
	return found, nil
}

// RemoveDevice remove device instance
func (d *DevPanel) RemoveDevice(deviceID string) error {
	d.serviceMutex.Lock()
	defer d.serviceMutex.Unlock()
	dev := d.devices[deviceID]
	delete(d.devices, deviceID)
	err := d.stopDev(dev, deviceID)
	if err != nil {
		return err
	}
	return nil
}

// stopDev stop device and goroutine
func (d *DevPanel) stopDev(dev *driver.CustomizedDev, id string) error {
	cancelFunc, ok := d.deviceMuxs[id]
	if !ok {
		return fmt.Errorf("can not find device %s from device muxs", id)
	}

	err := dev.CustomizedClient.StopDevice()
	if err != nil {
		klog.Errorf("stop device %s error: %v", id, err)
	}
	cancelFunc()
	return nil
}

// GetModel if the model exists, return device model
func (d *DevPanel) GetModel(modelName string) (common.DeviceModel, error) {
	d.serviceMutex.Lock()
	defer d.serviceMutex.Unlock()
	if model, ok := d.models[modelName]; ok {
		return model, nil
	}
	return common.DeviceModel{}, fmt.Errorf("deviceModel %s not found", modelName)
}

// UpdateModel update device model
func (d *DevPanel) UpdateModel(model *common.DeviceModel) {
	d.serviceMutex.Lock()
	d.models[model.Name] = *model
	d.serviceMutex.Unlock()
}

// RemoveModel remove device model
func (d *DevPanel) RemoveModel(modelName string) {
	d.serviceMutex.Lock()
	delete(d.models, modelName)
	d.serviceMutex.Unlock()
}
