package v1

import (
	"context"

	devicev1 "device-service/pkg/pb/device/v1"
	bridgeservice "edge-bridge-service/internal/service"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
)

type Implementation struct {
	edgebridgev1.UnimplementedEdgeBridgeServiceServer
	service  *bridgeservice.Service
	device   externalDeviceClient
	scenario externalScenarioClient
}

type externalDeviceClient interface {
	devicev1.DeviceServiceClient
}

type externalScenarioClient interface {
	scenariov1.ScenarioServiceClient
}

func New(service *bridgeservice.Service, device externalDeviceClient, scenario externalScenarioClient) *Implementation {
	return &Implementation{
		service:  service,
		device:   device,
		scenario: scenario,
	}
}

func (i *Implementation) RegisterEdge(ctx context.Context, req *edgebridgev1.RegisterEdgeRequest) (*edgebridgev1.RegisterEdgeResponse, error) {
	_, err := i.service.RegisterEdge(ctx, bridgeservice.EdgeRegistration{
		EdgeID:     req.GetEdge().GetEdgeId(),
		Name:       req.GetEdge().GetName(),
		PublicAddr: req.GetEdge().GetPublicAddr(),
	})
	if err != nil {
		return nil, err
	}
	return &edgebridgev1.RegisterEdgeResponse{Status: "registered"}, nil
}

func (i *Implementation) SyncInventory(ctx context.Context, req *edgebridgev1.SyncInventoryRequest) (*edgebridgev1.SyncInventoryResponse, error) {
	rooms := make([]bridgeservice.Room, 0, len(req.GetRooms()))
	for _, room := range req.GetRooms() {
		rooms = append(rooms, bridgeservice.Room{
			RoomID: room.GetRoomId(),
			Name:   room.GetName(),
		})
	}

	devices := make([]bridgeservice.Device, 0, len(req.GetDevices()))
	for _, device := range req.GetDevices() {
		devices = append(devices, bridgeservice.Device{
			DeviceID:   device.GetDeviceId(),
			RoomID:     device.GetRoomId(),
			EntityID:   device.GetEntityId(),
			Name:       device.GetName(),
			DeviceType: device.GetDeviceType(),
			State:      device.GetState(),
		})
	}

	syncID, err := i.service.SyncInventory(ctx, bridgeservice.InventorySync{
		EdgeID:  req.GetEdgeId(),
		Rooms:   rooms,
		Devices: devices,
	})
	if err != nil {
		return nil, err
	}
	return &edgebridgev1.SyncInventoryResponse{SyncId: syncID, Status: "synced"}, nil
}

func (i *Implementation) PublishEvent(ctx context.Context, req *edgebridgev1.PublishEventRequest) (*edgebridgev1.PublishEventResponse, error) {
	_, err := i.service.PublishEvent(ctx, bridgeservice.Event{
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
		return nil, err
	}
	return &edgebridgev1.PublishEventResponse{Status: "accepted"}, nil
}

func (i *Implementation) PollCommands(ctx context.Context, req *edgebridgev1.PollCommandsRequest) (*edgebridgev1.PollCommandsResponse, error) {
	commands, err := i.service.PollCommands(ctx, req.GetEdgeId())
	if err != nil {
		return nil, err
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
		return nil, err
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
