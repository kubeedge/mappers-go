package parse

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/kubeedge/kubeedge/cloud/pkg/devicecontroller/controller"
	v1 "github.com/kubeedge/mappers-go/pkg/apis/downstream/v1"
	"github.com/kubeedge/mappers-go/pkg/common"
)

func getProtocolNameFromGrpc(device *v1.Device) (string, error) {
	if device.Spec.Protocol.Modbus != nil {
		return controller.Modbus, nil
	}
	if device.Spec.Protocol.Opcua != nil {
		return controller.OPCUA, nil
	}
	if device.Spec.Protocol.Bluetooth != nil {
		return controller.Bluetooth, nil
	}
	if device.Spec.Protocol.CustomizedProtocol != nil {
		return controller.CustomizedProtocol, nil
	}
	return "", errors.New("can not parse device protocol")
}

func BuildProtocolFromGrpc(device *v1.Device) (common.Protocol, error) {
	protocolName, err := getProtocolNameFromGrpc(device)
	if err != nil {
		return common.Protocol{}, err
	}
	protocolCommonConfig, err := json.Marshal(device.Spec.Protocol.Common)
	if err != nil {
		return common.Protocol{}, err
	}
	var protocolConfig []byte
	switch protocolName {
	case controller.Modbus:
		protocolConfig, err = json.Marshal(device.Spec.Protocol.Modbus)
		if err != nil {
			return common.Protocol{}, err
		}
	case controller.OPCUA:
		protocolConfig, err = json.Marshal(device.Spec.Protocol.Opcua)
		if err != nil {
			return common.Protocol{}, err
		}
	case controller.Bluetooth:
		protocolConfig, err = json.Marshal(device.Spec.Protocol.Bluetooth)
		if err != nil {
			return common.Protocol{}, err
		}
	case controller.CustomizedProtocol:
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

func buildTwinsFromGrpc(device *v1.Device) []common.Twin {
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

func buildDataFromGrpc(device *v1.Device) common.Data {
	res := common.Data{}
	if len(device.Spec.PropertyVisitors) > 0 {
		res.Properties = make([]common.DataProperty, 0, len(device.Spec.PropertyVisitors))
		for _, property := range device.Spec.PropertyVisitors {
			cur := common.DataProperty{
				Metadatas:    common.DataMetadata{},
				PropertyName: property.PropertyName,
				PVisitor:     nil,
			}
			timestamp, ok := property.CustomizedValues.Data["timestamp"]
			if ok {
				t, _ := strconv.ParseInt(string(timestamp.GetValue()), 10, 64)
				cur.Metadatas.Timestamp = t
			}
			tpe, ok := property.CustomizedValues.Data["type"]
			if ok {
				cur.Metadatas.Type = string(tpe.GetValue())
			}
			res.Properties = append(res.Properties, cur)
		}
	}
	return res
}

func buildPropertyVisitorsFromGrpc(device *v1.Device) []common.PropertyVisitor {
	if len(device.Spec.PropertyVisitors) == 0 {
		return nil
	}
	protocolName, err := getProtocolNameFromGrpc(device)
	if err != nil {
		return nil
	}
	res := make([]common.PropertyVisitor, 0, len(device.Spec.PropertyVisitors))
	for _, pptv := range device.Spec.PropertyVisitors {
		var visitorConfig []byte
		switch protocolName {
		case controller.Modbus:
			visitorConfig, err = json.Marshal(pptv.Modbus)
			if err != nil {
				return nil
			}
		case controller.OPCUA:
			visitorConfig, err = json.Marshal(pptv.Opcua)
			if err != nil {
				return nil
			}
		case controller.Bluetooth:
			visitorConfig, err = json.Marshal(pptv.Bluetooth)
			if err != nil {
				return nil
			}
		case controller.CustomizedProtocol:
			visitorConfig, err = json.Marshal(pptv.CustomizedProtocol)
			if err != nil {
				return nil
			}
		}

		collectCycle, err := strconv.ParseInt(pptv.GetCollectCycle(), 10, 64)
		if err != nil {
			collectCycle = common.DefaultCollectCycle.Nanoseconds()
		}
		reportCycle, err := strconv.ParseInt(pptv.GetReportCycle(), 10, 64)
		if err != nil {
			reportCycle = common.DefaultReportCycle.Nanoseconds()
		}

		cur := common.PropertyVisitor{
			Name:          pptv.PropertyName,
			PropertyName:  pptv.PropertyName,
			ModelName:     device.Spec.GetDeviceModelRef(),
			CollectCycle:  collectCycle,
			ReportCycle:   reportCycle,
			Protocol:      protocolName,
			VisitorConfig: visitorConfig,
		}
		res = append(res, cur)
	}
	return res
}

func ParseDeviceModelFromGrpc(model *v1.DeviceModel) common.DeviceModel {
	cur := common.DeviceModel{
		Name: model.GetName(),
	}
	if model.GetSpec() == nil || len(model.GetSpec().GetProperties()) == 0 {
		return cur
	}
	properties := make([]common.Property, 0, len(model.Spec.Properties))
	for _, property := range model.Spec.Properties {
		p := common.Property{
			Name:        property.GetName(),
			Description: property.GetDescription(),
		}
		if property.Type.GetString_() != nil {
			p.DataType = "string"
			p.AccessMode = property.Type.String_.GetAccessMode()
			p.DefaultValue = property.Type.String_.GetDefaultValue()
		} else if property.Type.GetBytes() != nil {
			p.DataType = "bytes"
			p.AccessMode = property.Type.Bytes.GetAccessMode()
		} else if property.Type.GetBoolean() != nil {
			p.DataType = "boolean"
			p.AccessMode = property.Type.Boolean.GetAccessMode()
			p.DefaultValue = property.Type.Boolean.GetDefaultValue()
		} else if property.Type.GetInt() != nil {
			p.DataType = "int"
			p.AccessMode = property.Type.Int.GetAccessMode()
			p.DefaultValue = property.Type.Int.GetDefaultValue()
			p.Minimum = property.Type.Int.Minimum
			p.Maximum = property.Type.Int.Maximum
			p.Unit = property.Type.Int.Unit
		} else if property.Type.GetDouble() != nil {
			p.DataType = "double"
			p.AccessMode = property.Type.Double.GetAccessMode()
			p.DefaultValue = property.Type.Double.GetDefaultValue()
			p.Minimum = int64(property.Type.Double.Minimum)
			p.Maximum = int64(property.Type.Double.Maximum)
			p.Unit = property.Type.Double.Unit
		} else if property.Type.GetFloat() != nil {
			p.DataType = "float"
			p.AccessMode = property.Type.Float.GetAccessMode()
			p.DefaultValue = property.Type.Float.GetDefaultValue()
			p.Minimum = int64(property.Type.Float.Minimum)
			p.Maximum = int64(property.Type.Float.Maximum)
			p.Unit = property.Type.Float.Unit
		}
		properties = append(properties, p)
	}
	cur.Properties = properties
	return cur
}

func ParseDeviceFromGrpc(device *v1.Device, commonModel *common.DeviceModel) (*common.DeviceInstance, error) {
	protocolName, err := getProtocolNameFromGrpc(device)
	if err != nil {
		return nil, err
	}
	instance := &common.DeviceInstance{
		ID:               device.GetName(),
		Name:             device.GetName(),
		ProtocolName:     protocolName + "-" + device.GetName(),
		Model:            device.Spec.GetDeviceModelRef(),
		Twins:            buildTwinsFromGrpc(device),
		Datas:            buildDataFromGrpc(device),
		PropertyVisitors: buildPropertyVisitorsFromGrpc(device),
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
