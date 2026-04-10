package model

import (
	devicev1 "device-service/pkg/pb/device/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Room struct {
	ID        string    `json:"id"`
	EdgeID    string    `json:"edge_id"`
	Name      string    `json:"name"`
	Floor     string    `json:"floor,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r Room) ToProtoRoom() *devicev1.Room {
	return &devicev1.Room{
		RoomId: r.ID,
		Name:   r.Name,
		Floor:  r.Floor,
	}
}

type Device struct {
	ID             string    `json:"id"`
	EdgeID         string    `json:"edge_id"`
	RoomID         string    `json:"room_id,omitempty"`
	Name           string    `json:"name"`
	DeviceType     string    `json:"device_type"`
	EntityID       string    `json:"entity_id,omitempty"`
	State          string    `json:"state"`
	OfflineCapable bool      `json:"offline_capable"`
	LastChangedAt  time.Time `json:"last_changed_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (d Device) ToProto() *devicev1.Device {
	return &devicev1.Device{
		DeviceId:       d.ID,
		EdgeId:         d.EdgeID,
		RoomId:         d.RoomID,
		Name:           d.Name,
		DeviceType:     d.DeviceType,
		EntityId:       d.EntityID,
		State:          d.State,
		OfflineCapable: d.OfflineCapable,
		UpdatedAt:      timestamppb.New(d.UpdatedAt),
	}
}

type Command struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id,omitempty"`
	EntityID    string    `json:"entity_id,omitempty"`
	TargetState string    `json:"target_state"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ExecutedAt  time.Time `json:"executed_at"`
}
