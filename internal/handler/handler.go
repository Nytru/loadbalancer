package handler

import (
	"cloud-test/internal/clients"
	"cloud-test/internal/net"
	"cloud-test/internal/ratelimit"
	"encoding/json"
	"net/http"
	"time"
)

var _ http.Handler = (*SetupHandler)(nil)

type SetupHandler struct {
	manager *clients.ClientLimitManager
	limiter ratelimit.Limiter
}

func NewClientHandler(manager *clients.ClientLimitManager, limiter ratelimit.Limiter) *SetupHandler {
	return &SetupHandler{manager: manager, limiter: limiter}
}

func (h *SetupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handleCreate(w, r)
	case http.MethodPut:
		h.handleUpdate(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		net.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

type CreateLimitRequest struct {
	ID               string `json:"id"`
	Capacity         int    `json:"capacity"`
	RefillIntervalMs int64  `json:"refill_interval_ms"`
}

func (h *SetupHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	ID := r.URL.Query().Get("id")
	if ID == "" {
		net.WriteError(w, http.StatusBadRequest, "ID is required")
		return
	}

	clientLimit, err := h.manager.LookUp(r.Context(), ID)
	if err != nil {
		net.WriteError(w, http.StatusInternalServerError, "Failed to get client limit")
		return
	}

	if clientLimit.ID == "" {
		net.WriteError(w, http.StatusNotFound, "Client limit not found")
		return
	}

	response := CreateLimitRequest{
		ID:               clientLimit.ID,
		Capacity:         clientLimit.Capacity,
		RefillIntervalMs: int64(clientLimit.RefillInterval / time.Millisecond),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SetupHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var cfg CreateLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		net.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	c := clients.ClientLimit{
		ID:             cfg.ID,
		Capacity:       cfg.Capacity,
		RefillInterval: time.Duration(cfg.RefillIntervalMs) * time.Millisecond,
	}

	exists, err := h.manager.LookUp(r.Context(), c.ID)
	if err != nil {
		net.WriteError(w, http.StatusInternalServerError, "Failed to get client limit")
		return
	}
	if exists.ID != "" {
		net.WriteError(w, http.StatusConflict, "Client limit already exists")
		return
	}

	if c.ID == "" {
		net.WriteError(w, http.StatusBadRequest, "ID is required")
		return
	}

	err = h.manager.Add(r.Context(), c)
	if err != nil {
		net.WriteError(w, http.StatusInternalServerError, "Failed to add client")
		return
	}
	h.limiter.UpdateRules(r.Context(), c.ID, c.Capacity, c.RefillInterval)
	w.WriteHeader(http.StatusOK)
}

type UpdateLimitRequest struct {
	ID               string `json:"id"`
	Capacity         int    `json:"capacity"`
	RefillIntervalMs int64  `json:"refill_interval_ms"`
}

func (h *SetupHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	var cfg UpdateLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		net.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	c := clients.ClientLimit{
		ID:             cfg.ID,
		Capacity:       cfg.Capacity,
		RefillInterval: time.Duration(cfg.RefillIntervalMs) * time.Millisecond,
	}

	exists, err := h.manager.LookUp(r.Context(), c.ID)
	if err != nil {
		net.WriteError(w, http.StatusInternalServerError, "Failed to get client limit")
		return
	}
	if exists.ID == "" {
		net.WriteError(w, http.StatusNotFound, "Client limit not found")
		return
	}

	err = h.manager.Update(r.Context(), c)
	if err != nil {
		net.WriteError(w, http.StatusInternalServerError, "Failed to add client")
		return
	}
	h.limiter.UpdateRules(r.Context(), c.ID, c.Capacity, c.RefillInterval)

	w.WriteHeader(http.StatusOK)
}

func (h *SetupHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	ID := r.URL.Query().Get("id")
	if ID != "" {
		err := h.manager.Remove(r.Context(), ID)
		if err != nil {
			net.WriteError(w, http.StatusInternalServerError, "Failed to remove client")
			return
		}
	}

	h.manager.Remove(r.Context(), ID)
	h.limiter.SetDefault(r.Context(), ID)

	w.WriteHeader(http.StatusOK)
}
