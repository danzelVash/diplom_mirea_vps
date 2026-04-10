package v1

import (
	gatewayv1 "api-gateway/pkg/pb/api_gateway/v1"
	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	visionv1 "vision-service/pkg/pb/vision/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"
)

type Implementation struct {
	gatewayv1.UnimplementedGatewayServiceServer
	edgeBridge   externalEdgeBridgeClient
	device       externalDeviceClient
	context      externalContextClient
	scenario     externalScenarioClient
	voice        externalVoiceClient
	vision       externalVisionClient
	notification externalNotificationClient
}

type externalEdgeBridgeClient interface {
	edgebridgev1.EdgeBridgeServiceClient
}

type externalDeviceClient interface {
	devicev1.DeviceServiceClient
}

type externalContextClient interface {
	contextv1.ContextServiceClient
}

type externalScenarioClient interface {
	scenariov1.ScenarioServiceClient
}

type externalVoiceClient interface {
	voicev1.VoiceServiceClient
}

type externalVisionClient interface {
	visionv1.VisionServiceClient
}

type externalNotificationClient interface {
	notificationv1.NotificationServiceClient
}

func New(
	edgeBridge externalEdgeBridgeClient,
	device externalDeviceClient,
	context externalContextClient,
	scenario externalScenarioClient,
	voice externalVoiceClient,
	vision externalVisionClient,
	notification externalNotificationClient,
) *Implementation {
	return &Implementation{
		edgeBridge:   edgeBridge,
		device:       device,
		context:      context,
		scenario:     scenario,
		voice:        voice,
		vision:       vision,
		notification: notification,
	}
}
