package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/registryx/registryx/backend/pkg/middleware"
)

// ListUsers GET /api/v1/users
func (h *DashboardHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Security: Admin Only
	role := r.Context().Value(middleware.RoleKey)
	if role != "admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	users, err := h.Auth.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": users})
}

// InviteUser POST /api/v1/users
func (h *DashboardHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	// Security: Admin Only
	role := r.Context().Value(middleware.RoleKey)
	if role != "admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "developer"
	}

	user, tempPass, err := h.Auth.InviteUser(r.Context(), req.Email, req.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return the temporary password/key so the admin can share it
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
		"temporaryPassword": tempPass,
	})
}

// UpdateUserRole PUT /api/v1/users/{id}/role
func (h *DashboardHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Auth.UpdateRole(r.Context(), id, req.Role); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUser DELETE /api/v1/users/{id}
func (h *DashboardHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Security: Admin Only
	role := r.Context().Value(middleware.RoleKey)
	if role != "admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	log.Printf("[DeleteUser] Request to delete user: %s", idStr)
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("[DeleteUser] Invalid UUID: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Auth.DeleteUser(r.Context(), id); err != nil {
		log.Printf("[DeleteUser] Failed to delete user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[DeleteUser] Successfully deleted user: %s", id)
	w.WriteHeader(http.StatusNoContent)
}
