package grpcserver

import (
	"fmt"
	"net"
	"os"

	modbusdevice "github.com/kubeedge/mappers-go/mappers/modbus/device"
	pb "github.com/kubeedge/mappers-go/pkg/apis/downstream/v1"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/global"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog/v2"
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
	f, err := os.Stat(s.cfg.SockPath)
	if err != nil {
		return err
	}
	if f.IsDir() {
		return fmt.Errorf("socket file path is a directory")
	}

	lis, err := net.Listen("unix", s.cfg.SockPath)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterDeviceMapperServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	return grpcServer.Serve(lis)
}
