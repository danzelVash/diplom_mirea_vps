package internal

import (
	"fmt"
	"net"
	"net/http"

	devicev1 "device-service/pkg/pb/device/v1"
	"edge-bridge-service/config"
	"edge-bridge-service/internal/httpapi"
	bridgeservice "edge-bridge-service/internal/service"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (a *App) init() error {
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
	a.service = bridgeservice.New(a.deviceClient, a.scenarioClient)

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
