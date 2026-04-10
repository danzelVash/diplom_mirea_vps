package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	devicev1 "device-service/pkg/pb/device/v1"
	"scenario-service/internal/model"
	"scenario-service/internal/store"

	"google.golang.org/grpc"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	ListScenarios(ctx context.Context, edgeID string, offlineOnly bool) ([]model.Scenario, error)
	GetScenario(ctx context.Context, id string) (model.Scenario, error)
	UpsertScenario(ctx context.Context, scenario model.Scenario) (model.Scenario, error)
	DeleteScenario(ctx context.Context, id string) error
	SaveDecision(ctx context.Context, eventID, edgeID string, decision model.Decision) error
}

type DeviceClient interface {
	ExecuteCommand(ctx context.Context, in *devicev1.ExecuteCommandRequest, opts ...grpc.CallOption) (*devicev1.ExecuteCommandResponse, error)
}

type Service struct {
	store  Store
	device DeviceClient
}

func New(store Store, device DeviceClient) *Service {
	return &Service{
		store:  store,
		device: device,
	}
}

func (s *Service) ListScenarios(ctx context.Context, edgeID string) ([]model.Scenario, error) {
	return s.store.ListScenarios(ctx, edgeID, false)
}

func (s *Service) GetScenario(ctx context.Context, id string) (model.Scenario, error) {
	scenario, err := s.store.GetScenario(ctx, id)
	return scenario, mapStoreErr(err)
}

func (s *Service) SaveScenario(ctx context.Context, scenario model.Scenario) (model.Scenario, error) {
	if scenario.ID == "" {
		scenario.ID = newID("scn")
	}
	if strings.TrimSpace(scenario.EdgeID) == "" {
		return model.Scenario{}, fmt.Errorf("edge_id is required")
	}
	if strings.TrimSpace(scenario.Name) == "" {
		return model.Scenario{}, fmt.Errorf("name is required")
	}
	if len(scenario.Triggers) == 0 {
		return model.Scenario{}, fmt.Errorf("at least one trigger is required")
	}
	if len(scenario.Actions) == 0 {
		return model.Scenario{}, fmt.Errorf("at least one action is required")
	}

	for index, action := range scenario.Actions {
		if action.TargetState == "" {
			return model.Scenario{}, fmt.Errorf("action[%d].target_state is required", index)
		}
		if action.DeviceID == "" && action.EntityID == "" {
			return model.Scenario{}, fmt.Errorf("action[%d].device_id or entity_id is required", index)
		}
	}

	for index, trigger := range scenario.Triggers {
		if trigger.TriggerType == "voice_command" && strings.TrimSpace(trigger.CommandName) == "" {
			return model.Scenario{}, fmt.Errorf("trigger[%d].command_name is required for voice_command", index)
		}
	}

	return s.store.UpsertScenario(ctx, scenario)
}

func (s *Service) DeleteScenario(ctx context.Context, id string) error {
	return mapStoreErr(s.store.DeleteScenario(ctx, id))
}

func (s *Service) GetOfflineScenarios(ctx context.Context, edgeID string) ([]model.Scenario, error) {
	scenarios, err := s.store.ListScenarios(ctx, edgeID, true)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.Scenario, 0, len(scenarios))
	for _, scenario := range scenarios {
		if scenario.Enabled && scenario.OfflineEligible {
			filtered = append(filtered, scenario)
		}
	}
	return filtered, nil
}

func (s *Service) ListVoiceCommands(ctx context.Context, edgeID, roomID string) ([]model.VoiceCommand, error) {
	scenarios, err := s.store.ListScenarios(ctx, edgeID, false)
	if err != nil {
		return nil, err
	}

	commands := make([]model.VoiceCommand, 0)
	for _, scenario := range scenarios {
		if !scenario.Enabled {
			continue
		}

		commandName := voiceCommandName(scenario)
		if commandName == "" {
			continue
		}
		if roomID != "" && !scenarioMatchesRoom(scenario, roomID) {
			continue
		}

		action := firstDeviceAction(scenario.Actions)
		commands = append(commands, model.VoiceCommand{
			ScenarioID:   scenario.ID,
			ScenarioName: scenario.Name,
			EdgeID:       scenario.EdgeID,
			RoomID:       roomConditionValue(scenario),
			CommandName:  commandName,
			DeviceID:     action.DeviceID,
			EntityID:     action.EntityID,
			TargetState:  action.TargetState,
			Priority:     scenario.Priority,
		})
	}

	return commands, nil
}

