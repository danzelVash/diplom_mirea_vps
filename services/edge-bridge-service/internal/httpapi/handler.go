package httpapi

import (
	"encoding/json"
	"net/http"

	bridgeservice "edge-bridge-service/internal/service"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
)

type Handler struct {
	service *bridgeservice.Service
}

func New(service *bridgeservice.Service) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /api/v1/edges/register", h.registerEdge)
	mux.HandleFunc("POST /api/v1/edges/inventory/sync", h.syncInventory)
	mux.HandleFunc("POST /api/v1/edges/events", h.publishEvent)
	mux.HandleFunc("GET /api/v1/edges/{id}/commands", h.pollCommands)
	mux.HandleFunc("GET /api/v1/edges/{id}/offline-scenarios", h.listOfflineScenarios)
	mux.HandleFunc("GET /api/v1/edges/{id}/scenarios", h.listScenarios)
	mux.HandleFunc("GET /api/v1/edges/{id}/voice-commands", h.listVoiceCommands)
	mux.HandleFunc("GET /api/v1/edges/{id}/status", h.edgeStatus)

	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) registerEdge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Edge bridgeservice.EdgeRegistration `json:"edge"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	status, err := h.service.RegisterEdge(r.Context(), body.Edge)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "registered",
		"edge":   status,
	})
}

func (h *Handler) syncInventory(w http.ResponseWriter, r *http.Request) {
	var body bridgeservice.InventorySync
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	syncID, err := h.service.SyncInventory(r.Context(), body)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "synced", "sync_id": syncID})
}

func (h *Handler) publishEvent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Event bridgeservice.Event `json:"event"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := h.service.PublishEvent(r.Context(), body.Event)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) pollCommands(w http.ResponseWriter, r *http.Request) {
	commands, err := h.service.PollCommands(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, commands)
}

func (h *Handler) listOfflineScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios, err := h.service.GetOfflineScenarios(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mapScenarios(scenarios))
}

func (h *Handler) listScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios, err := h.service.ListScenarios(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mapScenarios(scenarios))
}

func (h *Handler) listVoiceCommands(w http.ResponseWriter, r *http.Request) {
	commands, err := h.service.ListVoiceCommands(r.Context(), r.PathValue("id"), r.URL.Query().Get("room_id"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	response := make([]map[string]any, 0, len(commands))
	for _, command := range commands {
		response = append(response, map[string]any{
			"scenario_id":   command.GetScenarioId(),
			"scenario_name": command.GetScenarioName(),
			"edge_id":       command.GetEdgeId(),
			"room_id":       command.GetRoomId(),
			"command_name":  command.GetCommandName(),
			"device_id":     command.GetDeviceId(),
			"entity_id":     command.GetEntityId(),
			"target_state":  command.GetTargetState(),
			"priority":      command.GetPriority(),
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) edgeStatus(w http.ResponseWriter, r *http.Request) {
	status, lastEvent, ok := h.service.GetEdgeStatus(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "edge not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"edge":       status,
		"last_event": lastEvent,
	})
}

func mapScenarios(values []*scenariov1.Scenario) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, scenario := range values {
		result = append(result, map[string]any{
			"id":               scenario.GetScenarioId(),
			"edge_id":          scenario.GetEdgeId(),
			"name":             scenario.GetName(),
			"enabled":          scenario.GetEnabled(),
			"priority":         scenario.GetPriority(),
			"offline_eligible": scenario.GetOfflineEligible(),
			"triggers":         mapTriggers(scenario.GetTriggers()),
			"conditions":       mapConditions(scenario.GetConditions()),
			"actions":          mapActions(scenario.GetActions()),
			"updated_at":       scenario.GetUpdatedAt().AsTime(),
		})
	}
	return result
}

func mapTriggers(values []*scenariov1.Trigger) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, trigger := range values {
		result = append(result, map[string]any{
			"trigger_type":   trigger.GetTriggerType(),
			"event_type":     trigger.GetEventType(),
			"entity_id":      trigger.GetEntityId(),
			"expected_state": trigger.GetExpectedState(),
			"command_name":   trigger.GetCommandName(),
		})
	}
	return result
}

func mapConditions(values []*scenariov1.Condition) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, condition := range values {
		result = append(result, map[string]any{
			"condition_type": condition.GetConditionType(),
			"field":          condition.GetField(),
			"expected_value": condition.GetExpectedValue(),
		})
	}
	return result
}

func mapActions(values []*scenariov1.Action) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, action := range values {
		result = append(result, map[string]any{
			"action_type":  action.GetActionType(),
			"device_id":    action.GetDeviceId(),
			"entity_id":    action.GetEntityId(),
			"target_state": action.GetTargetState(),
		})
	}
	return result
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}
