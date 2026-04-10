package v1

import devicev1 "device-service/pkg/pb/device/v1"

type Implementation struct {
	devicev1.UnimplementedDeviceServiceServer
}

func New() *Implementation {
	return &Implementation{}
}
