package adapters

import (
	"context"

	"google.golang.org/grpc"
	scenariov1 "platform-service/pkg/pb/scenario/v1"
)

type ScenarioClient struct {
	Server scenariov1.ScenarioServiceServer
}

func (c ScenarioClient) ListScenarios(ctx context.Context, in *scenariov1.ListScenariosRequest, _ ...grpc.CallOption) (*scenariov1.ListScenariosResponse, error) {
	return c.Server.ListScenarios(ctx, in)
}

func (c ScenarioClient) GetScenario(ctx context.Context, in *scenariov1.GetScenarioRequest, _ ...grpc.CallOption) (*scenariov1.GetScenarioResponse, error) {
	return c.Server.GetScenario(ctx, in)
}

func (c ScenarioClient) SaveScenario(ctx context.Context, in *scenariov1.SaveScenarioRequest, _ ...grpc.CallOption) (*scenariov1.SaveScenarioResponse, error) {
	return c.Server.SaveScenario(ctx, in)
}

func (c ScenarioClient) EvaluateEvent(ctx context.Context, in *scenariov1.EvaluateEventRequest, _ ...grpc.CallOption) (*scenariov1.EvaluateEventResponse, error) {
	return c.Server.EvaluateEvent(ctx, in)
}

func (c ScenarioClient) GetOfflineScenarios(ctx context.Context, in *scenariov1.GetOfflineScenariosRequest, _ ...grpc.CallOption) (*scenariov1.GetOfflineScenariosResponse, error) {
	return c.Server.GetOfflineScenarios(ctx, in)
}

func (c ScenarioClient) ListVoiceCommands(ctx context.Context, in *scenariov1.ListVoiceCommandsRequest, _ ...grpc.CallOption) (*scenariov1.ListVoiceCommandsResponse, error) {
	return c.Server.ListVoiceCommands(ctx, in)
}

func (c ScenarioClient) ExecuteVoiceCommand(ctx context.Context, in *scenariov1.ExecuteVoiceCommandRequest, _ ...grpc.CallOption) (*scenariov1.ExecuteVoiceCommandResponse, error) {
	return c.Server.ExecuteVoiceCommand(ctx, in)
}
