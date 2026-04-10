package v1

import voicev1 "voice-service/pkg/pb/voice/v1"

type Implementation struct {
	voicev1.UnimplementedVoiceServiceServer
}

func New() *Implementation {
	return &Implementation{}
}
