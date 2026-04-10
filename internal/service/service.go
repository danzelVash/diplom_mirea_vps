package service

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/danzelVash/diplom_mirea_vps/internal/model"
	"github.com/danzelVash/diplom_mirea_vps/internal/store"
)

var ErrNotFound = errors.New("not found")

type Service struct {
	store *store.Store
}

func New(st *store.Store) *Service {
	return &Service{store: st}
}

func (s *Service) RegisterEdge(id, name, publicAddr string) (model.EdgeNode, error) {
	now := time.Now()
	var registered model.EdgeNode

	err := s.store.Save(func(snapshot *model.Snapshot) error {
		idx := slices.IndexFunc(snapshot.Edges, func(edge model.EdgeNode) bool { return edge.ID == id })
		if idx >= 0 {
			snapshot.Edges[idx].Name = choose(name, snapshot.Edges[idx].Name)
			snapshot.Edges[idx].PublicAddr = publicAddr
			snapshot.Edges[idx].LastSeenAt = now
			snapshot.Edges[idx].ConnectionState = "online"
			registered = snapshot.Edges[idx]
			return nil
		}

		registered = model.EdgeNode{
			ID:              id,
			Name:            choose(name, id),
			PublicAddr:      publicAddr,
			LastSeenAt:      now,
			LastSyncAt:      now,
			ConnectionState: "online",
		}
		snapshot.Edges = append(snapshot.Edges, registered)
		return nil
	})

	return registered, err
}

func (s *Service) UpsertRooms(rooms []model.Room) error {
	return s.store.Save(func(snapshot *model.Snapshot) error {
		for _, room := range rooms {
			idx := slices.IndexFunc(snapshot.Rooms, func(item model.Room) bool { return item.ID == room.ID })
			if idx >= 0 {
				snapshot.Rooms[idx] = room
				continue
			}
			snapshot.Rooms = append(snapshot.Rooms, room)
		}
		return nil
	})
}

func (s *Service) UpsertDevices(edgeID string, devices []model.Device) error {
	updatedAt := time.Now()
	return s.store.Save(func(snapshot *model.Snapshot) error {
		for _, device := range devices {
			device.EdgeID = edgeID
			if device.UpdatedAt.IsZero() {
				device.UpdatedAt = updatedAt
			}
			if device.LastChanged.IsZero() {
				device.LastChanged = updatedAt
			}
			idx := slices.IndexFunc(snapshot.Devices, func(item model.Device) bool { return item.ID == device.ID })
			if idx >= 0 {
				snapshot.Devices[idx] = device
				continue
			}
			snapshot.Devices = append(snapshot.Devices, device)
		}
		edgeIdx := slices.IndexFunc(snapshot.Edges, func(edge model.EdgeNode) bool { return edge.ID == edgeID })
		if edgeIdx >= 0 {
			snapshot.Edges[edgeIdx].LastSyncAt = updatedAt
			snapshot.Edges[edgeIdx].LastSeenAt = updatedAt
			snapshot.Edges[edgeIdx].ConnectionState = "online"
		}
		return nil
	})
}

func (s *Service) CreateScenario(scenario model.Scenario) (model.Scenario, error) {
	if scenario.ID == "" {
		scenario.ID = newID("scn")
	}
	if scenario.CreatedAt.IsZero() {
		scenario.CreatedAt = time.Now()
	}
	return scenario, s.store.Save(func(snapshot *model.Snapshot) error {
		snapshot.Scenarios = append(snapshot.Scenarios, scenario)
		return nil
	})
}

func (s *Service) ListScenarios(edgeID string, offlineOnly bool) []model.Scenario {
	snapshot := s.store.Snapshot()
	var scenarios []model.Scenario
	for _, scenario := range snapshot.Scenarios {
		if edgeID != "" && scenario.EdgeID != "" && scenario.EdgeID != edgeID {
			continue
		}
		if offlineOnly && !scenario.OfflineEligible {
			continue
		}
		scenarios = append(scenarios, scenario)
	}
	slices.SortFunc(scenarios, func(a, b model.Scenario) int { return b.Priority - a.Priority })
	return scenarios
}

