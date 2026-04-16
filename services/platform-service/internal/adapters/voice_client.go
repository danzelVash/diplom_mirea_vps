package adapters

import (
	"context"

	"google.golang.org/grpc"
	voicev1 "platform-service/pkg/pb/voice/v1"
)

type VoiceClient struct {
	Server voicev1.VoiceServiceServer
}

func (c VoiceClient) ParseVoiceCommand(ctx context.Context, in *voicev1.ParseVoiceCommandRequest, _ ...grpc.CallOption) (*voicev1.ParseVoiceCommandResponse, error) {
	return c.Server.ParseVoiceCommand(ctx, in)
}

func (c VoiceClient) MatchOfflinePhrase(ctx context.Context, in *voicev1.MatchOfflinePhraseRequest, _ ...grpc.CallOption) (*voicev1.MatchOfflinePhraseResponse, error) {
	return c.Server.MatchOfflinePhrase(ctx, in)
}
