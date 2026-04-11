package internal

import (
	"fmt"
	"net"

	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	"voice-service/config"
	voiceservice "voice-service/internal/service"
	voicerecognitionv1 "voice-service/pkg/pb/voice_recognition/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (a *App) init() error {
	for name, target := range a.cfg.Targets {
		conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("dial %s: %w", name, err)
		}
		a.grpcConn[name] = conn
	}

	lis, err := net.Listen("tcp", a.grpcAddr())
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	a.grpcListener = lis
	a.grpcServer = grpc.NewServer()
	a.scenarioClient = scenariov1.NewScenarioServiceClient(a.grpcConn[config.ScenarioService])
	a.voiceRecognitionClient = voicerecognitionv1.NewVoiceRecognitionServiceClient(a.grpcConn[config.VoiceRecognitionService])
	a.service = voiceservice.New(a.scenarioClient, a.voiceRecognitionClient)
	return nil
}
