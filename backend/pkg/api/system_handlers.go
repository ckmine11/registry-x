package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/registryx/registryx/backend/pkg/middleware"
)

type GCReport struct {
	BlobsDeleted     int64   `json:"blobsDeleted"`
	ManifestsDeleted int64   `json:"manifestsDeleted"`
	SpaceFreed   int64   `json:"spaceFreedBytes"` // Best effort
	SpaceFreedMB string  `json:"spaceFreedMB"`
	Duration     string  `json:"duration"`
	Errors       []string `json:"errors,omitempty"`
}

func (h *DashboardHandler) GarbageCollect(w http.ResponseWriter, r *http.Request) {
	// Security: Require Admin/Auth
	user := r.Context().Value(middleware.UserKey)
	if user == nil || user == "anonymous" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	start := time.Now()
	report := &GCReport{}

	// Check if this is a dry-run (preview mode)
	dryRun := r.URL.Query().Get("dryRun") == "true"

	// 0. Delete Untagged Manifests (Step 4 Auto-Cleanup)
	// Must be done BEFORE fetching orphans, as deleting manifests might orphan more blobs.
	if !dryRun {
		mCount, err := h.Metadata.DeleteUntaggedManifests(r.Context())
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Failed to cleanup manifests: %v", err))
		} else {
			report.ManifestsDeleted = mCount
			fmt.Printf("[GC] Deleted %d untagged manifests\n", mCount)
		}
	}

	// 1. Get Orphaned Blobs
	orphans, err := h.Metadata.GetOrphanedBlobs(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get orphaned blobs: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate totals
	var totalSize int64
	for _, orphan := range orphans {
		totalSize += orphan.Size
	}

	report.BlobsDeleted = int64(len(orphans))
	report.SpaceFreed = totalSize
	report.SpaceFreedMB = fmt.Sprintf("%.2f MB", float64(totalSize)/1024/1024)

	// If dry-run, return preview without deleting
	if dryRun {
		report.Duration = time.Since(start).String()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
		return
	}

	// 2. Actually delete blobs
	var deletedCount int64
	var deletedSize int64
	for _, orphan := range orphans {
		// 2a. Delete from Storage (MinIO)
		blobPath := path.Join("blobs", orphan.Digest)
		
		err := h.Storage.Delete(r.Context(), blobPath)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Failed to delete blob %s from storage: %v", orphan.Digest, err))
			continue
		}

		// 2b. Delete from DB
		err = h.Metadata.DeleteBlob(r.Context(), orphan.Digest)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Failed to delete blob %s from DB: %v", orphan.Digest, err))
			continue
		}

		deletedCount++
		deletedSize += orphan.Size
	}

	report.BlobsDeleted = deletedCount
	report.SpaceFreed = deletedSize
	report.SpaceFreedMB = fmt.Sprintf("%.2f MB", float64(deletedSize)/1024/1024)
	report.Duration = time.Since(start).String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// HealthCheck returns the status of the service
func (h *DashboardHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    status := map[string]string{
        "status": "ok",
        "time": time.Now().Format(time.RFC3339),
        "version": "2.2",
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}
