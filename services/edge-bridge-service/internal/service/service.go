package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	devicev1 "device-service/pkg/pb/device/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

type Service struct {
	device   DeviceClient
	scenario ScenarioClient

	mu         sync.Mutex
	edges      map[string]EdgeStatus
	queues     map[string][]Command
	lastEvents map[string]EventSnapshot
}

type EdgeRegistration struct {
	EdgeID     string `json:"edge_id"`
	Name       string `json:"name"`
	PublicAddr string `json:"public_addr"`
}

type Room struct {
	RoomID string `json:"room_id"`
	Name   string `json:"name"`
	Floor  string `json:"floor,omitempty"`
}

type Device struct {
	DeviceID       string    `json:"device_id"`
	RoomID         string    `json:"room_id,omitempty"`
	EntityID       string    `json:"entity_id,omitempty"`
	Name           string    `json:"name"`
	DeviceType     string    `json:"device_type"`
	State          string    `json:"state"`
	OfflineCapable bool      `json:"offline_capable"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	Integration    string    `json:"integration,omitempty"`
	ExternalID     string    `json:"external_id,omitempty"`
	LastChangedAt  time.Time `json:"last_changed_at,omitempty"`
}

type InventorySync struct {
	EdgeID  string   `json:"edge_id"`
	Rooms   []Room   `json:"rooms"`
	Devices []Device `json:"devices"`
}

type Event struct {
	EventID    string                 `json:"event_id"`
	EdgeID     string                 `json:"edge_id"`
	RoomID     string                 `json:"room_id,omitempty"`
	DeviceID   string                 `json:"device_id,omitempty"`
	EntityID   string                 `json:"entity_id,omitempty"`
	EventType  string                 `json:"event_type"`
	State      string                 `json:"state,omitempty"`
	OccurredAt time.Time              `json:"occurred_at"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
}

type Command struct {
	CommandID   string    `json:"command_id"`
	DeviceID    string    `json:"device_id,omitempty"`
	EntityID    string    `json:"entity_id,omitempty"`
	TargetState string    `json:"target_state"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
}

type EdgeStatus struct {
	EdgeID            string    `json:"edge_id"`
	Name              string    `json:"name"`
	PublicAddr        string    `json:"public_addr,omitempty"`
	RegisteredAt      time.Time `json:"registered_at"`
	LastSeenAt        time.Time `json:"last_seen_at"`
	LastInventorySync time.Time `json:"last_inventory_sync,omitempty"`
	LastEventAt       time.Time `json:"last_event_at,omitempty"`
	LastPollAt        time.Time `json:"last_poll_at,omitempty"`
	PendingCommands   int       `json:"pending_commands"`
	LastError         string    `json:"last_error,omitempty"`
}

type EventResult struct {
	Status             string   `json:"status"`
	MatchedScenarioIDs []string `json:"matched_scenario_ids,omitempty"`
	QueuedCommands     int      `json:"queued_commands"`
}

func New(device DeviceClient, scenario ScenarioClient) *Service {
	return &Service{
		device:     device,
		scenario:   scenario,
		edges:      make(map[string]EdgeStatus),
		queues:     make(map[string][]Command),
		lastEvents: make(map[string]EventSnapshot),
	}
}

type EventSnapshot struct {
	EventType  string    `json:"event_type"`
	State      string    `json:"state,omitempty"`
	EntityID   string    `json:"entity_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (s *Service) RegisterEdge(ctx context.Context, req EdgeRegistration) (EdgeStatus, error) {
	if req.EdgeID == "" {
		return EdgeStatus{}, fmt.Errorf("edge_id is required")
	}

	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.edges[req.EdgeID]
	if current.RegisteredAt.IsZero() {
		current.RegisteredAt = now
	}
	current.EdgeID = req.EdgeID
	current.Name = req.Name
	current.PublicAddr = req.PublicAddr
	current.LastSeenAt = now
	current.PendingCommands = len(s.queues[req.EdgeID])
	current.LastError = ""
	s.edges[req.EdgeID] = current
	return current, nil
}

func (s *Service) SyncInventory(ctx context.Context, req InventorySync) (string, error) {
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

	_, err := s.device.SyncInventory(ctx, &devicev1.SyncInventoryRequest{
		EdgeId:  req.EdgeID,
		Rooms:   rooms,
		Devices: devices,
	})
	if err != nil {
		s.setEdgeError(req.EdgeID, err)
		return "", fmt.Errorf("sync inventory: %w", err)
	}

	s.mu.Lock()
	status := s.edges[req.EdgeID]
	status.EdgeID = req.EdgeID
	status.LastSeenAt = time.Now().UTC()
	status.LastInventorySync = time.Now().UTC()
	status.PendingCommands = len(s.queues[req.EdgeID])
	status.LastError = ""
	s.edges[req.EdgeID] = status
	s.mu.Unlock()

	return newID("sync"), nil
}

