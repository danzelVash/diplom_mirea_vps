package v1

import (
	"context"

	gatewayv1 "platform-service/pkg/pb/api_gateway/v1"
	devicev1 "platform-service/pkg/pb/device/v1"
	scenariov1 "platform-service/pkg/pb/scenario/v1"
	voicev1 "platform-service/pkg/pb/voice/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Implementation struct {
	gatewayv1.UnimplementedGatewayServiceServer
	device   externalDeviceClient
	scenario externalScenarioClient
	voice    externalVoiceClient
}

type externalDeviceClient interface {
	ExecuteCommand(ctx context.Context, in *devicev1.ExecuteCommandRequest, opts ...grpc.CallOption) (*devicev1.ExecuteCommandResponse, error)
}

type externalScenarioClient interface {
	SaveScenario(ctx context.Context, in *scenariov1.SaveScenarioRequest, opts ...grpc.CallOption) (*scenariov1.SaveScenarioResponse, error)
}

type externalVoiceClient interface {
	ParseVoiceCommand(ctx context.Context, in *voicev1.ParseVoiceCommandRequest, opts ...grpc.CallOption) (*voicev1.ParseVoiceCommandResponse, error)
}

func New(
	device externalDeviceClient,
	scenario externalScenarioClient,
	voice externalVoiceClient,
) *Implementation {
	return &Implementation{
		device:   device,
		scenario: scenario,
		voice:    voice,
	}
}

func (i *Implementation) SaveScenario(ctx context.Context, req *gatewayv1.SaveScenarioRequest) (*gatewayv1.SaveScenarioResponse, error) {
	if req.GetEdgeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "edge_id is required")
	}
	if req.GetVoiceCommand() == "" {
		return nil, status.Error(codes.InvalidArgument, "voice_command is required")
	}
	if req.GetTargetState() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_state is required")
	}
	if req.GetDeviceId() == "" && req.GetEntityId() == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id or entity_id is required")
	}

	enabled := req.GetEnabled()
	if req.GetScenarioId() == "" && !enabled {
		enabled = true
	}

	scenario := &scenariov1.Scenario{
		ScenarioId:      req.GetScenarioId(),
		EdgeId:          req.GetEdgeId(),
		Name:            req.GetName(),
		Enabled:         enabled,
		Priority:        req.GetPriority(),
		OfflineEligible: req.GetOfflineEligible(),
		Triggers: []*scenariov1.Trigger{
			{
				TriggerType: "voice_command",
				CommandName: req.GetVoiceCommand(),
			},
		},
		Actions: []*scenariov1.Action{
			{
				ActionType:  "device_command",
				DeviceId:    req.GetDeviceId(),
				EntityId:    req.GetEntityId(),
				TargetState: req.GetTargetState(),
			},
		},
	}
	if scenario.Name == "" {
		scenario.Name = req.GetVoiceCommand()
	}
	if req.GetRoomId() != "" {
		scenario.Conditions = []*scenariov1.Condition{
			{
				ConditionType: "equals",
				Field:         "room_id",
				ExpectedValue: req.GetRoomId(),
			},
		}
	}

	response, err := i.scenario.SaveScenario(ctx, &scenariov1.SaveScenarioRequest{Scenario: scenario})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "save scenario: %v", err)
	}

	return &gatewayv1.SaveScenarioResponse{
		ScenarioId:   response.GetScenario().GetScenarioId(),
		Status:       "saved",
		VoiceCommand: req.GetVoiceCommand(),
	}, nil
}

func (i *Implementation) ExecuteDeviceCommand(ctx context.Context, req *gatewayv1.ExecuteDeviceCommandRequest) (*gatewayv1.ExecuteDeviceCommandResponse, error) {
	if req.GetTargetState() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_state is required")
	}
	if req.GetDeviceId() == "" && req.GetEntityId() == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id or entity_id is required")
	}

	response, err := i.device.ExecuteCommand(ctx, &devicev1.ExecuteCommandRequest{
		DeviceId:    req.GetDeviceId(),
		EntityId:    req.GetEntityId(),
		TargetState: req.GetTargetState(),
		Source:      req.GetSource(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "execute device command: %v", err)
	}

	return &gatewayv1.ExecuteDeviceCommandResponse{
		CommandId: response.GetCommandId(),
		Status:    response.GetStatus(),
	}, nil
}

func (i *Implementation) ParseVoiceCommand(ctx context.Context, req *gatewayv1.ParseVoiceCommandRequest) (*gatewayv1.ParseVoiceCommandResponse, error) {
	source := req.GetSource()
	if source == "" {
		source = "api-gateway"
	}

	response, err := i.voice.ParseVoiceCommand(ctx, &voicev1.ParseVoiceCommandRequest{
		Audio:  req.GetAudio(),
		RoomId: req.GetRoomId(),
		EdgeId: req.GetEdgeId(),
		Source: source,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "parse voice command: %v", err)
	}

	command := response.GetCommand()
	return &gatewayv1.ParseVoiceCommandResponse{
		CommandId:         command.GetCommandId(),
		Intent:            command.GetIntent(),
		Status:            command.GetExecutionStatus(),
		RecognizedCommand: command.GetRecognizedCommand(),
		ScenarioId:        command.GetScenarioId(),
		ScenarioName:      command.GetScenarioName(),
		DeviceId:          command.GetDeviceId(),
		EntityId:          command.GetEntityId(),
		TargetState:       command.GetTargetState(),
	}, nil
}
