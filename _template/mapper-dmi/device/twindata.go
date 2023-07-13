package device

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	dmiapi "github.com/kubeedge/kubeedge/pkg/apis/dmi/v1alpha1"
	"github.com/kubeedge/mappers-go/mappers/Template/driver"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/util/grpcclient"
	"github.com/kubeedge/mappers-go/pkg/util/parse"
)

type TwinData struct {
	DeviceName    string
	Client        *driver.CustomizedClient
	Name          string
	Type          string
	VisitorConfig *driver.TemplateVisitorConfig
	Results       interface{}
	Topic         string
}

func (td *TwinData) GetPayLoad() ([]byte, error) {
	var err error
	td.Results, err = td.Client.GetDeviceData(td.VisitorConfig)
	if err != nil {
		return nil, fmt.Errorf("get device data failed: %v", err)
	}
	sData, err := common.ConvertToString(td.Results)
	if err != nil {
		klog.Errorf("Failed to convert %s %s value as string : %v", td.DeviceName, td.Name, err)
		return nil, err
	}
	if len(sData) > 30 {
		klog.V(4).Infof("Get %s : %s ,value is %s......", td.DeviceName, td.Name, sData[:30])
	} else {
		klog.V(4).Infof("Get %s : %s ,value is %s", td.DeviceName, td.Name, sData)
	}
	var payload []byte
	if strings.Contains(td.Topic, "$hw") {
		if payload, err = common.CreateMessageTwinUpdate(td.Name, td.Type, sData); err != nil {
			return nil, fmt.Errorf("create message twin update failed: %v", err)
		}
	} else {
		if payload, err = common.CreateMessageData(td.Name, td.Type, sData); err != nil {
			return nil, fmt.Errorf("create message data failed: %v", err)
		}
	}
	return payload, nil
}

func (td *TwinData) Run() {
	payload, err := td.GetPayLoad()
	if err != nil {
		klog.Errorf("twindata %s unmarshal failed, err: %s", td.Name, err)
		return
	}

	var msg common.DeviceTwinUpdate
	if err = json.Unmarshal(payload, &msg); err != nil {
		klog.Errorf("twindata %s unmarshal failed, err: %s", td.Name, err)
		return
	}

	twins := parse.ConvMsgTwinToGrpc(msg.Twin)

	var rdsr = &dmiapi.ReportDeviceStatusRequest{
		DeviceName: td.DeviceName,
		ReportedDevice: &dmiapi.DeviceStatus{
			Twins: twins,
			State: "OK",
		},
	}

	if err := grpcclient.ReportDeviceStatus(rdsr); err != nil {
		klog.Errorf("fail to report device status of %s with err: %+v", rdsr.DeviceName, err)
	}
}
