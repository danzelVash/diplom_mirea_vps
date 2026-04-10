package v1

import (
	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	notificationv1 "notification-service/pkg/pb/notification/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
)

type Implementation struct {
	scenariov1.UnimplementedScenarioServiceServer
	device       externalDeviceClient
	context      externalContextClient
	notification externalNotificationClient
}

type externalDeviceClient interface {
	devicev1.DeviceServiceClient
}

type externalContextClient interface {
	contextv1.ContextServiceClient
}

type externalNotificationClient interface {
	notificationv1.NotificationServiceClient
}

func New(
	device externalDeviceClient,
	context externalContextClient,
	notification externalNotificationClient,
) *Implementation {
	return &Implementation{
		device:       device,
		context:      context,
		notification: notification,
	}
}
