package v1

import (
	"context"

	"edge-bridge-service/internal/model"
	bridgeservice "edge-bridge-service/internal/service"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Implementation struct {
	edgebridgev1.UnimplementedEdgeBridgeServiceServer
	service *bridgeservice.Service
}

func New(service *bridgeservice.Service) *Implementation {
	return &Implementation{service: service}
}

func (i *Implementation) RegisterEdge(ctx context.Context, req *edgebridgev1.RegisterEdgeRequest) (*edgebridgev1.RegisterEdgeResponse, error) {
	_, err := i.service.RegisterEdge(ctx, model.EdgeRegistration{
		EdgeID:     req.GetEdge().GetEdgeId(),
		Name:       req.GetEdge().GetName(),
		PublicAddr: req.GetEdge().GetPublicAddr(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "register edge: %v", err)
	}
	return &edgebridgev1.RegisterEdgeResponse{Status: "registered"}, nil
}

func (i *Implementation) SyncInventory(ctx context.Context, req *edgebridgev1.SyncInventoryRequest) (*edgebridgev1.SyncInventoryResponse, error) {
	rooms := make([]model.Room, 0, len(req.GetRooms()))
	for _, room := range req.GetRooms() {
		rooms = append(rooms, model.Room{
			RoomID: room.GetRoomId(),
			Name:   room.GetName(),
		})
	}

	devices := make([]model.Device, 0, len(req.GetDevices()))
	for _, device := range req.GetDevices() {
		devices = append(devices, model.Device{
			DeviceID:   device.GetDeviceId(),
			RoomID:     device.GetRoomId(),
			EntityID:   device.GetEntityId(),
			Name:       device.GetName(),
			DeviceType: device.GetDeviceType(),
			State:      device.GetState(),
		})
	}

	syncID, err := i.service.SyncInventory(ctx, model.InventorySync{
		EdgeID:  req.GetEdgeId(),
		Rooms:   rooms,
		Devices: devices,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "sync inventory: %v", err)
	}
	return &edgebridgev1.SyncInventoryResponse{SyncId: syncID, Status: "synced"}, nil
}

func (i *Implementation) PublishEvent(ctx context.Context, req *edgebridgev1.PublishEventRequest) (*edgebridgev1.PublishEventResponse, error) {
	result, err := i.service.PublishEvent(ctx, model.Event{
		EventID:    req.GetEvent().GetEventId(),
		EdgeID:     req.GetEvent().GetEdgeId(),
		RoomID:     req.GetEvent().GetRoomId(),
		DeviceID:   req.GetEvent().GetDeviceId(),
		EntityID:   req.GetEvent().GetEntityId(),
		EventType:  req.GetEvent().GetEventType(),
		State:      req.GetEvent().GetState(),
		OccurredAt: req.GetEvent().GetOccurredAt().AsTime(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "publish event: %v", err)
	}
	return &edgebridgev1.PublishEventResponse{Status: result.Status}, nil
}

func (i *Implementation) PollCommands(ctx context.Context, req *edgebridgev1.PollCommandsRequest) (*edgebridgev1.PollCommandsResponse, error) {
	commands, err := i.service.PollCommands(ctx, req.GetEdgeId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "poll commands: %v", err)
	}

	response := &edgebridgev1.PollCommandsResponse{
		Commands: make([]*edgebridgev1.EdgeCommand, 0, len(commands)),
	}
	for _, command := range commands {
		response.Commands = append(response.Commands, &edgebridgev1.EdgeCommand{
			CommandId:   command.CommandID,
			DeviceId:    command.DeviceID,
			EntityId:    command.EntityID,
			TargetState: command.TargetState,
			Source:      command.Source,
		})
	}
	return response, nil
}

func (i *Implementation) GetOfflineScenarios(ctx context.Context, req *edgebridgev1.GetOfflineScenariosRequest) (*edgebridgev1.GetOfflineScenariosResponse, error) {
	scenarios, err := i.service.GetOfflineScenarios(ctx, req.GetEdgeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get offline scenarios: %v", err)
	}

	response := &edgebridgev1.GetOfflineScenariosResponse{
		Scenarios: make([]*edgebridgev1.OfflineScenario, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		response.Scenarios = append(response.Scenarios, &edgebridgev1.OfflineScenario{
			ScenarioId: scenario.GetScenarioId(),
			Name:       scenario.GetName(),
			Priority:   scenario.GetPriority(),
		})
	}
	return response, nil
}

func (i *Implementation) ExecuteVoiceCommand(ctx context.Context, req *edgebridgev1.ExecuteVoiceCommandRequest) (*edgebridgev1.ExecuteVoiceCommandResponse, error) {
	command, err := i.service.ExecuteVoiceCommand(ctx, req.GetEdgeId(), req.GetRoomId(), req.GetAudio(), req.GetSource())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "execute voice command: %v", err)
	}

	return &edgebridgev1.ExecuteVoiceCommandResponse{
		CommandId:         command.GetCommandId(),
		RecognizedCommand: command.GetRecognizedCommand(),
		ScenarioId:        command.GetScenarioId(),
		ScenarioName:      command.GetScenarioName(),
		DeviceId:          command.GetDeviceId(),
		EntityId:          command.GetEntityId(),
		TargetState:       command.GetTargetState(),
		Status:            command.GetExecutionStatus(),
	}, nil
}
