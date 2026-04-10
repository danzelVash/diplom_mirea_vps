package v1

import (
	contextv1 "context-service/pkg/pb/context/v1"
	devicev1 "device-service/pkg/pb/device/v1"
	visionv1 "vision-service/pkg/pb/vision/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"
)

type Implementation struct {
	contextv1.UnimplementedContextServiceServer
	device externalDeviceClient
	voice  externalVoiceClient
	vision externalVisionClient
}

type externalDeviceClient interface {
	devicev1.DeviceServiceClient
}

type externalVoiceClient interface {
	voicev1.VoiceServiceClient
}

type externalVisionClient interface {
	visionv1.VisionServiceClient
}

func New(device externalDeviceClient, voice externalVoiceClient, vision externalVisionClient) *Implementation {
	return &Implementation{
		device: device,
		voice:  voice,
		vision: vision,
	}
}
