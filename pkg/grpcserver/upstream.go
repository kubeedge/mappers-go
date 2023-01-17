package grpcserver

import (
	"context"
	"errors"

	"k8s.io/klog/v2"

	dmiapi "github.com/kubeedge/kubeedge/pkg/apis/dmi/v1alpha1"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/driver/modbus"
)

func (s *Server) ReportDeviceStatus(ctx context.Context, request *dmiapi.ReportDeviceStatusRequest) (*dmiapi.ReportDeviceStatusResponse, error) {
	if request.GetDeviceName() == "" {
		return nil, errors.New("device name is nil while report device status")
	}
	device, err := s.devPanel.GetDevice(request.GetDeviceName())
	if err != nil {
		return nil, err
	}
	reportedTwins := make([]*dmiapi.Twin, 0)
	switch s.cfg.Protocol {
	case common.ProtocolModbus:
		d := device.(*modbus.ModbusDev)
		for _, twin := range d.Instance.Twins {
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
			reportedTwins = append(reportedTwins, cur)
		}
	case common.ProtocolBlueTooth:
		// TODO
	case common.ProtocolOpcua:
		// TODO
	case common.ProtocolOnvif:
		// TODO
	case common.ProtocolCustomized:
		// TODO
	default:
		klog.Fatalf("unknown device protocol %s for grpc server", s.cfg.Protocol)
	}

	// TODO report to edgecore
	return &dmiapi.ReportDeviceStatusResponse{}, nil
}
