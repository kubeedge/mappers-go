package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kubeedge/kubeedge/cloud/pkg/devicecontroller/constants"
	"github.com/kubeedge/kubeedge/pkg/apis/devices/v1alpha2"
	dmiapi "github.com/kubeedge/kubeedge/pkg/apis/dmi/v1alpha1"
	"github.com/kubeedge/mappers-go/pkg/common"
)

func getProtocolName(device *v1alpha2.Device) (string, error) {
	if device.Spec.Protocol.Modbus != nil {
		return constants.Modbus, nil
	}
	if device.Spec.Protocol.OpcUA != nil {
		return constants.OPCUA, nil
	}
	if device.Spec.Protocol.Bluetooth != nil {
		return constants.Bluetooth, nil
	}
	if device.Spec.Protocol.CustomizedProtocol != nil {
		return constants.CustomizedProtocol, nil
	}
	return "", errors.New("can not parse device protocol")
}

func BuildProtocol(device *v1alpha2.Device) (common.Protocol, error) {
	protocolName, err := getProtocolName(device)
	if err != nil {
		return common.Protocol{}, err
	}
	protocolCommonConfig, err := json.Marshal(device.Spec.Protocol.Common)
	if err != nil {
		return common.Protocol{}, err
	}
	var protocolConfig []byte
	switch protocolName {
	case constants.Modbus:
		protocolConfig, err = json.Marshal(device.Spec.Protocol.Modbus)
		if err != nil {
			return common.Protocol{}, err
		}
	case constants.OPCUA:
		protocolConfig, err = json.Marshal(device.Spec.Protocol.OpcUA)
		if err != nil {
			return common.Protocol{}, err
		}
	case constants.Bluetooth:
		protocolConfig, err = json.Marshal(device.Spec.Protocol.Bluetooth)
		if err != nil {
			return common.Protocol{}, err
		}
	case constants.CustomizedProtocol:
		protocolConfig, err = json.Marshal(device.Spec.Protocol.CustomizedProtocol)
		if err != nil {
			return common.Protocol{}, err
		}
	}
	return common.Protocol{
		Name:                 protocolName + "-" + device.Name,
		Protocol:             protocolName,
		ProtocolConfigs:      protocolConfig,
		ProtocolCommonConfig: protocolCommonConfig,
	}, nil
}

func buildTwins(device *v1alpha2.Device) []common.Twin {
	if len(device.Status.Twins) == 0 {
		return nil
	}
	res := make([]common.Twin, 0, len(device.Status.Twins))
	for _, twin := range device.Status.Twins {
		cur := common.Twin{
			PropertyName: twin.PropertyName,
			Desired: common.DesiredData{
				Value: twin.Desired.Value,
				Metadatas: common.Metadata{
					Timestamp: twin.Desired.Metadata["timestamp"],
					Type:      twin.Desired.Metadata["type"],
				},
			},
			Reported: common.ReportedData{
				Value: twin.Reported.Value,
				Metadatas: common.Metadata{
					Timestamp: twin.Desired.Metadata["timestamp"],
					Type:      twin.Desired.Metadata["type"],
				},
			},
		}
		res = append(res, cur)
	}
	return res
}

func buildData(device *v1alpha2.Device) common.Data {
	res := common.Data{}
	if len(device.Spec.Data.DataProperties) > 0 {
		res.Properties = make([]common.DataProperty, 0, len(device.Spec.Data.DataProperties))
		for _, property := range device.Spec.Data.DataProperties {
			timestamp, ok := property.Metadata["timestamp"]
			var t int64
			if ok {
				t, _ = strconv.ParseInt(timestamp, 10, 64)
			}
			cur := common.DataProperty{
				Metadatas: common.DataMetadata{
					Timestamp: t,
					Type:      property.Metadata["type"],
				},
				PropertyName: property.PropertyName,
				PVisitor:     nil,
			}
			res.Properties = append(res.Properties, cur)
		}
	}
	if strings.TrimSpace(device.Spec.Data.DataTopic) != "" {
		res.Topic = device.Spec.Data.DataTopic
	}
	return res
}

