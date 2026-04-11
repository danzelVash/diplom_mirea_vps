package model

import "time"

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
	EdgeID      string    `json:"edge_id,omitempty"`
	DeviceID    string    `json:"device_id,omitempty"`
	EntityID    string    `json:"entity_id,omitempty"`
	TargetState string    `json:"target_state"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
}

type CommandAck struct {
	CommandID string `json:"command_id"`
	DeviceID  string `json:"device_id,omitempty"`
	EntityID  string `json:"entity_id,omitempty"`
	State     string `json:"state,omitempty"`
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

type EventSnapshot struct {
	EventType  string    `json:"event_type"`
	State      string    `json:"state,omitempty"`
	DeviceID   string    `json:"device_id,omitempty"`
	EntityID   string    `json:"entity_id,omitempty"`
	RoomID     string    `json:"room_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

type EventResult struct {
	Status             string   `json:"status"`
	MatchedScenarioIDs []string `json:"matched_scenario_ids,omitempty"`
	QueuedCommands     int      `json:"queued_commands"`
}
