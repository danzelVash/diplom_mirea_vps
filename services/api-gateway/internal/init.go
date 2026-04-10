package internal

import (
	"fmt"
	"net"

	"api-gateway/config"
	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	visionv1 "vision-service/pkg/pb/vision/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"

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

	a.edgeBridgeClient = edgebridgev1.NewEdgeBridgeServiceClient(a.grpcConn[config.EdgeBridgeService])
	a.deviceClient = devicev1.NewDeviceServiceClient(a.grpcConn[config.DeviceService])
	a.contextClient = contextv1.NewContextServiceClient(a.grpcConn[config.ContextService])
	a.scenarioClient = scenariov1.NewScenarioServiceClient(a.grpcConn[config.ScenarioService])
	a.voiceClient = voicev1.NewVoiceServiceClient(a.grpcConn[config.VoiceService])
	a.visionClient = visionv1.NewVisionServiceClient(a.grpcConn[config.VisionService])
	a.notificationClient = notificationv1.NewNotificationServiceClient(a.grpcConn[config.NotificationService])
	return nil
}
