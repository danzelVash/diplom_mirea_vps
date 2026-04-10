package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"vision-service/config"
	visionhandler "vision-service/internal/app/vision/v1"
	visionv1 "vision-service/pkg/pb/vision/v1"

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
	visionv1.RegisterVisionServiceServer(a.grpcServer, visionhandler.New())

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
