package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"edge-bridge-service/config"
	edgebridgehandler "edge-bridge-service/internal/app/edge_bridge/v1"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"

	devicev1 "device-service/pkg/pb/device/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"

	"google.golang.org/grpc"
)

type App struct {
	name         string
	cfg          config.Config
	grpcServer   *grpc.Server
	grpcListener net.Listener
	grpcConn     map[string]*grpc.ClientConn

	deviceClient   devicev1.DeviceServiceClient
	scenarioClient scenariov1.ScenarioServiceClient
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
	edgebridgev1.RegisterEdgeBridgeServiceServer(a.grpcServer, edgebridgehandler.New(
		a.deviceClient,
		a.scenarioClient,
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
