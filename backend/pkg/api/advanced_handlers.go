package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/registryx/registryx/backend/pkg/costs"
	"github.com/registryx/registryx/backend/pkg/intelligence"
	"github.com/registryx/registryx/backend/pkg/middleware"
)

// AdvancedHandler handles advanced feature endpoints
type AdvancedHandler struct {
	Intelligence *intelligence.Service
	Costs        *costs.Service
}

// NewAdvancedHandler creates a new advanced features handler
func NewAdvancedHandler(intel *intelligence.Service, costSvc *costs.Service) *AdvancedHandler {
	return &AdvancedHandler{
		Intelligence: intel,
		Costs:        costSvc,
	}
}

// GetPrioritizedVulnerabilities returns vulnerabilities sorted by priority
func (h *AdvancedHandler) GetPrioritizedVulnerabilities(w http.ResponseWriter, r *http.Request) {
	manifestIDStr := r.URL.Query().Get("manifest_id")
	if manifestIDStr == "" {
		http.Error(w, "manifest_id required", http.StatusBadRequest)
		return
	}

	manifestID, err := uuid.Parse(manifestIDStr)
	if err != nil {
		http.Error(w, "invalid manifest_id", http.StatusBadRequest)
		return
	}

	priorities, err := h.Intelligence.GetPrioritizedVulnerabilities(r.Context(), manifestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(priorities)
}

// GetVulnIntelligence returns intelligence data for a CVE
func (h *AdvancedHandler) GetVulnIntelligence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cveID := vars["cve"]

	if cveID == "" {
		http.Error(w, "CVE ID required", http.StatusBadRequest)
		return
	}

	intel, err := h.Intelligence.GetVulnIntelligence(r.Context(), cveID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(intel)
}

// RefreshEPSS triggers a refresh of EPSS data
func (h *AdvancedHandler) RefreshEPSS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		// Run in background
		err := h.Intelligence.RefreshEPSSData(context.Background())
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("EPSS refresh error: %s\n", err.Error())
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "refresh started",
	})
}

// GetCostDashboard returns the cost dashboard summary
func (h *AdvancedHandler) GetCostDashboard(w http.ResponseWriter, r *http.Request) {
	// Extract User & Role
	role, _ := r.Context().Value(middleware.RoleKey).(string)
	
	// Parse UserID
	var userID uuid.UUID
	userIDRaw := r.Context().Value(middleware.UserKey)
	if userIDRaw != nil {
		if uidStr, ok := userIDRaw.(string); ok {
			userID, _ = uuid.Parse(uidStr)
		} else if uid, ok := userIDRaw.(uuid.UUID); ok {
			userID = uid
		}
	} else if role != "admin" { // Only admin might not need user ID if logic allowed (but here admin needs role)
		// If unauthenticated (middleware should catch, but if not):
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dashboard, err := h.Costs.GetDashboard(r.Context(), userID, role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// GetZombieImages returns list of zombie images
func (h *AdvancedHandler) GetZombieImages(w http.ResponseWriter, r *http.Request) {
	// Extract User & Role
	role, _ := r.Context().Value(middleware.RoleKey).(string)
	
	// Parse UserID
	var userID uuid.UUID
	userIDRaw := r.Context().Value(middleware.UserKey)
	if userIDRaw != nil {
		if uidStr, ok := userIDRaw.(string); ok {
			userID, _ = uuid.Parse(uidStr)
		} else if uid, ok := userIDRaw.(uuid.UUID); ok {
			userID = uid
		}
	} else if role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	zombies, err := h.Costs.DetectZombieImages(r.Context(), 90, userID, role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(zombies)
}

// RefreshCosts triggers a cost recalculation
func (h *AdvancedHandler) RefreshCosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		// Run in background with independent context
		err := h.Costs.RefreshAllCosts(context.Background())
		if err != nil {
			println("Cost refresh error:", err.Error())
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "refresh started",
	})
}

// CleanupZombies deletes zombie images
func (h *AdvancedHandler) CleanupZombies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract User & Role
	role, _ := r.Context().Value(middleware.RoleKey).(string)
	
	// Parse UserID
	var userID uuid.UUID
	userIDRaw := r.Context().Value(middleware.UserKey)
	if userIDRaw != nil {
		if uidStr, ok := userIDRaw.(string); ok {
			userID, _ = uuid.Parse(uidStr)
		} else if uid, ok := userIDRaw.(uuid.UUID); ok {
			userID = uid
		}
	} else if role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Default values
	daysThreshold := 180
	dryRun := true

	// 1. Try to parse from query parameters first
	q := r.URL.Query()
	if dt := q.Get("days_threshold"); dt != "" {
		if val, err := strconv.Atoi(dt); err == nil && val > 0 {
			daysThreshold = val
		}
	}
	if dr := q.Get("dry_run"); dr != "" {
		if val, err := strconv.ParseBool(dr); err == nil {
			dryRun = val
		}
	}

	// 2. Parse request body for options (overrides query params if present)
	var req struct {
		DaysThreshold int  `json:"days_threshold"`
		DryRun        *bool `json:"dry_run"` // Use pointer to check presence
	}

	// Try to decode ONLY if body is not empty
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.DaysThreshold > 0 {
				daysThreshold = req.DaysThreshold
			}
			if req.DryRun != nil {
				dryRun = *req.DryRun
			}
		}
	}

	// Ensure we don't accidentally delete everything if threshold is too low
	if daysThreshold < 30 {
		daysThreshold = 30
	}
	
	fmt.Printf("[API] CleanupZombies: dry_run=%v, days_threshold=%d, role=%s\n", dryRun, daysThreshold, role)

	count, err := h.Costs.CleanupZombies(r.Context(), daysThreshold, dryRun, userID, role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted_count": count,
		"dry_run":       dryRun,
	})
}
