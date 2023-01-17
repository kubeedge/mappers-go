package grpcserver

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog/v2"

	dmiapi "github.com/kubeedge/kubeedge/pkg/apis/dmi/v1alpha1"
	modbusdevice "github.com/kubeedge/mappers-go/mappers/modbus-dmi/device"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/global"
)

type Config struct {
	SockPath string `json:"sock_path"`
	Protocol string `json:"protocol"`
}

type Server struct {
	cfg      Config
	devPanel global.DevPanel
}

func NewServer(cfg Config) Server {
	s := Server{cfg: cfg}
	switch cfg.Protocol {
	case common.ProtocolModbus:
		s.devPanel = modbusdevice.NewDevPanel()
	case common.ProtocolBlueTooth:
		// TODO
	case common.ProtocolOpcua:
		// TODO
	case common.ProtocolOnvif:
		// TODO
	case common.ProtocolCustomized:
		// TODO
	default:
		klog.Fatalf("unknown device protocol %s for grpc server", cfg.Protocol)
	}
	return s
}

func (s *Server) Start() error {
	klog.Infof("uds socket path: %s", s.cfg.SockPath)
	err := initSock(s.cfg.SockPath)
	if err != nil {
		klog.Fatalf("failed to remove uds socket with err: %v", err)
		return err
	}

	lis, err := net.Listen("unix", s.cfg.SockPath)
	if err != nil {
		klog.Fatalf("failed to remove uds socket with err: %v", err)
		return err
	}
	grpcServer := grpc.NewServer()
	dmiapi.RegisterDeviceMapperServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	return grpcServer.Serve(lis)
}

func initSock(sockPath string) error {
	klog.Infof("init uds socket: %s", sockPath)
	_, err := os.Stat(sockPath)
	if err == nil {
		err = os.Remove(sockPath)
		if err != nil {
			return err
		}
		return nil
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return fmt.Errorf("fail to stat uds socket path")
	}
}