func buildPropertyVisitors(device *v1alpha2.Device) []common.PropertyVisitor {
	if len(device.Spec.PropertyVisitors) == 0 {
		return nil
	}
	protocolName, err := getProtocolName(device)
	if err != nil {
		return nil
	}
	res := make([]common.PropertyVisitor, 0, len(device.Spec.PropertyVisitors))
	for _, pptv := range device.Spec.PropertyVisitors {
		var visitorConfig []byte
		switch protocolName {
		case constants.Modbus:
			visitorConfig, err = json.Marshal(pptv.Modbus)
			if err != nil {
				return nil
			}
		case constants.OPCUA:
			visitorConfig, err = json.Marshal(pptv.OpcUA)
			if err != nil {
				return nil
			}
		case constants.Bluetooth:
			visitorConfig, err = json.Marshal(pptv.Bluetooth)
			if err != nil {
				return nil
			}
		case constants.CustomizedProtocol:
			visitorConfig, err = json.Marshal(pptv.CustomizedProtocol)
			if err != nil {
				return nil
			}
		}
		cur := common.PropertyVisitor{
			Name:          pptv.PropertyName,
			PropertyName:  pptv.PropertyName,
			ModelName:     device.Spec.DeviceModelRef.Name,
			CollectCycle:  pptv.CollectCycle,
			ReportCycle:   pptv.ReportCycle,
			Protocol:      protocolName,
			VisitorConfig: visitorConfig,
		}
		res = append(res, cur)
	}
	return res
}

func ParseDeviceModel(model *v1alpha2.DeviceModel) common.DeviceModel {
	cur := common.DeviceModel{
		Name: model.Name,
	}
	if len(model.Spec.Properties) == 0 {
		return cur
	}
	properties := make([]common.Property, 0, len(model.Spec.Properties))
	for _, property := range model.Spec.Properties {
		p := common.Property{
			Name:        property.Name,
			Description: property.Description,
		}
		if property.Type.String != nil {
			p.DataType = "string"
			p.AccessMode = string(property.Type.String.AccessMode)
			p.DefaultValue = property.Type.String.DefaultValue
		} else if property.Type.Bytes != nil {
			p.DataType = "bytes"
			p.AccessMode = string(property.Type.Bytes.AccessMode)
		} else if property.Type.Boolean != nil {
			p.DataType = "boolean"
			p.AccessMode = string(property.Type.Boolean.AccessMode)
			p.DefaultValue = property.Type.Boolean.DefaultValue
		} else if property.Type.Int != nil {
			p.DataType = "int"
			p.AccessMode = string(property.Type.Int.AccessMode)
			p.DefaultValue = property.Type.Int.DefaultValue
			p.Minimum = property.Type.Int.Minimum
			p.Maximum = property.Type.Int.Maximum
			p.Unit = property.Type.Int.Unit
		} else if property.Type.Double != nil {
			p.DataType = "double"
			p.AccessMode = string(property.Type.Double.AccessMode)
			p.DefaultValue = property.Type.Double.DefaultValue
			p.Minimum = int64(property.Type.Double.Minimum)
			p.Maximum = int64(property.Type.Double.Maximum)
			p.Unit = property.Type.Double.Unit
		} else if property.Type.Float != nil {
			p.DataType = "float"
			p.AccessMode = string(property.Type.Float.AccessMode)
			p.DefaultValue = property.Type.Float.DefaultValue
			p.Minimum = int64(property.Type.Float.Minimum)
			p.Maximum = int64(property.Type.Float.Maximum)
			p.Unit = property.Type.Float.Unit
		}
		properties = append(properties, p)
	}
	cur.Properties = properties
	return cur
}