func (s *Service) PublishEvent(ctx context.Context, event Event) (EventResult, error) {
	if event.EdgeID == "" {
		return EventResult{}, fmt.Errorf("edge_id is required")
	}
	if event.EventType == "" {
		return EventResult{}, fmt.Errorf("event_type is required")
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	if event.DeviceID != "" || event.EntityID != "" {
		_, err := s.device.UpsertDeviceState(ctx, &devicev1.UpsertDeviceStateRequest{
			EdgeId: event.EdgeID,
			State: &devicev1.DeviceState{
				DeviceId:  event.DeviceID,
				EntityId:  event.EntityID,
				State:     event.State,
				ChangedAt: timestamppb.New(event.OccurredAt.UTC()),
			},
		})
		if err != nil {
			s.setEdgeError(event.EdgeID, err)
			return EventResult{}, fmt.Errorf("upsert device state: %w", err)
		}
	}

	response, err := s.scenario.EvaluateEvent(ctx, &scenariov1.EvaluateEventRequest{
		Event: &scenariov1.EventEnvelope{
			EventId:    event.EventID,
			EdgeId:     event.EdgeID,
			RoomId:     event.RoomID,
			EntityId:   event.EntityID,
			EventType:  event.EventType,
			State:      event.State,
			OccurredAt: timestamppb.New(event.OccurredAt.UTC()),
		},
	})
	if err != nil {
		s.setEdgeError(event.EdgeID, err)
		return EventResult{}, fmt.Errorf("evaluate event: %w", err)
	}

	decision := response.GetDecision()
	queued := s.enqueueDecisionCommands(event.EdgeID, decision, "scenario:"+decision.GetDecisionId())

	s.mu.Lock()
	status := s.edges[event.EdgeID]
	status.EdgeID = event.EdgeID
	status.LastSeenAt = time.Now().UTC()
	status.LastEventAt = event.OccurredAt.UTC()
	status.PendingCommands = len(s.queues[event.EdgeID])
	status.LastError = ""
	s.edges[event.EdgeID] = status
	s.lastEvents[event.EdgeID] = EventSnapshot{
		EventType:  event.EventType,
		State:      event.State,
		EntityID:   event.EntityID,
		OccurredAt: event.OccurredAt.UTC(),
	}
	s.mu.Unlock()

	return EventResult{
		Status:             decision.GetStatus(),
		MatchedScenarioIDs: append([]string(nil), decision.GetMatchedScenarioIds()...),
		QueuedCommands:     queued,
	}, nil
}

func (s *Service) PollCommands(_ context.Context, edgeID string) ([]Command, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	commands := append([]Command(nil), s.queues[edgeID]...)
	delete(s.queues, edgeID)

	status := s.edges[edgeID]
	status.EdgeID = edgeID
	status.LastSeenAt = time.Now().UTC()
	status.LastPollAt = time.Now().UTC()
	status.PendingCommands = 0
	s.edges[edgeID] = status

	return commands, nil
}

func (s *Service) GetOfflineScenarios(ctx context.Context, edgeID string) ([]*scenariov1.Scenario, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("edge_id is required")
	}
	response, err := s.scenario.GetOfflineScenarios(ctx, &scenariov1.GetOfflineScenariosRequest{EdgeId: edgeID})
	if err != nil {
		s.setEdgeError(edgeID, err)
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
		s.setEdgeError(edgeID, err)
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
		s.setEdgeError(edgeID, err)
		return nil, fmt.Errorf("list voice commands: %w", err)
	}
	return response.GetCommands(), nil
}

func (s *Service) GetEdgeStatus(edgeID string) (EdgeStatus, EventSnapshot, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status, ok := s.edges[edgeID]
	if !ok {
		return EdgeStatus{}, EventSnapshot{}, false
	}
	status.PendingCommands = len(s.queues[edgeID])
	return status, s.lastEvents[edgeID], true
}

func (s *Service) enqueueDecisionCommands(edgeID string, decision *scenariov1.Decision, defaultSource string) int {
	if decision == nil {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, action := range decision.GetActions() {
		actionType := action.GetActionType()
		if actionType != "" && actionType != "device_command" {
			continue
		}
		if action.GetTargetState() == "" {
			continue
		}
		s.queues[edgeID] = append(s.queues[edgeID], Command{
			CommandID:   newID("cmd"),
			DeviceID:    action.GetDeviceId(),
			EntityID:    action.GetEntityId(),
			TargetState: action.GetTargetState(),
			Source:      defaultSource,
			CreatedAt:   time.Now().UTC(),
		})
		count++
	}

	status := s.edges[edgeID]
	status.EdgeID = edgeID
	status.PendingCommands = len(s.queues[edgeID])
	s.edges[edgeID] = status

	return count
}

func (s *Service) setEdgeError(edgeID string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status := s.edges[edgeID]
	status.EdgeID = edgeID
	status.LastSeenAt = time.Now().UTC()
	status.LastError = err.Error()
	status.PendingCommands = len(s.queues[edgeID])
	s.edges[edgeID] = status
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
