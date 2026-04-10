package internal

import (
	"fmt"
	"net"

	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"
	"scenario-service/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (a *App) init() error {
	lis, err := net.Listen("tcp", a.grpcAddr())
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	a.grpcListener = lis
	a.grpcServer = grpc.NewServer()

	for name, target := range a.cfg.Targets {
		conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("dial %s: %w", name, err)
		}
		a.grpcConn[name] = conn
	}

	a.deviceClient = devicev1.NewDeviceServiceClient(a.grpcConn[config.DeviceService])
	a.contextClient = contextv1.NewContextServiceClient(a.grpcConn[config.ContextService])
	a.notificationClient = notificationv1.NewNotificationServiceClient(a.grpcConn[config.NotificationService])
	return nil
}
