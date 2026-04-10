package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danzelVash/diplom_mirea_vps/internal/model"
	"github.com/danzelVash/diplom_mirea_vps/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) http.Handler {
	h := &Handler{svc: svc}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/api/v1/edges/register", h.registerEdge)
	mux.HandleFunc("/api/v1/sync/devices", h.syncDevices)
	mux.HandleFunc("/api/v1/events", h.ingestEvents)
	mux.HandleFunc("/api/v1/rooms", h.rooms)
	mux.HandleFunc("/api/v1/devices", h.devices)
	mux.HandleFunc("/api/v1/scenarios", h.scenarios)
	mux.HandleFunc("/api/v1/edges/", h.edgeSubroutes)

	return logMiddleware(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (h *Handler) registerEdge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		PublicAddr string `json:"public_addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	edge, err := h.svc.RegisterEdge(req.ID, req.Name, req.PublicAddr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, edge)
}

func (h *Handler) syncDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		EdgeID  string         `json:"edge_id"`
		Rooms   []model.Room   `json:"rooms"`
		Devices []model.Device `json:"devices"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.EdgeID == "" {
		writeError(w, http.StatusBadRequest, "edge_id is required")
		return
	}
	if err := h.svc.UpsertRooms(req.Rooms); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.svc.UpsertDevices(req.EdgeID, req.Devices); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "accepted"})
}

func (h *Handler) ingestEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req model.Event
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.EdgeID == "" || req.Type == "" {
		writeError(w, http.StatusBadRequest, "edge_id and type are required")
		return
	}
	event, commands, err := h.svc.IngestEvent(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"event":    event,
		"commands": commands,
	})
}

func (h *Handler) rooms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.svc.ListRooms())
	case http.MethodPost:
		var room model.Room
		if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if room.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		created, err := h.svc.CreateRoom(room)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) devices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.svc.ListDevices(r.URL.Query().Get("edge_id")))
	case http.MethodPost:
		var device model.Device
		if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if device.EdgeID == "" || device.Name == "" || device.DeviceType == "" {
			writeError(w, http.StatusBadRequest, "edge_id, name and device_type are required")
			return
		}
		created, err := h.svc.CreateDevice(device)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) scenarios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		edgeID := r.URL.Query().Get("edge_id")
		offlineOnly := r.URL.Query().Get("offline_only") == "true"
		writeJSON(w, http.StatusOK, h.svc.ListScenarios(edgeID, offlineOnly))
	case http.MethodPost:
		var scenario model.Scenario
		if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if scenario.Name == "" || len(scenario.Triggers) == 0 || len(scenario.Actions) == 0 {
			writeError(w, http.StatusBadRequest, "name, triggers and actions are required")
			return
		}
		created, err := h.svc.CreateScenario(scenario)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) edgeSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/edges/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	edgeID := parts[0]
	resource := parts[1]

	switch resource {
	case "commands":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		commands, err := h.svc.PollCommands(edgeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, commands)
	case "offline-scenarios":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, h.svc.ListScenarios(edgeID, true))
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func mapError(err error) int {
	if errors.Is(err, service.ErrNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
