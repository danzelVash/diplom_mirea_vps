package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	devicev1 "device-service/pkg/pb/device/v1"
	"edge-bridge-service/internal/model"
	edgebridgestore "edge-bridge-service/internal/store"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
	voicev1 "voice-service/pkg/pb/voice/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Store interface {
	RegisterEdge(ctx context.Context, req model.EdgeRegistration) (model.EdgeStatus, error)
	MarkInventorySynced(ctx context.Context, edgeID string) error
	RecordEvent(ctx context.Context, event model.Event) error
	EnqueueCommands(ctx context.Context, edgeID string, commands []model.Command) error
	PollCommands(ctx context.Context, edgeID string, retryAfter time.Duration) ([]model.Command, error)
	AckCommand(ctx context.Context, edgeID, commandID string) error
	SetEdgeError(ctx context.Context, edgeID, message string) error
	GetEdgeStatus(ctx context.Context, edgeID string) (model.EdgeStatus, error)
	GetLastEvent(ctx context.Context, edgeID string) (model.EventSnapshot, error)
}

type DeviceClient interface {
	SyncInventory(ctx context.Context, in *devicev1.SyncInventoryRequest, opts ...grpc.CallOption) (*devicev1.SyncInventoryResponse, error)
	UpsertDeviceState(ctx context.Context, in *devicev1.UpsertDeviceStateRequest, opts ...grpc.CallOption) (*devicev1.UpsertDeviceStateResponse, error)
}

type ScenarioClient interface {
	EvaluateEvent(ctx context.Context, in *scenariov1.EvaluateEventRequest, opts ...grpc.CallOption) (*scenariov1.EvaluateEventResponse, error)
	GetOfflineScenarios(ctx context.Context, in *scenariov1.GetOfflineScenariosRequest, opts ...grpc.CallOption) (*scenariov1.GetOfflineScenariosResponse, error)
	ListScenarios(ctx context.Context, in *scenariov1.ListScenariosRequest, opts ...grpc.CallOption) (*scenariov1.ListScenariosResponse, error)
	ListVoiceCommands(ctx context.Context, in *scenariov1.ListVoiceCommandsRequest, opts ...grpc.CallOption) (*scenariov1.ListVoiceCommandsResponse, error)
}

type VoiceClient interface {
	ParseVoiceCommand(ctx context.Context, in *voicev1.ParseVoiceCommandRequest, opts ...grpc.CallOption) (*voicev1.ParseVoiceCommandResponse, error)
}

type Service struct {
	store      Store
	device     DeviceClient
	scenario   ScenarioClient
	voice      VoiceClient
	retryAfter time.Duration
}

func New(store Store, device DeviceClient, scenario ScenarioClient, voice VoiceClient, retryAfter time.Duration) *Service {
	return &Service{
		store:      store,
		device:     device,
		scenario:   scenario,
		voice:      voice,
		retryAfter: retryAfter,
	}
}

func (s *Service) RegisterEdge(ctx context.Context, req model.EdgeRegistration) (model.EdgeStatus, error) {
	if req.EdgeID == "" {
		return model.EdgeStatus{}, fmt.Errorf("edge_id is required")
	}

	status, err := s.store.RegisterEdge(ctx, req)
	if err != nil {
		return model.EdgeStatus{}, err
	}
	return status, nil
}

func (s *Service) SyncInventory(ctx context.Context, req model.InventorySync) (string, error) {
	if req.EdgeID == "" {
		return "", fmt.Errorf("edge_id is required")
	}

	rooms := make([]*devicev1.Room, 0, len(req.Rooms))
	for _, room := range req.Rooms {
		rooms = append(rooms, &devicev1.Room{
			RoomId: room.RoomID,
			Name:   room.Name,
			Floor:  room.Floor,
		})
	}

	devices := make([]*devicev1.Device, 0, len(req.Devices))
	for _, device := range req.Devices {
		var updatedAt *timestamppb.Timestamp
		if !device.UpdatedAt.IsZero() {
			updatedAt = timestamppb.New(device.UpdatedAt.UTC())
		}
		devices = append(devices, &devicev1.Device{
			DeviceId:       device.DeviceID,
			EdgeId:         req.EdgeID,
			RoomId:         device.RoomID,
			Name:           device.Name,
			DeviceType:     device.DeviceType,
			EntityId:       device.EntityID,
			State:          device.State,
			OfflineCapable: device.OfflineCapable,
			UpdatedAt:      updatedAt,
		})
	}

	if _, err := s.device.SyncInventory(ctx, &devicev1.SyncInventoryRequest{
		EdgeId:  req.EdgeID,
		Rooms:   rooms,
		Devices: devices,
	}); err != nil {
		_ = s.store.SetEdgeError(ctx, req.EdgeID, err.Error())
		return "", fmt.Errorf("sync inventory: %w", err)
	}

	if err := s.store.MarkInventorySynced(ctx, req.EdgeID); err != nil {
		return "", err
	}
	return newID("sync"), nil
}

