package adapters

import (
	"google.golang.org/grpc"
	voicev1 "platform-service/pkg/pb/voice_recognition/v1"
)

func NewRecognitionClient(addr string) (voicev1.VoiceRecognitionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	return voicev1.NewVoiceRecognitionServiceClient(conn), conn, nil
}
