package v1

import (
	"context"

	"platform-service/internal/scenario/model"
	"platform-service/internal/scenario/service"
	scenariov1 "platform-service/pkg/pb/scenario/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Implementation struct {
	scenariov1.UnimplementedScenarioServiceServer
	service *service.Service
}

func New(
	service *service.Service,
) *Implementation {
	return &Implementation{
		service: service,
	}
}

func (i *Implementation) ListScenarios(ctx context.Context, req *scenariov1.ListScenariosRequest) (*scenariov1.ListScenariosResponse, error) {
	scenarios, err := i.service.ListScenarios(ctx, req.GetEdgeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list scenarios: %v", err)
	}

	response := &scenariov1.ListScenariosResponse{
		Scenarios: make([]*scenariov1.Scenario, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		response.Scenarios = append(response.Scenarios, scenario.ToProto())
	}
	return response, nil
}

func (i *Implementation) GetScenario(ctx context.Context, req *scenariov1.GetScenarioRequest) (*scenariov1.GetScenarioResponse, error) {
	scenario, err := i.service.GetScenario(ctx, req.GetScenarioId())
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "get scenario: %v", err)
	}
	return &scenariov1.GetScenarioResponse{Scenario: scenario.ToProto()}, nil
}

func (i *Implementation) SaveScenario(ctx context.Context, req *scenariov1.SaveScenarioRequest) (*scenariov1.SaveScenarioResponse, error) {
	saved, err := i.service.SaveScenario(ctx, model.ScenarioFromProto(req.GetScenario()))
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.InvalidArgument, "save scenario: %v", err)
	}
	return &scenariov1.SaveScenarioResponse{Scenario: saved.ToProto()}, nil
}

func (i *Implementation) EvaluateEvent(ctx context.Context, req *scenariov1.EvaluateEventRequest) (*scenariov1.EvaluateEventResponse, error) {
	decision, err := i.service.EvaluateEvent(ctx, model.EventFromProto(req.GetEvent()), req.GetDeferExecution())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "evaluate event: %v", err)
	}
	return &scenariov1.EvaluateEventResponse{Decision: decision.ToProto()}, nil
}

func (i *Implementation) GetOfflineScenarios(ctx context.Context, req *scenariov1.GetOfflineScenariosRequest) (*scenariov1.GetOfflineScenariosResponse, error) {
	scenarios, err := i.service.GetOfflineScenarios(ctx, req.GetEdgeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get offline scenarios: %v", err)
	}

	response := &scenariov1.GetOfflineScenariosResponse{
		Scenarios: make([]*scenariov1.Scenario, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		response.Scenarios = append(response.Scenarios, scenario.ToProto())
	}
	return response, nil
}

func (i *Implementation) ListVoiceCommands(ctx context.Context, req *scenariov1.ListVoiceCommandsRequest) (*scenariov1.ListVoiceCommandsResponse, error) {
	commands, err := i.service.ListVoiceCommands(ctx, req.GetEdgeId(), req.GetRoomId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list voice commands: %v", err)
	}

	response := &scenariov1.ListVoiceCommandsResponse{
		Commands: make([]*scenariov1.VoiceCommand, 0, len(commands)),
	}
	for _, command := range commands {
		response.Commands = append(response.Commands, command.ToProto())
	}
	return response, nil
}

func (i *Implementation) ExecuteVoiceCommand(ctx context.Context, req *scenariov1.ExecuteVoiceCommandRequest) (*scenariov1.ExecuteVoiceCommandResponse, error) {
	scenario, decision, err := i.service.ExecuteVoiceCommand(
		ctx,
		req.GetEdgeId(),
		req.GetRoomId(),
		req.GetCommandName(),
		req.GetSource(),
		req.GetDeferExecution(),
	)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.InvalidArgument, "execute voice command: %v", err)
	}

	return &scenariov1.ExecuteVoiceCommandResponse{
		ScenarioId:   scenario.ID,
		ScenarioName: scenario.Name,
		Decision:     decision.ToProto(),
		Status:       decision.Status,
	}, nil
}

func errorsIsNotFound(err error) bool {
	return err == service.ErrNotFound
}
