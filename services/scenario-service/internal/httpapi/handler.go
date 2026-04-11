package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"scenario-service/internal/model"
	"scenario-service/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.health)

	mux.HandleFunc("GET /api/v1/scenarios", h.listScenarios)
	mux.HandleFunc("POST /api/v1/scenarios", h.saveScenario)
	mux.HandleFunc("GET /api/v1/scenarios/offline", h.listOfflineScenarios)
	mux.HandleFunc("GET /api/v1/scenarios/voice/commands", h.listVoiceCommands)
	mux.HandleFunc("POST /api/v1/scenarios/voice/execute", h.executeVoiceCommand)
	mux.HandleFunc("POST /api/v1/scenarios/evaluate", h.evaluateEvent)
	mux.HandleFunc("GET /api/v1/scenarios/{id}", h.getScenario)
	mux.HandleFunc("PUT /api/v1/scenarios/{id}", h.saveScenario)
	mux.HandleFunc("DELETE /api/v1/scenarios/{id}", h.deleteScenario)

	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios, err := h.service.ListScenarios(r.Context(), r.URL.Query().Get("edge_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, scenarios)
}

func (h *Handler) listOfflineScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios, err := h.service.GetOfflineScenarios(r.Context(), r.URL.Query().Get("edge_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, scenarios)
}

func (h *Handler) listVoiceCommands(w http.ResponseWriter, r *http.Request) {
	commands, err := h.service.ListVoiceCommands(r.Context(), r.URL.Query().Get("edge_id"), r.URL.Query().Get("room_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, commands)
}

func (h *Handler) getScenario(w http.ResponseWriter, r *http.Request) {
	scenario, err := h.service.GetScenario(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, scenario)
}

func (h *Handler) saveScenario(w http.ResponseWriter, r *http.Request) {
	var scenario model.Scenario
	if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if id := r.PathValue("id"); id != "" {
		scenario.ID = id
	}

	saved, err := h.service.SaveScenario(r.Context(), scenario)
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *Handler) deleteScenario(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteScenario(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) evaluateEvent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		model.EventEnvelope
		DeferExecution bool `json:"defer_execution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	decision, err := h.service.EvaluateEvent(r.Context(), body.EventEnvelope, body.DeferExecution)
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, decision)
}

func (h *Handler) executeVoiceCommand(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EdgeID         string `json:"edge_id"`
		RoomID         string `json:"room_id"`
		CommandName    string `json:"command_name"`
		Source         string `json:"source"`
		DeferExecution bool   `json:"defer_execution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	scenario, decision, err := h.service.ExecuteVoiceCommand(r.Context(), body.EdgeID, body.RoomID, body.CommandName, body.Source, body.DeferExecution)
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"scenario_id":   scenario.ID,
		"scenario_name": scenario.Name,
		"decision":      decision,
		"status":        decision.Status,
	})
}

func mapError(err error) int {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}
