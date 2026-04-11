package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"

	devicev1 "device-service/pkg/pb/device/v1"
	"edge-bridge-service/config"
	"edge-bridge-service/internal/httpapi"
	bridgeservice "edge-bridge-service/internal/service"
	edgebridgestore "edge-bridge-service/internal/store"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (a *App) init() error {
	pool, err := pgxpool.New(context.Background(), a.cfg.PostgresDSN())
	if err != nil {
		return fmt.Errorf("open postgres pool: %w", err)
	}
	a.pool = pool

	bridgeStore := edgebridgestore.New(pool)
	if err := bridgeStore.Migrate(context.Background()); err != nil {
		return fmt.Errorf("migrate edge bridge storage: %w", err)
	}

	lis, err := net.Listen("tcp", a.grpcAddr())
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	a.grpcListener = lis
	a.grpcServer = grpc.NewServer()

	for name, target := range a.cfg.Targets {
		conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("dial %s: %w", name, err)
		}
		a.grpcConn[name] = conn
	}

	a.deviceClient = devicev1.NewDeviceServiceClient(a.grpcConn[config.DeviceService])
	a.scenarioClient = scenariov1.NewScenarioServiceClient(a.grpcConn[config.ScenarioService])
	a.voiceClient = voicev1.NewVoiceServiceClient(a.grpcConn[config.VoiceService])
	a.service = bridgeservice.New(bridgeStore, a.deviceClient, a.scenarioClient, a.voiceClient, a.cfg.RetryAfter)

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
