package model

import (
	"time"

	scenariov1 "scenario-service/pkg/pb/scenario/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Trigger struct {
	TriggerType   string `json:"trigger_type"`
	EventType     string `json:"event_type"`
	EntityID      string `json:"entity_id"`
	ExpectedState string `json:"expected_state"`
	CommandName   string `json:"command_name"`
}

func (t Trigger) ToProto() *scenariov1.Trigger {
	return &scenariov1.Trigger{
		TriggerType:   t.TriggerType,
		EventType:     t.EventType,
		EntityId:      t.EntityID,
		ExpectedState: t.ExpectedState,
		CommandName:   t.CommandName,
	}
}

type Condition struct {
	ConditionType string `json:"condition_type"`
	Field         string `json:"field"`
	ExpectedValue string `json:"expected_value"`
}

func (c Condition) ToProto() *scenariov1.Condition {
	return &scenariov1.Condition{
		ConditionType: c.ConditionType,
		Field:         c.Field,
		ExpectedValue: c.ExpectedValue,
	}
}

type Action struct {
	ActionType  string `json:"action_type"`
	DeviceID    string `json:"device_id"`
	EntityID    string `json:"entity_id"`
	TargetState string `json:"target_state"`
}

func (a Action) ToProto() *scenariov1.Action {
	return &scenariov1.Action{
		ActionType:  a.ActionType,
		DeviceId:    a.DeviceID,
		EntityId:    a.EntityID,
		TargetState: a.TargetState,
	}
}

