package grpcserver

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/kubeedge/mappers-go/pkg/apis/downstream/v1"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/modbus"
	"github.com/kubeedge/mappers-go/pkg/util/parse"

	"k8s.io/klog/v2"
)

func (s *Server) CreateDevice(ctx context.Context, request *pb.CreateDeviceRequest) (*pb.CreateDeviceResponse, error) {
	klog.V(2).Info("CreateDevice")
	config := request.GetConfig()
	if config == nil || config.Device == nil || config.Model == nil {
		return nil, errors.New("device config is nil")
	}
	device := config.Device
	deviceModel := config.Model
	if _, err := s.devPanel.GetDevice(device.Name); err == nil {
		return nil, fmt.Errorf("add device %s failed, has existed", device.Name)
	}

	model := parse.ParseDeviceModelFromGrpc(deviceModel)
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

	return &pb.CreateDeviceResponse{DeviceName: device.Name}, nil
}

func (s *Server) RemoveDevice(ctx context.Context, request *pb.RemoveDeviceRequest) (*pb.RemoveDeviceResponse, error) {
	if request.GetDeviceName() == "" {
		return nil, errors.New("device name is nil")
	}

	return nil, s.devPanel.RemoveDevice(request.GetDeviceName())
}

func (s *Server) UpdateDevice(ctx context.Context, request *pb.UpdateDeviceRequest) (*pb.UpdateDeviceResponse, error) {
	klog.V(2).Info("UpdateDevice")
	config := request.GetConfig()
	if config == nil || config.Device == nil || config.Model == nil {
		return nil, errors.New("device config is nil")
	}
	device := config.Device
	deviceModel := config.Model
	if _, err := s.devPanel.GetDevice(device.Name); err != nil {
		return nil, fmt.Errorf("add device %s failed, not existed", device.Name)
	}

	model := parse.ParseDeviceModelFromGrpc(deviceModel)
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

	return &pb.UpdateDeviceResponse{}, nil
}

func (s *Server) UpdateDeviceStatus(ctx context.Context, request *pb.UpdateDeviceStatusRequest) (*pb.UpdateDeviceStatusResponse, error) {
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

func (s *Server) GetDevice(ctx context.Context, request *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	if request.GetDeviceName() == "" {
		return nil, errors.New("device name is nil")
	}

	device, err := s.devPanel.GetDevice(request.GetDeviceName())
	if err != nil {
		return nil, err
	}
	res := &pb.GetDeviceResponse{Status: &pb.DeviceStatus{}}
	switch s.cfg.Protocol {
	case common.ProtocolModbus:
		d := device.(*modbus.ModbusDev)
		twins, err := parse.ConvTwinsToGrpc(d.Instance.Twins)
		if err != nil {
			return nil, err
		}
		res.Status.Twins = twins
		res.Status.State = common.DEVSTOK
	default:
		return nil, fmt.Errorf("current mapper only support protocol %s's device", s.cfg.Protocol)
	}
	return res, nil
}