func (s *Service) ExecuteVoiceCommand(ctx context.Context, edgeID, roomID, commandName, source string) (model.Scenario, model.Decision, error) {
	if edgeID == "" {
		return model.Scenario{}, model.Decision{}, fmt.Errorf("edge_id is required")
	}
	if strings.TrimSpace(commandName) == "" {
		return model.Scenario{}, model.Decision{}, fmt.Errorf("command_name is required")
	}

	scenarios, err := s.store.ListScenarios(ctx, edgeID, false)
	if err != nil {
		return model.Scenario{}, model.Decision{}, err
	}

	normalizedCommand := normalize(commandName)
	var selected model.Scenario
	found := false
	for _, scenario := range scenarios {
		if !scenario.Enabled {
			continue
		}
		if roomID != "" && !scenarioMatchesRoom(scenario, roomID) {
			continue
		}
		if normalize(voiceCommandName(scenario)) != normalizedCommand {
			continue
		}
		if !found || scenario.Priority > selected.Priority {
			selected = scenario
			found = true
		}
	}

	if !found {
		return model.Scenario{}, model.Decision{}, ErrNotFound
	}

	decision, err := s.executeActions(ctx, selected, "voice:"+source)
	if err != nil {
		return model.Scenario{}, model.Decision{}, err
	}

	return selected, decision, nil
}

