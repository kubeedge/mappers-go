package register

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/kubeedge/mappers-go/config"
	dmiapi "github.com/kubeedge/mappers-go/pkg/apis/dmi/v1"
	"github.com/kubeedge/mappers-go/pkg/common"
	"google.golang.org/grpc"
)

// RegisterMapper if withData is true, edgecore will send device and model list.
func RegisterMapper(cfg *config.Config, withData bool) ([]*dmiapi.Device, []*dmiapi.DeviceModel, error) {
	// 连接grpc服务器
	//conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	conn, err := grpc.Dial(cfg.Common.EdgeCoreSock,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				unixAddress, err := net.ResolveUnixAddr("unix", cfg.Common.EdgeCoreSock)
				if err != nil {
					return nil, err
				}
				return net.DialUnix("unix", nil, unixAddress)
			},
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("did not connect: %v", err)
	}
	// 延迟关闭连接
	defer conn.Close()

	// 初始化Greeter服务客户端
	c := dmiapi.NewDeviceManagerServiceClient(conn)

	// 初始化上下文，设置请求超时时间为1秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 延迟关闭请求会话
	defer cancel()

	// 调用SayHello接口，发送一条消息
	resp, err := c.MapperRegister(ctx, &dmiapi.MapperRegisterRequest{
		WithData: withData,
		Mapper: &dmiapi.MapperInfo{
			Name:       cfg.Common.Name,
			Version:    cfg.Common.Version,
			ApiVersion: cfg.Common.APIVersion,
			Protocol:   cfg.Common.Protocol,
			Address:    []byte(cfg.GrpcServer.SocketPath),
			State:      common.DEVSTOK,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return resp.DeviceList, resp.ModelList, err
}
