package internal

import (
	"fmt"
	"net"

	"context-service/config"
	devicev1 "device-service/pkg/pb/device/v1"
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

	a.deviceClient = devicev1.NewDeviceServiceClient(a.grpcConn[config.DeviceService])
	a.voiceClient = voicev1.NewVoiceServiceClient(a.grpcConn[config.VoiceService])
	a.visionClient = visionv1.NewVisionServiceClient(a.grpcConn[config.VisionService])
	return nil
}