func (s *Service) ListRooms() []model.Room {
	return s.store.Snapshot().Rooms
}

func (s *Service) ListDevices(edgeID string) []model.Device {
	snapshot := s.store.Snapshot()
	var devices []model.Device
	for _, device := range snapshot.Devices {
		if edgeID != "" && device.EdgeID != edgeID {
			continue
		}
		devices = append(devices, device)
	}
	return devices
}

func (s *Service) IngestEvent(event model.Event) (model.Event, []model.Command, error) {
	if event.ID == "" {
		event.ID = newID("evt")
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	var generated []model.Command
	err := s.store.Save(func(snapshot *model.Snapshot) error {
		snapshot.Events = append(snapshot.Events, event)

		for i := range snapshot.Devices {
			if snapshot.Devices[i].ID == event.DeviceID {
				if event.Status != "" && snapshot.Devices[i].Status != event.Status {
					snapshot.Devices[i].Status = event.Status
					snapshot.Devices[i].LastChanged = event.OccurredAt
				}
				snapshot.Devices[i].UpdatedAt = time.Now()
			}
		}

		for _, scenario := range snapshot.Scenarios {
			if !scenario.Enabled {
				continue
			}
			if scenario.EdgeID != "" && scenario.EdgeID != event.EdgeID {
				continue
			}
			if !matchesScenario(scenario, event) {
				continue
			}
			for _, action := range scenario.Actions {
				switch action.Type {
				case "device_command", "device_status":
					command := model.Command{
						ID:           newID("cmd"),
						EdgeID:       event.EdgeID,
						ScenarioID:   scenario.ID,
						DeviceID:     choose(action.DeviceID, event.DeviceID),
						EntityID:     choose(action.EntityID, event.EntityID),
						TargetStatus: action.TargetStatus,
						Source:       fmt.Sprintf("scenario:%s", scenario.Name),
						CreatedAt:    time.Now(),
					}
					snapshot.Commands[event.EdgeID] = append(snapshot.Commands[event.EdgeID], command)
					generated = append(generated, command)
				}
			}
		}

		return nil
	})

	return event, generated, err
}

func (s *Service) PollCommands(edgeID string) ([]model.Command, error) {
	var commands []model.Command
	err := s.store.Save(func(snapshot *model.Snapshot) error {
		commands = append(commands, snapshot.Commands[edgeID]...)
		delete(snapshot.Commands, edgeID)
		return nil
	})
	return commands, err
}

func (s *Service) CreateRoom(room model.Room) (model.Room, error) {
	if room.ID == "" {
		room.ID = newID("room")
	}
	if room.CreatedAt.IsZero() {
		room.CreatedAt = time.Now()
	}
	return room, s.UpsertRooms([]model.Room{room})
}

func (s *Service) CreateDevice(device model.Device) (model.Device, error) {
	if device.ID == "" {
		device.ID = newID("dev")
	}
	if device.UpdatedAt.IsZero() {
		device.UpdatedAt = time.Now()
	}
	if device.LastChanged.IsZero() {
		device.LastChanged = device.UpdatedAt
	}
	return device, s.UpsertDevices(device.EdgeID, []model.Device{device})
}

func matchesScenario(scenario model.Scenario, event model.Event) bool {
	for _, trigger := range scenario.Triggers {
		if trigger.Type != "event" && trigger.Type != "state" {
			continue
		}
		if trigger.Event != "" && trigger.Event != event.Type {
			continue
		}
		if trigger.DeviceID != "" && trigger.DeviceID != event.DeviceID {
			continue
		}
		if trigger.EntityID != "" && trigger.EntityID != event.EntityID {
			continue
		}
		if trigger.Status != "" && trigger.Status != event.Status {
			continue
		}
		return true
	}
	return false
}

func choose(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
