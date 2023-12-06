package driver

import (
	"github.com/blackjack/webcam"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

var Cam *webcam.Webcam

func NewClient(protocol ProtocolConfig) (*CustomizedClient, error) {
	client := &CustomizedClient{
		ProtocolConfig: protocol,
		deviceMutex:    sync.Mutex{},
		// TODO initialize the variables you added
	}
	return client, nil
}

func (c *CustomizedClient) InitDevice() error {
	// TODO: add init operation
	// you can use c.ProtocolConfig
	cam, err := inner(c.SerialPort)
	if err != nil {
		return err
	}
	Cam = cam

	return nil
}

func inner(path string) (*webcam.Webcam, error) {
	var cam *webcam.Webcam = nil
	var err error
	for {
		cam, err = webcam.Open(path)
		if err == nil {
			break
		} else {
			klog.Errorf("open the carmera failed with err: %v", err)
			time.Sleep(1 * time.Second)
		}
	}
	return cam, err
}

func (c *CustomizedClient) GetDeviceData(visitor *VisitorConfig) (interface{}, error) {
	// TODO: add the code to get device's data
	// you can use c.ProtocolConfig and visitor
	featureName := visitor.FeatureName
	switch featureName {
	case Framerate:
		return Cam.GetFramerate()
	case Input:
		return Cam.GetInput()
	case BusInfo:
		return Cam.GetBusInfo()
	case Gain, Contrast, Saturation, WhiteBalanceTemperature,
		WhiteBalanceTemperatureAuto, Sharpness, PowerLineFrequency, ExposureAuto, ExposureAbsolute, Brightness:
		controls := Cam.GetControls()
		return GetControl(controls, featureName)
	case ImageTrigger:
		return GetImage(c.Width, c.Height, c.Format)
	}
	return nil, nil
}

func (c *CustomizedClient) SetDeviceData(data interface{}, visitor *VisitorConfig) error {
	// TODO: set device's data
	// you can use c.ProtocolConfig and visitor
	featureName := visitor.FeatureName
	switch featureName {
	case Framerate:
		return Cam.SetFramerate(float32(data.(float64)))
	case BufferCount:
		return Cam.SetBufferCount(uint32(data.(uint64)))
	case Gain, Contrast, Saturation,
		WhiteBalanceTemperatureAuto, Sharpness, PowerLineFrequency, ExposureAuto, Brightness:
		controls := Cam.GetControls()
		return SetControl(controls, featureName, data)
		//case Width:
		//	_, err := GetOrSetSize(data.(int), Width)
		//	return err
		//case Height:
		//	_, err := GetOrSetSize(data.(int), Height)
		//	return err
	}
	return nil
}

func (c *CustomizedClient) StopDevice() error {
	// TODO: stop device
	// you can use c.ProtocolConfig
	err := Cam.Close()
	if err != nil {
		return err
	}
	return nil
}
