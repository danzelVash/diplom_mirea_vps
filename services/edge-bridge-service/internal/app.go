package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"edge-bridge-service/config"
	edgebridgehandler "edge-bridge-service/internal/app/edge_bridge/v1"
	bridgeservice "edge-bridge-service/internal/service"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"

	devicev1 "device-service/pkg/pb/device/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type App struct {
	name         string
	cfg          config.Config
	grpcServer   *grpc.Server
	grpcListener net.Listener
	httpServer   *http.Server
	httpListener net.Listener
	grpcConn     map[string]*grpc.ClientConn
	pool         *pgxpool.Pool

	deviceClient   devicev1.DeviceServiceClient
	scenarioClient scenariov1.ScenarioServiceClient
	voiceClient    voicev1.VoiceServiceClient
	service        *bridgeservice.Service
}

func New() *App {
	app := &App{
		name:     config.AppName,
		cfg:      config.LoadDefault(),
		grpcConn: make(map[string]*grpc.ClientConn),
	}
	if err := app.init(); err != nil {
		panic(err)
	}
	return app
}

func (a *App) Run(_ context.Context) {
	edgebridgev1.RegisterEdgeBridgeServiceServer(a.grpcServer, edgebridgehandler.New(a.service))

	slog.Info("service skeleton started",
		"service", a.name,
		"role", a.cfg.Role,
		"http_port", a.cfg.HTTP.Port,
		"grpc_port", a.cfg.GRPC.Port,
	)

	go func() {
		if err := a.httpServer.Serve(a.httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped", "service", a.name, "error", err.Error())
		}
	}()

	if err := a.grpcServer.Serve(a.grpcListener); err != nil {
		slog.Error("grpc server stopped", "service", a.name, "error", err.Error())
	}
}

func (a *App) grpcAddr() string {
	return fmt.Sprintf("%s:%d", a.cfg.GRPC.Host, a.cfg.GRPC.Port)
}

func (a *App) httpAddr() string {
	return fmt.Sprintf("%s:%d", a.cfg.HTTP.Host, a.cfg.HTTP.Port)
}
