package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"context-service/config"
	contexthandler "context-service/internal/app/context/v1"
	contextv1 "context-service/pkg/pb/context/v1"

	devicev1 "device-service/pkg/pb/device/v1"
	visionv1 "vision-service/pkg/pb/vision/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"

	"google.golang.org/grpc"
)

type App struct {
	name         string
	cfg          config.Config
	grpcServer   *grpc.Server
	grpcListener net.Listener
	grpcConn     map[string]*grpc.ClientConn

	deviceClient devicev1.DeviceServiceClient
	voiceClient  voicev1.VoiceServiceClient
	visionClient visionv1.VisionServiceClient
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
	contextv1.RegisterContextServiceServer(a.grpcServer, contexthandler.New(
		a.deviceClient,
		a.voiceClient,
		a.visionClient,
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
