package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	voiceservice "platform-service/internal/voice"
	voicev1 "platform-service/pkg/pb/voice/v1"
)

type Implementation struct {
	voicev1.UnimplementedVoiceServiceServer
	service *voiceservice.Service
}

func New(service *voiceservice.Service) *Implementation {
	return &Implementation{service: service}
}

func (i *Implementation) ParseVoiceCommand(ctx context.Context, req *voicev1.ParseVoiceCommandRequest) (*voicev1.ParseVoiceCommandResponse, error) {
	command, err := i.service.ParseAndExecute(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse voice command: %v", err)
	}
	return &voicev1.ParseVoiceCommandResponse{Command: command}, nil
}

func (i *Implementation) MatchOfflinePhrase(ctx context.Context, req *voicev1.MatchOfflinePhraseRequest) (*voicev1.MatchOfflinePhraseResponse, error) {
	command, err := i.service.ExecutePhrase(ctx, req.GetPhrase(), req.GetEdgeId(), req.GetRoomId(), req.GetSource())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "match offline phrase: %v", err)
	}
	return &voicev1.MatchOfflinePhraseResponse{Command: command}, nil
}
