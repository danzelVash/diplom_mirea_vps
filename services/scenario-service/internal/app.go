package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"scenario-service/config"
	scenariohandler "scenario-service/internal/app/scenario/v1"
	"scenario-service/internal/service"
	"scenario-service/internal/store"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"

	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"

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

	pool    *pgxpool.Pool
	store   *store.Store
	service *service.Service

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
	if err := app.init(); err != nil {
		panic(err)
	}
	return app
}

func (a *App) Run(_ context.Context) {
	scenariov1.RegisterScenarioServiceServer(a.grpcServer, scenariohandler.New(a.service))

	slog.Info("service started",
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