func (s *Service) PublishEvent(ctx context.Context, event model.Event) (model.EventResult, error) {
	if event.EdgeID == "" {
		return model.EventResult{}, fmt.Errorf("edge_id is required")
	}
	if event.EventType == "" {
		return model.EventResult{}, fmt.Errorf("event_type is required")
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	if event.DeviceID != "" || event.EntityID != "" {
		if _, err := s.device.UpsertDeviceState(ctx, &devicev1.UpsertDeviceStateRequest{
			EdgeId: event.EdgeID,
			State: &devicev1.DeviceState{
				DeviceId:  event.DeviceID,
				EntityId:  event.EntityID,
				State:     event.State,
				ChangedAt: timestamppb.New(event.OccurredAt.UTC()),
			},
		}); err != nil {
			_ = s.store.SetEdgeError(ctx, event.EdgeID, err.Error())
			return model.EventResult{}, fmt.Errorf("upsert device state: %w", err)
		}
	}

	response, err := s.scenario.EvaluateEvent(ctx, &scenariov1.EvaluateEventRequest{
		Event: &scenariov1.EventEnvelope{
			EventId:    event.EventID,
			EdgeId:     event.EdgeID,
			RoomId:     event.RoomID,
			DeviceId:   event.DeviceID,
			EntityId:   event.EntityID,
			EventType:  event.EventType,
			State:      event.State,
			OccurredAt: timestamppb.New(event.OccurredAt.UTC()),
		},
		DeferExecution: true,
	})
	if err != nil {
		_ = s.store.SetEdgeError(ctx, event.EdgeID, err.Error())
		return model.EventResult{}, fmt.Errorf("evaluate event: %w", err)
	}

	decision := response.GetDecision()
	queuedCommands := commandsFromDecision(event.EdgeID, decision, "scenario:"+decision.GetDecisionId())
	if err := s.store.EnqueueCommands(ctx, event.EdgeID, queuedCommands); err != nil {
		return model.EventResult{}, err
	}
	if err := s.store.RecordEvent(ctx, event); err != nil {
		return model.EventResult{}, err
	}

	return model.EventResult{
		Status:             decision.GetStatus(),
		MatchedScenarioIDs: append([]string(nil), decision.GetMatchedScenarioIds()...),
		QueuedCommands:     len(queuedCommands),
	}, nil
}

func (s *Service) PollCommands(ctx context.Context, edgeID string) ([]model.Command, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	return s.store.PollCommands(ctx, edgeID, s.retryAfter)
}

func (s *Service) AckCommands(ctx context.Context, edgeID string, acks []model.CommandAck) error {
	if edgeID == "" {
		return fmt.Errorf("edge_id is required")
	}

	for _, ack := range acks {
		if ack.CommandID == "" {
			return fmt.Errorf("command_id is required")
		}
		if err := s.store.AckCommand(ctx, edgeID, ack.CommandID); err != nil {
			if err == edgebridgestore.ErrNotFound {
				return err
			}
			return err
		}

		if ack.State == "" || (ack.DeviceID == "" && ack.EntityID == "") {
			continue
		}
		if _, err := s.device.UpsertDeviceState(ctx, &devicev1.UpsertDeviceStateRequest{
			EdgeId: edgeID,
			State: &devicev1.DeviceState{
				DeviceId:  ack.DeviceID,
				EntityId:  ack.EntityID,
				State:     ack.State,
				ChangedAt: timestamppb.Now(),
			},
		}); err != nil {
			_ = s.store.SetEdgeError(ctx, edgeID, err.Error())
			return fmt.Errorf("ack command state update: %w", err)
		}
	}
	return nil
}

func (s *Service) GetOfflineScenarios(ctx context.Context, edgeID string) ([]*scenariov1.Scenario, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	response, err := s.scenario.GetOfflineScenarios(ctx, &scenariov1.GetOfflineScenariosRequest{EdgeId: edgeID})
	if err != nil {
		_ = s.store.SetEdgeError(ctx, edgeID, err.Error())
		return nil, fmt.Errorf("get offline scenarios: %w", err)
	}
	return response.GetScenarios(), nil
}

func (s *Service) ListScenarios(ctx context.Context, edgeID string) ([]*scenariov1.Scenario, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	response, err := s.scenario.ListScenarios(ctx, &scenariov1.ListScenariosRequest{EdgeId: edgeID})
	if err != nil {
		_ = s.store.SetEdgeError(ctx, edgeID, err.Error())
		return nil, fmt.Errorf("list scenarios: %w", err)
	}
	return response.GetScenarios(), nil
}

func (s *Service) ListVoiceCommands(ctx context.Context, edgeID, roomID string) ([]*scenariov1.VoiceCommand, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	response, err := s.scenario.ListVoiceCommands(ctx, &scenariov1.ListVoiceCommandsRequest{
		EdgeId: edgeID,
		RoomId: roomID,
	})
	if err != nil {
		_ = s.store.SetEdgeError(ctx, edgeID, err.Error())
		return nil, fmt.Errorf("list voice commands: %w", err)
	}
	return response.GetCommands(), nil
}

func (s *Service) ExecuteVoiceCommand(ctx context.Context, edgeID, roomID string, audio []byte, source string) (*voicev1.ParsedVoiceCommand, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	if len(audio) == 0 {
		return nil, fmt.Errorf("audio is required")
	}
	if source == "" {
		source = "edge-bridge-service"
	}

	response, err := s.voice.ParseVoiceCommand(ctx, &voicev1.ParseVoiceCommandRequest{
		Audio:          audio,
		EdgeId:         edgeID,
		RoomId:         roomID,
		Source:         source,
		DeferExecution: true,
	})
	if err != nil {
		_ = s.store.SetEdgeError(ctx, edgeID, err.Error())
		return nil, fmt.Errorf("parse voice command: %w", err)
	}

	command := response.GetCommand()
	if command == nil {
		return &voicev1.ParsedVoiceCommand{}, nil
	}
	if command.GetTargetState() != "" && (command.GetDeviceId() != "" || command.GetEntityId() != "") {
		queueCommand := model.Command{
			CommandID:   newID("cmd"),
			EdgeID:      edgeID,
			DeviceID:    command.GetDeviceId(),
			EntityID:    command.GetEntityId(),
			TargetState: command.GetTargetState(),
			Source:      "voice:" + source,
			CreatedAt:   time.Now().UTC(),
		}
		if err := s.store.EnqueueCommands(ctx, edgeID, []model.Command{queueCommand}); err != nil {
			return nil, err
		}
		command.CommandId = queueCommand.CommandID
		command.ExecutionStatus = "queued"
	}
	return command, nil
}

func (s *Service) GetEdgeStatus(ctx context.Context, edgeID string) (model.EdgeStatus, model.EventSnapshot, bool, error) {
	status, err := s.store.GetEdgeStatus(ctx, edgeID)
	if err != nil {
		if err == edgebridgestore.ErrNotFound {
			return model.EdgeStatus{}, model.EventSnapshot{}, false, nil
		}
		return model.EdgeStatus{}, model.EventSnapshot{}, false, err
	}

	lastEvent, err := s.store.GetLastEvent(ctx, edgeID)
	if err != nil && err != edgebridgestore.ErrNotFound {
		return model.EdgeStatus{}, model.EventSnapshot{}, false, err
	}
	return status, lastEvent, true, nil
}

func commandsFromDecision(edgeID string, decision *scenariov1.Decision, defaultSource string) []model.Command {
	if decision == nil {
		return nil
	}

	result := make([]model.Command, 0)
	for _, action := range decision.GetActions() {
		if actionType := action.GetActionType(); actionType != "" && actionType != "device_command" {
			continue
		}
		if action.GetTargetState() == "" {
			continue
		}
		if action.GetDeviceId() == "" && action.GetEntityId() == "" {
			continue
		}
		result = append(result, model.Command{
			CommandID:   newID("cmd"),
			EdgeID:      edgeID,
			DeviceID:    action.GetDeviceId(),
			EntityID:    action.GetEntityId(),
			TargetState: action.GetTargetState(),
			Source:      defaultSource,
			CreatedAt:   time.Now().UTC(),
		})
	}
	return result
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), rand.Intn(1000000))
}
