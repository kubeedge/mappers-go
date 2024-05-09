package driver

import (
	"sync"

	goonvif "github.com/use-go/onvif"

	"github.com/kubeedge/mapper-framework/pkg/common"
)

// CustomizedDev is the customized device configuration and client information.
type CustomizedDev struct {
	Instance         common.DeviceInstance
	CustomizedClient *CustomizedClient
}

type CustomizedClient struct {
	deviceMutex sync.Mutex
	ProtocolConfig
	dev *goonvif.Device //Save the device controller created by the device driver
}

type ProtocolConfig struct {
	ProtocolName string `json:"protocolName"`
	ConfigData   `json:"configData"`
}

type ConfigData struct {
	URL      string `json:"url,omitempty"` // the url of onvif device
	UserName string `json:"userName"`      // the username of onvif device
	Password string `json:"password"`      // the password of device user
}

type VisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

type VisitorConfigData struct {
	DataType      string `json:"dataType"`
	Format        string `json:"format"`        // datatype of device data
	OutputDir     string `json:"outputDir"`     // the url of onvif device
	FrameCount    int    `json:"frameCount"`    // the username of onvif device
	FrameInterval int    `json:"frameInterval"` // the password of device user
	VideoNum      int    `json:"videoNum"`      // number of videos collected
}