func ParseDevice(device *v1alpha2.Device, commonModel *common.DeviceModel) (*common.DeviceInstance, error) {
	protocolName, err := getProtocolName(device)
	if err != nil {
		return nil, err
	}
	instance := &common.DeviceInstance{
		ID:               device.Name,
		Name:             device.Name,
		ProtocolName:     protocolName + "-" + device.Name,
		Model:            device.Spec.DeviceModelRef.Name,
		Twins:            buildTwins(device),
		Datas:            buildData(device),
		PropertyVisitors: buildPropertyVisitors(device),
	}
	propertyVisitorMap := make(map[string]common.PropertyVisitor)
	for i := 0; i < len(instance.PropertyVisitors); i++ {
		if commonModel == nil {
			continue
		}

		for _, property := range commonModel.Properties {
			if property.Name == instance.PropertyVisitors[i].PropertyName {
				instance.PropertyVisitors[i].PProperty = property
				break
			}
		}
		propertyVisitorMap[instance.PropertyVisitors[i].PProperty.Name] = instance.PropertyVisitors[i]
	}
	for i := 0; i < len(instance.Twins); i++ {
		if v, ok := propertyVisitorMap[instance.Twins[i].PropertyName]; ok {
			instance.Twins[i].PVisitor = &v
		}
	}
	for i := 0; i < len(instance.Datas.Properties); i++ {
		if v, ok := propertyVisitorMap[instance.Datas.Properties[i].PropertyName]; ok {
			instance.Datas.Properties[i].PVisitor = &v
		}
	}
	return instance, nil
}

func ConvTwinsToGrpc(twins []common.Twin) ([]*dmiapi.Twin, error) {
	res := make([]*dmiapi.Twin, 0, len(twins))
	for _, twin := range twins {
		cur := &dmiapi.Twin{
			PropertyName: twin.PropertyName,
			Desired: &dmiapi.TwinProperty{
				Value: twin.Desired.Value,
				Metadata: map[string]string{
					"type":      twin.Desired.Metadatas.Type,
					"timestamp": twin.Desired.Metadatas.Timestamp,
				},
			},
			Reported: &dmiapi.TwinProperty{
				Value: twin.Reported.Value,
				Metadata: map[string]string{
					"type":      twin.Reported.Metadatas.Type,
					"timestamp": twin.Reported.Metadatas.Timestamp,
				},
			},
		}
		res = append(res, cur)
	}
	return res, nil
}

func ConvGrpcToTwins(twins []*dmiapi.Twin, srcTwins []common.Twin) ([]common.Twin, error) {
	res := make([]common.Twin, 0, len(twins))
	for _, twin := range twins {
		var srcTwin common.Twin
		for _, found := range srcTwins {
			if twin.GetPropertyName() == found.PropertyName {
				srcTwin = found
				break
			}
		}
		if srcTwin.PropertyName == "" {
			return nil, fmt.Errorf("not found src twin name %s while update status", twin.GetPropertyName())
		}
		desiredMeta := twin.Desired.GetMetadata()
		reportedMeta := twin.Reported.GetMetadata()
		cur := common.Twin{
			PropertyName: twin.GetPropertyName(),
			PVisitor:     srcTwin.PVisitor,
			Desired: common.DesiredData{
				Value: twin.Desired.GetValue(),
			},
			Reported: common.ReportedData{
				Value: twin.Reported.GetValue(),
			},
		}
		if desiredMeta != nil {
			cur.Desired.Metadatas = common.Metadata{
				Timestamp: twin.Desired.GetMetadata()["timestamp"],
				Type:      twin.Desired.GetMetadata()["type"],
			}
		}
		if reportedMeta != nil {
			cur.Reported.Metadatas = common.Metadata{
				Timestamp: twin.Reported.GetMetadata()["timestamp"],
				Type:      twin.Reported.GetMetadata()["type"],
			}
		}
		res = append(res, cur)
	}
	return res, nil
}

func ConvMsgTwinToGrpc(msgTwin map[string]*common.MsgTwin) []*dmiapi.Twin {
	var twins []*dmiapi.Twin
	for name, twin := range msgTwin {
		twinData := &dmiapi.Twin{
			PropertyName: name,
			Desired: &dmiapi.TwinProperty{
				Value: *twin.Expected.Value,
				Metadata: map[string]string{
					"type":      twin.Metadata.Type,
					"timestamp": twin.Expected.Metadata.Timestamp,
				}},
			Reported: &dmiapi.TwinProperty{
				Value: *twin.Actual.Value,
				Metadata: map[string]string{
					"type":      twin.Metadata.Type,
					"timestamp": twin.Actual.Metadata.Timestamp,
				}},
		}
		twins = append(twins, twinData)
	}

	return twins
}
