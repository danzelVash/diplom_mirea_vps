package v1

import notificationv1 "notification-service/pkg/pb/notification/v1"

type Implementation struct {
	notificationv1.UnimplementedNotificationServiceServer
}

func New() *Implementation {
	return &Implementation{}
}