func (s *Service) EvaluateEvent(ctx context.Context, event model.EventEnvelope) (model.Decision, error) {
	if event.EventType == "" {
		return model.Decision{}, fmt.Errorf("event_type is required")
	}
	if event.EdgeID == "" {
		return model.Decision{}, fmt.Errorf("edge_id is required")
	}
	if event.ID == "" {
		event.ID = newID("event")
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	scenarios, err := s.store.ListScenarios(ctx, event.EdgeID, false)
	if err != nil {
		return model.Decision{}, err
	}

	decision := model.Decision{
		ID:        newID("decision"),
		Status:    "no_match",
		CreatedAt: time.Now().UTC(),
	}

	matchedScenarios := make([]model.Scenario, 0)
	for _, scenario := range scenarios {
		if !scenario.Enabled {
			continue
		}
		if matchesScenario(scenario, event) {
			matchedScenarios = append(matchedScenarios, scenario)
			decision.MatchedScenarioIDs = append(decision.MatchedScenarioIDs, scenario.ID)
			decision.Actions = append(decision.Actions, scenario.Actions...)
		}
	}

	if len(matchedScenarios) == 0 {
		if err := s.store.SaveDecision(ctx, event.ID, event.EdgeID, decision); err != nil {
			return model.Decision{}, err
		}
		return decision, nil
	}

	executedTotal := 0
	executableTotal := 0
	for _, scenario := range matchedScenarios {
		currentDecision, err := s.executeActions(ctx, scenario, "scenario:"+scenario.ID)
		if err != nil {
			return model.Decision{}, err
		}
		executableTotal += countExecutableActions(currentDecision.Actions)
		executedTotal += executedActionCount(currentDecision.Status, currentDecision.Actions)
	}

	decision.Status = aggregateStatus(executableTotal, executedTotal)
	if err := s.store.SaveDecision(ctx, event.ID, event.EdgeID, decision); err != nil {
		return model.Decision{}, err
	}
	return decision, nil
}

func (s *Service) executeActions(ctx context.Context, scenario model.Scenario, source string) (model.Decision, error) {
	decision := model.Decision{
		ID:                 newID("decision"),
		Status:             "matched",
		Actions:            append([]model.Action(nil), scenario.Actions...),
		MatchedScenarioIDs: []string{scenario.ID},
		CreatedAt:          time.Now().UTC(),
	}

	executable := 0
	executed := 0
	for _, action := range scenario.Actions {
		if !isDeviceCommand(action) {
			continue
		}
		executable++
		if s.device == nil {
			continue
		}

		_, err := s.device.ExecuteCommand(ctx, &devicev1.ExecuteCommandRequest{
			CommandId:   newID("cmd"),
			DeviceId:    action.DeviceID,
			EntityId:    action.EntityID,
			TargetState: action.TargetState,
			Source:      source,
		})
		if err == nil {
			executed++
		}
	}

	decision.Status = aggregateStatus(executable, executed)
	if err := s.store.SaveDecision(ctx, "", scenario.EdgeID, decision); err != nil {
		return model.Decision{}, err
	}
	return decision, nil
}

func matchesScenario(scenario model.Scenario, event model.EventEnvelope) bool {
	if scenario.EdgeID != "" && scenario.EdgeID != event.EdgeID {
		return false
	}
	if len(scenario.Triggers) == 0 {
		return false
	}

	triggerMatched := false
	for _, trigger := range scenario.Triggers {
		if matchesTrigger(trigger, event) {
			triggerMatched = true
			break
		}
	}
	if !triggerMatched {
		return false
	}

	for _, condition := range scenario.Conditions {
		if !matchesCondition(condition, event) {
			return false
		}
	}
	return true
}

func matchesTrigger(trigger model.Trigger, event model.EventEnvelope) bool {
	if trigger.TriggerType != "" && trigger.TriggerType != "event" {
		return false
	}
	if trigger.EventType != "" && trigger.EventType != event.EventType {
		return false
	}
	if trigger.EntityID != "" && trigger.EntityID != event.EntityID {
		return false
	}
	if trigger.ExpectedState != "" && trigger.ExpectedState != event.State {
		return false
	}
	return true
}

func matchesCondition(condition model.Condition, event model.EventEnvelope) bool {
	actual := fieldValue(event, condition.Field)
	switch condition.ConditionType {
	case "", "equals":
		return actual == condition.ExpectedValue
	case "not_equals":
		return actual != condition.ExpectedValue
	default:
		return false
	}
}

func fieldValue(event model.EventEnvelope, field string) string {
	switch field {
	case "edge_id":
		return event.EdgeID
	case "room_id":
		return event.RoomID
	case "entity_id":
		return event.EntityID
	case "event_type":
		return event.EventType
	case "state":
		return event.State
	default:
		return ""
	}
}

func isDeviceCommand(action model.Action) bool {
	return action.ActionType == "" || action.ActionType == "device_command"
}

func voiceCommandName(scenario model.Scenario) string {
	for _, trigger := range scenario.Triggers {
		if trigger.TriggerType == "voice_command" && trigger.CommandName != "" {
			return trigger.CommandName
		}
	}
	return ""
}

func scenarioMatchesRoom(scenario model.Scenario, roomID string) bool {
	expectedRoom := roomConditionValue(scenario)
	return expectedRoom == "" || expectedRoom == roomID
}

func roomConditionValue(scenario model.Scenario) string {
	for _, condition := range scenario.Conditions {
		if condition.Field == "room_id" && (condition.ConditionType == "" || condition.ConditionType == "equals") {
			return condition.ExpectedValue
		}
	}
	return ""
}

func firstDeviceAction(actions []model.Action) model.Action {
	for _, action := range actions {
		if isDeviceCommand(action) {
			return action
		}
	}
	return model.Action{}
}

func countExecutableActions(actions []model.Action) int {
	count := 0
	for _, action := range actions {
		if isDeviceCommand(action) {
			count++
		}
	}
	return count
}

func executedActionCount(status string, actions []model.Action) int {
	switch status {
	case "executed":
		return countExecutableActions(actions)
	case "partial":
		if countExecutableActions(actions) > 0 {
			return countExecutableActions(actions) - 1
		}
	}
	return 0
}

func aggregateStatus(executable, executed int) string {
	switch {
	case executable == 0:
		return "matched"
	case executed == 0:
		return "matched"
	case executed == executable:
		return "executed"
	default:
		return "partial"
	}
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
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
