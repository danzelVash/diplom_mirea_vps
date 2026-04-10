package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"notification-service/config"
	notificationhandler "notification-service/internal/app/notification/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"

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
	notificationv1.RegisterNotificationServiceServer(a.grpcServer, notificationhandler.New())

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
