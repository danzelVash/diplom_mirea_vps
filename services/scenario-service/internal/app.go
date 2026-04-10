package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"scenario-service/config"
	scenariohandler "scenario-service/internal/app/scenario/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"

	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"

	"google.golang.org/grpc"
)

type App struct {
	name         string
	cfg          config.Config
	grpcServer   *grpc.Server
	grpcListener net.Listener
	grpcConn     map[string]*grpc.ClientConn

	deviceClient       devicev1.DeviceServiceClient
	contextClient      contextv1.ContextServiceClient
	notificationClient notificationv1.NotificationServiceClient
}

func New() *App {
	app := &App{
		name:     config.AppName,
		cfg:      config.LoadDefault(),
		grpcConn: make(map[string]*grpc.ClientConn),
	}
	_ = app.init()
	return app
}

func (a *App) Run(_ context.Context) {
	scenariov1.RegisterScenarioServiceServer(a.grpcServer, scenariohandler.New(
		a.deviceClient,
		a.contextClient,
		a.notificationClient,
	))

	slog.Info("service skeleton started",
		"service", a.name,
		"role", a.cfg.Role,
		"http_port", a.cfg.HTTP.Port,
		"grpc_port", a.cfg.GRPC.Port,
	)

	if err := a.grpcServer.Serve(a.grpcListener); err != nil {
		slog.Error("grpc server stopped", "service", a.name, "error", err.Error())
	}
}

func (a *App) grpcAddr() string {
	return fmt.Sprintf("%s:%d", a.cfg.GRPC.Host, a.cfg.GRPC.Port)
}
