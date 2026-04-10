package model

import "time"

type EdgeNode struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	PublicAddr      string    `json:"public_addr,omitempty"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	LastSyncAt      time.Time `json:"last_sync_at"`
	ConnectionState string    `json:"connection_state"`
}

type Room struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Floor     string    `json:"floor,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Device struct {
	ID           string    `json:"id"`
	EdgeID       string    `json:"edge_id"`
	Name         string    `json:"name"`
	RoomID       string    `json:"room_id,omitempty"`
	DeviceType   string    `json:"device_type"`
	EntityID     string    `json:"entity_id,omitempty"`
	Status       string    `json:"status"`
	LastChanged  time.Time `json:"last_changed"`
	UpdatedAt    time.Time `json:"updated_at"`
	OfflineCapab bool      `json:"offline_capable"`
}

type Event struct {
	ID         string                 `json:"id"`
	EdgeID     string                 `json:"edge_id"`
	DeviceID   string                 `json:"device_id,omitempty"`
	EntityID   string                 `json:"entity_id,omitempty"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status,omitempty"`
	OccurredAt time.Time              `json:"occurred_at"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
}

type Trigger struct {
	Type     string `json:"type"`
	Cron     string `json:"cron,omitempty"`
	Event    string `json:"event,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
	EntityID string `json:"entity_id,omitempty"`
	Status   string `json:"status,omitempty"`
}

type Action struct {
	Type         string `json:"type"`
	DeviceID     string `json:"device_id,omitempty"`
	EntityID     string `json:"entity_id,omitempty"`
	TargetStatus string `json:"target_status,omitempty"`
	Message      string `json:"message,omitempty"`
}

type Scenario struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Enabled         bool      `json:"enabled"`
	Priority        int       `json:"priority"`
	EdgeID          string    `json:"edge_id,omitempty"`
	OfflineEligible bool      `json:"offline_eligible"`
	Triggers        []Trigger `json:"triggers"`
	Actions         []Action  `json:"actions"`
	CreatedAt       time.Time `json:"created_at"`
}

type Command struct {
	ID          string    `json:"id"`
	EdgeID       string    `json:"edge_id"`
	ScenarioID  string    `json:"scenario_id,omitempty"`
	DeviceID    string    `json:"device_id,omitempty"`
	EntityID    string    `json:"entity_id,omitempty"`
	TargetStatus string    `json:"target_status"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
	AckedAt     time.Time `json:"acked_at,omitempty"`
}

type Snapshot struct {
	Rooms     []Room              `json:"rooms"`
	Devices   []Device            `json:"devices"`
	Scenarios []Scenario          `json:"scenarios"`
	Edges     []EdgeNode          `json:"edges"`
	Commands  map[string][]Command `json:"commands"`
	Events    []Event             `json:"events"`
}
