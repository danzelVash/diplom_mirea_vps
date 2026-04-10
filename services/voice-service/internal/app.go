package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	"voice-service/config"
	voicehandler "voice-service/internal/app/voice/v1"
	voiceservice "voice-service/internal/service"
	voicev1 "voice-service/pkg/pb/voice/v1"
	voicerecognitionv1 "voice-service/pkg/pb/voice_recognition/v1"

	"google.golang.org/grpc"
)

type App struct {
	name         string
	cfg          config.Config
	grpcServer   *grpc.Server
	grpcListener net.Listener
	grpcConn     map[string]*grpc.ClientConn

	scenarioClient         scenariov1.ScenarioServiceClient
	voiceRecognitionClient voicerecognitionv1.VoiceRecognitionServiceClient
	service                *voiceservice.Service
}

func New() *App {
	app := &App{
		name:     config.AppName,
		cfg:      config.LoadDefault(),
		grpcConn: make(map[string]*grpc.ClientConn),
	}
	if err := app.init(); err != nil {
		panic(err)
	}
	return app
}

func (a *App) Run(_ context.Context) {
	voicev1.RegisterVoiceServiceServer(a.grpcServer, voicehandler.New(a.service))

	slog.Info("service started",
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
