package grpcserver

import (
	"context"
	"errors"
	"fmt"

	dmiapi "github.com/kubeedge/mappers-go/pkg/apis/dmi/v1"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/modbus"
	"github.com/kubeedge/mappers-go/pkg/util/parse"

	"k8s.io/klog/v2"
)

func (s *Server) CreateDevice(ctx context.Context, request *dmiapi.CreateDeviceRequest) (*dmiapi.CreateDeviceResponse, error) {
	klog.V(2).Info("CreateDevice")
	device := request.GetDevice()
	if device == nil {
		return nil, errors.New("device is nil")
	}
	if _, err := s.devPanel.GetDevice(device.Name); err == nil {
		return nil, fmt.Errorf("add device %s failed, has existed", device.Name)
	}

	model, err := s.devPanel.GetModel(device.Spec.DeviceModelReference)
	if err != nil {
		return nil, fmt.Errorf("deviceModel %s not found, err: %s", device.Spec.DeviceModelReference, err)
	}
	protocol, err := parse.BuildProtocolFromGrpc(device)
	if err != nil {
		return nil, fmt.Errorf("parse device %s protocol failed, err: %s", device.Name, err)
	}
	deviceInstance, err := parse.ParseDeviceFromGrpc(device, &model)
	if err != nil {
		return nil, fmt.Errorf("parse device %s instance failed, err: %s", device.Name, err)
	}
	deviceInstance.PProtocol = protocol

	s.devPanel.UpdateDev(&model, deviceInstance, &protocol)

	// TODO need edgecore publish?
	// publish device twin to mqtt
	//topic := dtcommon.DeviceETPrefix + device.Name + dtcommon.DeviceETUpdatedSuffix
	// publish device to mqtt
	//topic := dtcommon.MemETPrefix + d.NodeName + dtcommon.MemETUpdateSuffix

	return &dmiapi.CreateDeviceResponse{DeviceName: device.Name}, nil
}

func (s *Server) RemoveDevice(ctx context.Context, request *dmiapi.RemoveDeviceRequest) (*dmiapi.RemoveDeviceResponse, error) {
	if request.GetDeviceName() == "" {
		return nil, errors.New("device name is nil")
	}

	return nil, s.devPanel.RemoveDevice(request.GetDeviceName())
}

func (s *Server) UpdateDevice(ctx context.Context, request *dmiapi.UpdateDeviceRequest) (*dmiapi.UpdateDeviceResponse, error) {
	klog.V(2).Info("UpdateDevice")
	device := request.GetDevice()
	if device == nil {
		return nil, errors.New("device is nil")
	}
	if _, err := s.devPanel.GetDevice(device.Name); err != nil {
		return nil, fmt.Errorf("update device %s failed, not existed", device.Name)
	}

	model, err := s.devPanel.GetModel(device.Spec.DeviceModelReference)
	if err != nil {
		return nil, fmt.Errorf("deviceModel %s not found, err: %s", device.Spec.DeviceModelReference, err)
	}
	protocol, err := parse.BuildProtocolFromGrpc(device)
	if err != nil {
		return nil, fmt.Errorf("parse device %s protocol failed, err: %s", device.Name, err)
	}
	deviceInstance, err := parse.ParseDeviceFromGrpc(device, &model)
	if err != nil {
		return nil, fmt.Errorf("parse device %s instance failed, err: %s", device.Name, err)
	}
	deviceInstance.PProtocol = protocol

	s.devPanel.UpdateDev(&model, deviceInstance, &protocol)

	// TODO need edgecore publish?
	// publish device twin to mqtt
	//topic := dtcommon.DeviceETPrefix + device.Name + dtcommon.DeviceETUpdatedSuffix
	// publish device to mqtt
	//topic := dtcommon.MemETPrefix + d.NodeName + dtcommon.MemETUpdateSuffix

	return &dmiapi.UpdateDeviceResponse{}, nil
}

func (s *Server) CreateDeviceModel(ctx context.Context, request *dmiapi.CreateDeviceModelRequest) (*dmiapi.CreateDeviceModelResponse, error) {
	deviceModel := request.GetModel()
	if deviceModel == nil {
		return nil, errors.New("deviceModel is nil")
	}
	if _, err := s.devPanel.GetModel(deviceModel.Name); err != nil {
		return nil, fmt.Errorf("add deviceModel %s failed, has existed", deviceModel.Name)
	}

	model := parse.ParseDeviceModelFromGrpc(deviceModel)

	s.devPanel.UpdateModel(&model)

	return &dmiapi.CreateDeviceModelResponse{DeviceModelName: deviceModel.Name}, nil
}

func (s *Server) UpdateDeviceModel(ctx context.Context, request *dmiapi.UpdateDeviceModelRequest) (*dmiapi.UpdateDeviceModelResponse, error) {
	deviceModel := request.GetModel()
	if deviceModel == nil {
		return nil, errors.New("deviceModel is nil")
	}
	if _, err := s.devPanel.GetModel(deviceModel.Name); err != nil {
		return nil, fmt.Errorf("update deviceModel %s failed, not existed", deviceModel.Name)
	}

	model := parse.ParseDeviceModelFromGrpc(deviceModel)

	s.devPanel.UpdateModel(&model)

	return &dmiapi.UpdateDeviceModelResponse{}, nil
}

func (s *Server) RemoveDeviceModel(ctx context.Context, request *dmiapi.RemoveDeviceModelRequest) (*dmiapi.RemoveDeviceModelResponse, error) {
	s.devPanel.RemoveModel(request.ModelName)

	return &dmiapi.RemoveDeviceModelResponse{}, nil
}

func (s *Server) UpdateDeviceStatus(ctx context.Context, request *dmiapi.UpdateDeviceStatusRequest) (*dmiapi.UpdateDeviceStatusResponse, error) {
	if request.GetDeviceName() == "" {
		return nil, errors.New("device name is nil")
	}

	deviceStatus := request.GetDesiredDevice()
	device, err := s.devPanel.GetDevice(request.GetDeviceName())
	if err != nil {
		return nil, err
	}

	// only can update twin property desired value
	switch s.cfg.Protocol {
	case common.ProtocolModbus:
		d := device.(*modbus.ModbusDev)
		twins, err := parse.ConvGrpcToTwins(deviceStatus.Twins, d.Instance.Twins)
		if err != nil {
			return nil, err
		}
		return nil, s.devPanel.UpdateDevTwins(request.GetDeviceName(), twins)
	default:
		return nil, fmt.Errorf("current mapper only support protocol %s's device", s.cfg.Protocol)
	}
}

func (s *Server) GetDevice(ctx context.Context, request *dmiapi.GetDeviceRequest) (*dmiapi.GetDeviceResponse, error) {
	if request.GetDeviceName() == "" {
		return nil, errors.New("device name is nil")
	}

	device, err := s.devPanel.GetDevice(request.GetDeviceName())
	if err != nil {
		return nil, err
	}
	res := &dmiapi.GetDeviceResponse{
		Device: &dmiapi.Device{
			Status: &dmiapi.DeviceStatus{},
		},
	}
	switch s.cfg.Protocol {
	case common.ProtocolModbus:
		d := device.(*modbus.ModbusDev)
		twins, err := parse.ConvTwinsToGrpc(d.Instance.Twins)
		if err != nil {
			return nil, err
		}
		res.Device.Status.Twins = twins
		res.Device.Status.State = common.DEVSTOK
	default:
		return nil, fmt.Errorf("current mapper only support protocol %s's device", s.cfg.Protocol)
	}
	return res, nil
}
