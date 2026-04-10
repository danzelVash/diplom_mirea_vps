package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"api-gateway/config"
	gatewayhandler "api-gateway/internal/app/gateway/v1"
	gatewayv1 "api-gateway/pkg/pb/api_gateway/v1"

	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
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

	edgeBridgeClient   edgebridgev1.EdgeBridgeServiceClient
	deviceClient       devicev1.DeviceServiceClient
	contextClient      contextv1.ContextServiceClient
	scenarioClient     scenariov1.ScenarioServiceClient
	voiceClient        voicev1.VoiceServiceClient
	visionClient       visionv1.VisionServiceClient
	notificationClient notificationv1.NotificationServiceClient
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
	gatewayv1.RegisterGatewayServiceServer(a.grpcServer, gatewayhandler.New(
		a.edgeBridgeClient,
		a.deviceClient,
		a.contextClient,
		a.scenarioClient,
		a.voiceClient,
		a.visionClient,
		a.notificationClient,
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
