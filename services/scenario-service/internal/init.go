package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"

	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"
	"scenario-service/config"
	"scenario-service/internal/httpapi"
	"scenario-service/internal/service"
	"scenario-service/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (a *App) init() error {
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

	pool, err := pgxpool.New(context.Background(), a.cfg.PostgresDSN())
	if err != nil {
		return fmt.Errorf("open postgres pool: %w", err)
	}
	a.pool = pool

	scenarioStore := store.New(pool)
	if err := scenarioStore.Migrate(context.Background()); err != nil {
		return fmt.Errorf("migrate scenario storage: %w", err)
	}
	a.store = scenarioStore
	a.service = service.New(scenarioStore, a.deviceClient)

	grpcLis, err := net.Listen("tcp", a.grpcAddr())
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	a.grpcListener = grpcLis
	a.grpcServer = grpc.NewServer()

	httpLis, err := net.Listen("tcp", a.httpAddr())
	if err != nil {
		return fmt.Errorf("listen http: %w", err)
	}
	a.httpListener = httpLis
	a.httpServer = &http.Server{
		Handler: httpapi.New(a.service),
	}

	return nil
}
