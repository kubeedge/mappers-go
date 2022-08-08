package grpcclient

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	dmiapi "github.com/kubeedge/mappers-go/pkg/apis/dmi/v1"
)

// ReportDeviceStatus report device status to edgecore
func ReportDeviceStatus(request *dmiapi.ReportDeviceStatusRequest) error {
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
		return fmt.Errorf("did not connect: %v", err)
	}
	// 延迟关闭连接
	defer conn.Close()

	// 初始化Greeter服务客户端
	c := dmiapi.NewDeviceManagerServiceClient(conn)

	// 初始化上下文，设置请求超时时间为1秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 延迟关闭请求会话
	defer cancel()

	_, err = c.ReportDeviceStatus(ctx, request)
	return err
}
