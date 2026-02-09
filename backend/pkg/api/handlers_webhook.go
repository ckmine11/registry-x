package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/registryx/registryx/backend/pkg/middleware"
)

// ListWebhooks GET /api/v1/system/webhooks
func (h *DashboardHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	// Security: Admin Only
	role := r.Context().Value(middleware.RoleKey)
	if role != "admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	hooks, err := h.Webhook.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": hooks})
}

// CreateWebhook POST /api/v1/system/webhooks
func (h *DashboardHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	// Security: Admin Only
	role := r.Context().Value(middleware.RoleKey)
	if role != "admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	var req struct {
		URL    string   `json:"url"`
		Type   string   `json:"type"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CreateWebhook] Failed to decode request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("[CreateWebhook] Request: URL=%s, Type=%s, Events=%v", req.URL, req.Type, req.Events)

	if req.URL == "" {
		log.Printf("[CreateWebhook] URL is empty")
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	hook, err := h.Webhook.Create(r.Context(), req.URL, req.Type, req.Events)
	if err != nil {
		log.Printf("[CreateWebhook] Failed to create webhook: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[CreateWebhook] Successfully created webhook: %s", hook.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hook)
}

// DeleteWebhook DELETE /api/v1/system/webhooks/{id}
func (h *DashboardHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	// Security: Admin Only
	role := r.Context().Value(middleware.RoleKey)
	if role != "admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Webhook.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestWebhook POST /api/v1/system/webhooks/{id}/test
func (h *DashboardHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	// Security: Admin Only
	role := r.Context().Value(middleware.RoleKey)
	if role != "admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Webhook.Test(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Often 400 if test fails
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Test payload sent successfully"})
}
