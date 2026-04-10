package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"device-service/config"
	devicev1 "device-service/pkg/pb/device/v1"

	devicehandler "device-service/internal/app/device/v1"

	"google.golang.org/grpc"
)

type App struct {
	name         string
	cfg          config.Config
	grpcServer   *grpc.Server
	grpcListener net.Listener
}

func New() *App {
	app := &App{
		name: config.AppName,
		cfg:  config.LoadDefault(),
	}
	_ = app.init()
	return app
}

func (a *App) Run(_ context.Context) {
	devicev1.RegisterDeviceServiceServer(a.grpcServer, devicehandler.New())

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
