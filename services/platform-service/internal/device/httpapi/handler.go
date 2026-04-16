package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"google.golang.org/protobuf/types/known/timestamppb"
	"platform-service/internal/device/model"
	"platform-service/internal/device/service"
	devicev1 "platform-service/pkg/pb/device/v1"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.health)

	mux.HandleFunc("GET /api/v1/rooms", h.listRooms)
	mux.HandleFunc("POST /api/v1/rooms", h.upsertRoom)
	mux.HandleFunc("GET /api/v1/rooms/{id}", h.getRoom)
	mux.HandleFunc("PUT /api/v1/rooms/{id}", h.upsertRoom)
	mux.HandleFunc("DELETE /api/v1/rooms/{id}", h.deleteRoom)

	mux.HandleFunc("GET /api/v1/devices", h.listDevices)
	mux.HandleFunc("POST /api/v1/devices", h.upsertDevice)
	mux.HandleFunc("GET /api/v1/devices/{id}", h.getDevice)
	mux.HandleFunc("PUT /api/v1/devices/{id}", h.upsertDevice)
	mux.HandleFunc("DELETE /api/v1/devices/{id}", h.deleteDevice)
	mux.HandleFunc("POST /api/v1/devices/{id}/state", h.updateDeviceState)
	mux.HandleFunc("POST /api/v1/devices/commands", h.executeCommand)
	mux.HandleFunc("POST /api/v1/inventory/sync", h.syncInventory)

	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.service.ListRooms(r.Context(), r.URL.Query().Get("edge_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, rooms)
}

func (h *Handler) getRoom(w http.ResponseWriter, r *http.Request) {
	room, err := h.service.GetRoom(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, room)
}

func (h *Handler) upsertRoom(w http.ResponseWriter, r *http.Request) {
	var room model.Room
	if err := json.NewDecoder(r.Body).Decode(&room); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if id := r.PathValue("id"); id != "" {
		room.ID = id
	}

	saved, err := h.service.UpsertRoom(r.Context(), room)
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *Handler) deleteRoom(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteRoom(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.service.ListDevices(r.Context(), service.ListDevicesFilter{
		EdgeID: r.URL.Query().Get("edge_id"),
		RoomID: r.URL.Query().Get("room_id"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, devices)
}

func (h *Handler) getDevice(w http.ResponseWriter, r *http.Request) {
	device, err := h.service.GetDevice(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (h *Handler) upsertDevice(w http.ResponseWriter, r *http.Request) {
	var device model.Device
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if id := r.PathValue("id"); id != "" {
		device.ID = id
	}

	saved, err := h.service.UpsertDevice(r.Context(), device)
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *Handler) deleteDevice(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteDevice(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) updateDeviceState(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EntityID string `json:"entity_id"`
		State    string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	err := h.service.UpsertDeviceState(r.Context(), &devicev1.UpsertDeviceStateRequest{
		EdgeId: r.URL.Query().Get("edge_id"),
		State: &devicev1.DeviceState{
			DeviceId:  r.PathValue("id"),
			EntityId:  body.EntityID,
			State:     body.State,
			ChangedAt: timestamppb.Now(),
		},
	})
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) executeCommand(w http.ResponseWriter, r *http.Request) {
	var req devicev1.ExecuteCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	commandID, err := h.service.ExecuteCommand(r.Context(), &req)
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"command_id": commandID, "status": "executed"})
}

func (h *Handler) syncInventory(w http.ResponseWriter, r *http.Request) {
	var req devicev1.SyncInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	syncID, err := h.service.SyncInventory(r.Context(), &req)
	if err != nil {
		writeError(w, mapError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"sync_id": syncID, "status": "synced"})
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