type Scenario struct {
	ID              string      `json:"id"`
	EdgeID          string      `json:"edge_id"`
	Name            string      `json:"name"`
	Enabled         bool        `json:"enabled"`
	Priority        int32       `json:"priority"`
	OfflineEligible bool        `json:"offline_eligible"`
	Triggers        []Trigger   `json:"triggers"`
	Conditions      []Condition `json:"conditions"`
	Actions         []Action    `json:"actions"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func (s Scenario) ToProto() *scenariov1.Scenario {
	return &scenariov1.Scenario{
		ScenarioId:      s.ID,
		EdgeId:          s.EdgeID,
		Name:            s.Name,
		Enabled:         s.Enabled,
		Priority:        s.Priority,
		OfflineEligible: s.OfflineEligible,
		Triggers:        triggersToProto(s.Triggers),
		Conditions:      conditionsToProto(s.Conditions),
		Actions:         actionsToProto(s.Actions),
		UpdatedAt:       timestamppb.New(s.UpdatedAt),
	}
}

type EventEnvelope struct {
	ID         string    `json:"id"`
	EdgeID     string    `json:"edge_id"`
	RoomID     string    `json:"room_id"`
	EntityID   string    `json:"entity_id"`
	EventType  string    `json:"event_type"`
	State      string    `json:"state"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (e EventEnvelope) ToProto() *scenariov1.EventEnvelope {
	return &scenariov1.EventEnvelope{
		EventId:    e.ID,
		EdgeId:     e.EdgeID,
		RoomId:     e.RoomID,
		EntityId:   e.EntityID,
		EventType:  e.EventType,
		State:      e.State,
		OccurredAt: timestamppb.New(e.OccurredAt),
	}
}

type Decision struct {
	ID                 string    `json:"id"`
	Status             string    `json:"status"`
	Actions            []Action  `json:"actions"`
	MatchedScenarioIDs []string  `json:"matched_scenario_ids"`
	CreatedAt          time.Time `json:"created_at"`
}

func (d Decision) ToProto() *scenariov1.Decision {
	return &scenariov1.Decision{
		DecisionId:         d.ID,
		Status:             d.Status,
		Actions:            actionsToProto(d.Actions),
		MatchedScenarioIds: d.MatchedScenarioIDs,
	}
}

type VoiceCommand struct {
	ScenarioID   string `json:"scenario_id"`
	ScenarioName string `json:"scenario_name"`
	EdgeID       string `json:"edge_id"`
	RoomID       string `json:"room_id"`
	CommandName  string `json:"command_name"`
	DeviceID     string `json:"device_id"`
	EntityID     string `json:"entity_id"`
	TargetState  string `json:"target_state"`
	Priority     int32  `json:"priority"`
}

func (c VoiceCommand) ToProto() *scenariov1.VoiceCommand {
	return &scenariov1.VoiceCommand{
		ScenarioId:   c.ScenarioID,
		ScenarioName: c.ScenarioName,
		EdgeId:       c.EdgeID,
		RoomId:       c.RoomID,
		CommandName:  c.CommandName,
		DeviceId:     c.DeviceID,
		EntityId:     c.EntityID,
		TargetState:  c.TargetState,
		Priority:     c.Priority,
	}
}

func ScenarioFromProto(src *scenariov1.Scenario) Scenario {
	if src == nil {
		return Scenario{}
	}

	return Scenario{
		ID:              src.GetScenarioId(),
		EdgeID:          src.GetEdgeId(),
		Name:            src.GetName(),
		Enabled:         src.GetEnabled(),
		Priority:        src.GetPriority(),
		OfflineEligible: src.GetOfflineEligible(),
		Triggers:        triggersFromProto(src.GetTriggers()),
		Conditions:      conditionsFromProto(src.GetConditions()),
		Actions:         actionsFromProto(src.GetActions()),
		UpdatedAt:       fromProtoTime(src.GetUpdatedAt()),
	}
}

func EventFromProto(src *scenariov1.EventEnvelope) EventEnvelope {
	if src == nil {
		return EventEnvelope{}
	}

	return EventEnvelope{
		ID:         src.GetEventId(),
		EdgeID:     src.GetEdgeId(),
		RoomID:     src.GetRoomId(),
		EntityID:   src.GetEntityId(),
		EventType:  src.GetEventType(),
		State:      src.GetState(),
		OccurredAt: fromProtoTime(src.GetOccurredAt()),
	}
}

func triggersToProto(values []Trigger) []*scenariov1.Trigger {
	result := make([]*scenariov1.Trigger, 0, len(values))
	for _, value := range values {
		result = append(result, value.ToProto())
	}
	return result
}

func conditionsToProto(values []Condition) []*scenariov1.Condition {
	result := make([]*scenariov1.Condition, 0, len(values))
	for _, value := range values {
		result = append(result, value.ToProto())
	}
	return result
}

func actionsToProto(values []Action) []*scenariov1.Action {
	result := make([]*scenariov1.Action, 0, len(values))
	for _, value := range values {
		result = append(result, value.ToProto())
	}
	return result
}

func triggersFromProto(values []*scenariov1.Trigger) []Trigger {
	result := make([]Trigger, 0, len(values))
	for _, value := range values {
		result = append(result, Trigger{
			TriggerType:   value.GetTriggerType(),
			EventType:     value.GetEventType(),
			EntityID:      value.GetEntityId(),
			ExpectedState: value.GetExpectedState(),
			CommandName:   value.GetCommandName(),
		})
	}
	return result
}

func conditionsFromProto(values []*scenariov1.Condition) []Condition {
	result := make([]Condition, 0, len(values))
	for _, value := range values {
		result = append(result, Condition{
			ConditionType: value.GetConditionType(),
			Field:         value.GetField(),
			ExpectedValue: value.GetExpectedValue(),
		})
	}
	return result
}

func actionsFromProto(values []*scenariov1.Action) []Action {
	result := make([]Action, 0, len(values))
	for _, value := range values {
		result = append(result, Action{
			ActionType:  value.GetActionType(),
			DeviceID:    value.GetDeviceId(),
			EntityID:    value.GetEntityId(),
			TargetState: value.GetTargetState(),
		})
	}
	return result
}

func fromProtoTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().UTC()
}
