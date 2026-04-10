package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"device-service/internal/httpapi"
	"device-service/internal/service"
	"device-service/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func (a *App) init() error {
	pool, err := pgxpool.New(context.Background(), a.cfg.PostgresDSN())
	if err != nil {
		return fmt.Errorf("open postgres pool: %w", err)
	}
	a.pool = pool

	deviceStore := store.New(pool)
	if err := deviceStore.Migrate(context.Background()); err != nil {
		return fmt.Errorf("migrate device storage: %w", err)
	}

	a.store = deviceStore
	a.service = service.New(deviceStore)

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
