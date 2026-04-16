package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"platform-service/internal/device/model"
	"platform-service/internal/device/store"
	devicev1 "platform-service/pkg/pb/device/v1"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	ListRooms(ctx context.Context, edgeID string) ([]model.Room, error)
	GetRoom(ctx context.Context, id string) (model.Room, error)
	UpsertRoom(ctx context.Context, room model.Room) (model.Room, error)
	DeleteRoom(ctx context.Context, id string) error
	ListDevices(ctx context.Context, edgeID, roomID string) ([]model.Device, error)
	GetDevice(ctx context.Context, id string) (model.Device, error)
	UpsertDevice(ctx context.Context, device model.Device) (model.Device, error)
	DeleteDevice(ctx context.Context, id string) error
	SyncInventory(ctx context.Context, edgeID string, rooms []model.Room, devices []model.Device) error
	UpdateDeviceState(ctx context.Context, edgeID, deviceID, entityID, state string, changedAt time.Time) error
	ExecuteCommand(ctx context.Context, command model.Command) (string, error)
}

type ListDevicesFilter struct {
	EdgeID string
	RoomID string
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListRooms(ctx context.Context, edgeID string) ([]model.Room, error) {
	return s.store.ListRooms(ctx, edgeID)
}

func (s *Service) GetRoom(ctx context.Context, id string) (model.Room, error) {
	room, err := s.store.GetRoom(ctx, id)
	return room, mapStoreErr(err)
}

func (s *Service) UpsertRoom(ctx context.Context, room model.Room) (model.Room, error) {
	if room.ID == "" {
		room.ID = newID("room")
	}
	if room.Name == "" {
		return model.Room{}, fmt.Errorf("room name is required")
	}
	return s.store.UpsertRoom(ctx, room)
}

func (s *Service) DeleteRoom(ctx context.Context, id string) error {
	return mapStoreErr(s.store.DeleteRoom(ctx, id))
}

func (s *Service) ListDevices(ctx context.Context, filter ListDevicesFilter) ([]model.Device, error) {
	return s.store.ListDevices(ctx, filter.EdgeID, filter.RoomID)
}

func (s *Service) GetDevice(ctx context.Context, id string) (model.Device, error) {
	device, err := s.store.GetDevice(ctx, id)
	return device, mapStoreErr(err)
}

func (s *Service) UpsertDevice(ctx context.Context, device model.Device) (model.Device, error) {
	if device.ID == "" {
		device.ID = newID("device")
	}
	if device.Name == "" || device.DeviceType == "" {
		return model.Device{}, fmt.Errorf("device name and type are required")
	}
	if device.State == "" {
		device.State = "unknown"
	}
	return s.store.UpsertDevice(ctx, device)
}

func (s *Service) DeleteDevice(ctx context.Context, id string) error {
	return mapStoreErr(s.store.DeleteDevice(ctx, id))
}

func (s *Service) SyncInventory(ctx context.Context, req *devicev1.SyncInventoryRequest) (string, error) {
	if req.GetEdgeId() == "" {
		return "", fmt.Errorf("edge_id is required")
	}

	rooms := make([]model.Room, 0, len(req.GetRooms()))
	for _, room := range req.GetRooms() {
		rooms = append(rooms, model.Room{
			ID:     room.GetRoomId(),
			EdgeID: req.GetEdgeId(),
			Name:   room.GetName(),
			Floor:  room.GetFloor(),
		})
	}

	devices := make([]model.Device, 0, len(req.GetDevices()))
	for _, device := range req.GetDevices() {
		devices = append(devices, model.Device{
			ID:             device.GetDeviceId(),
			EdgeID:         req.GetEdgeId(),
			RoomID:         device.GetRoomId(),
			Name:           device.GetName(),
			DeviceType:     device.GetDeviceType(),
			EntityID:       device.GetEntityId(),
			State:          device.GetState(),
			OfflineCapable: device.GetOfflineCapable(),
			UpdatedAt:      fromProtoTime(device.GetUpdatedAt()),
		})
	}

	if err := s.store.SyncInventory(ctx, req.GetEdgeId(), rooms, devices); err != nil {
		return "", err
	}
	return newID("sync"), nil
}

func (s *Service) UpsertDeviceState(ctx context.Context, req *devicev1.UpsertDeviceStateRequest) error {
	state := req.GetState()
	if state == nil {
		return fmt.Errorf("device state is required")
	}
	if state.GetDeviceId() == "" && state.GetEntityId() == "" {
		return fmt.Errorf("device_id or entity_id is required")
	}
	changedAt := fromProtoTime(state.GetChangedAt())
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	return mapStoreErr(s.store.UpdateDeviceState(
		ctx,
		req.GetEdgeId(),
		state.GetDeviceId(),
		state.GetEntityId(),
		state.GetState(),
		changedAt,
	))
}

func (s *Service) ExecuteCommand(ctx context.Context, req *devicev1.ExecuteCommandRequest) (string, error) {
	commandID := req.GetCommandId()
	if commandID == "" {
		commandID = newID("cmd")
	}
	if req.GetTargetState() == "" {
		return "", fmt.Errorf("target_state is required")
	}
	if req.GetDeviceId() == "" && req.GetEntityId() == "" {
		return "", fmt.Errorf("device_id or entity_id is required")
	}
	return s.store.ExecuteCommand(ctx, model.Command{
		ID:          commandID,
		DeviceID:    req.GetDeviceId(),
		EntityID:    req.GetEntityId(),
		TargetState: req.GetTargetState(),
		Source:      req.GetSource(),
		Status:      "executed",
		CreatedAt:   time.Now().UTC(),
		ExecutedAt:  time.Now().UTC(),
	})
}

func fromProtoTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().UTC()
}

func mapStoreErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), rand.Intn(1000000))
}
