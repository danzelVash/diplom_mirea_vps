package v1

import visionv1 "vision-service/pkg/pb/vision/v1"

type Implementation struct {
	visionv1.UnimplementedVisionServiceServer
}

func New() *Implementation {
	return &Implementation{}
}
