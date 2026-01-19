package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/registryx/registryx/backend/pkg/audit"
	"github.com/registryx/registryx/backend/pkg/config"
	"github.com/registryx/registryx/backend/pkg/metadata"
	"github.com/registryx/registryx/backend/pkg/middleware"
	"github.com/registryx/registryx/backend/pkg/policy"
	"github.com/registryx/registryx/backend/pkg/queue"
	"github.com/registryx/registryx/backend/pkg/scanner"
	"github.com/registryx/registryx/backend/pkg/storage"
	"github.com/registryx/registryx/backend/pkg/webhook"
)

type Handler struct {
	Config   *config.Config
	Storage  storage.Driver
	Metadata *metadata.Service
	Scanner  *scanner.Service
	Policy   *policy.Service
	Queue    *queue.Service
	Webhook  *webhook.Service
	Audit    *audit.Service
}

func NewHandler(cfg *config.Config, store storage.Driver, meta *metadata.Service, scan *scanner.Service, pol *policy.Service, q *queue.Service, hook *webhook.Service, aud *audit.Service) *Handler {
	return &Handler{
		Config:   cfg,
		Storage:  store,
		Metadata: meta,
		Scanner:  scan,
		Policy:   pol,
		Queue:    q,
		Webhook:  hook,
		Audit:    aud,
	}
}

// getUserFromContext extracts the authenticated user ID from the request context.
// Returns "anonymous" if no user is found in the context.
func getUserFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value(middleware.UserKey).(string); ok && userID != "" {
		return userID
	}
	return "anonymous"
}

// BaseCheck implements GET /v2/
func (h *Handler) BaseCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

// Catalog implements GET /v2/_catalog
func (h *Handler) Catalog(w http.ResponseWriter, r *http.Request) {
    // Extract User & Role
    userRole, _ := r.Context().Value(middleware.RoleKey).(string)
    var userID uuid.UUID
    
    userIDStr := getUserFromContext(r)
    if userIDStr != "anonymous" {
        if uid, err := uuid.Parse(userIDStr); err == nil {
            userID = uid
        }
    }
    
	repos, err := h.Metadata.GetRepositories(r.Context(), userID, userRole)
	if err != nil {
		http.Error(w, "Failed to list repositories", http.StatusInternalServerError)
		return
	}
	
	resp := struct {
		Repositories []string `json:"repositories"`
	}{
		Repositories: repos,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(resp)
}

// StartBlobUpload implements POST /v2/<name>/blobs/uploads/
func (h *Handler) StartBlobUpload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["name"]
	uploadID := uuid.New().String()

	fmt.Printf("Starting upload for repo: %s (UUID: %s)\n", repoName, uploadID)

	// location: /v2/<name>/blobs/uploads/<uuid>
	location := fmt.Sprintf("/v2/%s/blobs/uploads/%s", repoName, uploadID)

	w.Header().Set("Docker-Upload-UUID", uploadID)
	w.Header().Set("Location", location)
	w.Header().Set("Range", "0-0")
	w.WriteHeader(http.StatusAccepted)
}

// PatchBlobData implements PATCH /v2/<name>/blobs/uploads/<uuid>
func (h *Handler) PatchBlobData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["name"]
	uploadID := vars["uuid"]
	
	fmt.Printf("Patching blob for %s (UUID: %s)\n", repoName, uploadID)
	
	// Stream request body to temporary storage in MinIO
	// Path: uploads/<uuid>
	tempPath := path.Join("uploads", uploadID)
	
	writer, err := h.Storage.Writer(r.Context(), tempPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer writer.Close()
	
	// Copy data
	n, err := io.Copy(writer, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return updated location and range
	location := fmt.Sprintf("/v2/%s/blobs/uploads/%s", repoName, uploadID)
	w.Header().Set("Location", location)
	w.Header().Set("Docker-Upload-UUID", uploadID)
	w.Header().Set("Range", fmt.Sprintf("0-%d", n-1))
	w.WriteHeader(http.StatusAccepted)
}

// PutBlobUpload implements PUT /v2/<name>/blobs/uploads/<uuid>
func (h *Handler) PutBlobUpload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["name"]
	uploadID := vars["uuid"]
	digest := r.URL.Query().Get("digest")
	
	fmt.Printf("Finishing upload for %s (UUID: %s, Digest: %s)\n", repoName, uploadID, digest)
	
	if digest == "" {
		http.Error(w, "Digest required", http.StatusBadRequest)
		return
	}
	
	// In a real registry, we would concatenate chunks. 
	// For this MVP, we support Monolithic Upload (PUT with data) by writing directly to final path.
	// If it was a chunked upload, the data is in uploads/<uuid>, and we should move it.
	// We'll implementing a hybrid: Try to read body.
	
	blobPath := path.Join("blobs", digest)
	writer, err := h.Storage.Writer(r.Context(), blobPath)
	if err != nil {
		fmt.Printf("Storage writer failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer writer.Close()

	n, err := io.Copy(writer, r.Body)
	if err != nil {
		fmt.Printf("Blob write failed: %v\n", err)
		http.Error(w, "failed to write blob", http.StatusInternalServerError)
		return
	}
	
	fmt.Printf("Wrote blob %s (%d bytes)\n", digest, n)
	
    // Register Blob in DB
    // We don't know the exact media type at this stage (it's verified at manifest time), so generic.
    if err := h.Metadata.RegisterBlob(r.Context(), digest, n, "application/octet-stream"); err != nil {
        fmt.Printf("Failed to register blob metadata: %v\n", err)
        // Non-fatal, just stats will be off
    }

	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}

// CheckBlob implements HEAD /v2/<name>/blobs/<digest>
func (h *Handler) CheckBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["name"]
	digest := vars["digest"]
	
	blobPath := path.Join("blobs", digest)
	
	// Check if blob exists in storage
	reader, err := h.Storage.Reader(r.Context(), blobPath)
	if err != nil {
		fmt.Printf("Blob %s not found in storage for %s\n", digest, repoName)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	// Get blob size by reading to the end (or use Stat if available)
	// For now, we'll close the reader and trust the blob exists
	var blobSize int64
	if seeker, ok := reader.(io.ReadSeeker); ok {
		// If it's seekable, get size efficiently
		size, err := seeker.Seek(0, io.SeekEnd)
		if err == nil {
			blobSize = size
		}
		seeker.Seek(0, io.SeekStart)
	}
	reader.Close()
	
	// SELF-HEALING: Ensure blob is registered in database
	// This prevents scan failures when DB and storage are out of sync
	// Check if blob exists in DB, if not, register it
	exists, err := h.Metadata.BlobExists(r.Context(), digest)
	if err != nil {
		fmt.Printf("Failed to check blob existence in DB: %v\n", err)
	} else if !exists {
		// Blob exists in storage but not in DB - auto-register it
		fmt.Printf("[SELF-HEAL] Registering orphaned blob %s (size: %d)\n", digest, blobSize)
		if err := h.Metadata.RegisterBlob(r.Context(), digest, blobSize, "application/octet-stream"); err != nil {
			fmt.Printf("[SELF-HEAL] Failed to register blob %s: %v\n", digest, err)
		} else {
			fmt.Printf("[SELF-HEAL] Successfully registered blob %s\n", digest)
		}
	}
	
	// Return 200 OK with Content-Length if we have it
	if blobSize > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", blobSize))
	}
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

// GetBlob implements GET /v2/<name>/blobs/<digest>
func (h *Handler) GetBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	digest := vars["digest"]
	
	// Blob path: blobs/<digest>
	blobPath := path.Join("blobs", digest)
	
	reader, err := h.Storage.Reader(r.Context(), blobPath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	// SELF-HEALING: Ensure blob is registered in database before serving
	// This prevents scan failures when DB and storage are out of sync
	exists, err := h.Metadata.BlobExists(r.Context(), digest)
	if err != nil {
		fmt.Printf("Failed to check blob existence in DB: %v\n", err)
	} else if !exists {
		// Get blob size for registration
		var blobSize int64
		if seeker, ok := reader.(io.ReadSeeker); ok {
			size, err := seeker.Seek(0, io.SeekEnd)
			if err == nil {
				blobSize = size
			}
			seeker.Seek(0, io.SeekStart)
		}
		
		// Blob exists in storage but not in DB - auto-register it
		fmt.Printf("[SELF-HEAL] Registering orphaned blob %s (size: %d) during GET\n", digest, blobSize)
		if err := h.Metadata.RegisterBlob(r.Context(), digest, blobSize, "application/octet-stream"); err != nil {
			fmt.Printf("[SELF-HEAL] Failed to register blob %s: %v\n", digest, err)
		} else {
			fmt.Printf("[SELF-HEAL] Successfully registered blob %s\n", digest)
		}
	}
	
	defer reader.Close()
	
	w.Header().Set("Docker-Content-Digest", digest)
	// We should set Content-Type if known, usually application/octet-stream
	w.Header().Set("Content-Type", "application/octet-stream")
	
	if _, err := io.Copy(w, reader); err != nil {
		fmt.Printf("Failed to write blob %s: %v\n", digest, err)
	}
}

// PutManifest implements PUT /v2/<name>/manifests/<reference>
func (h *Handler) PutManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["name"]
	reference := vars["reference"]
	
	fmt.Printf("Put Manifest: %s:%s\n", repoName, reference)
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	
	if h.Config.EnableImmutableTags && !strings.HasPrefix(reference, "sha256:") {
		exists, err := h.Metadata.TagExists(r.Context(), repoName, reference)
		if err != nil {
			fmt.Printf("Tag check error: %v\n", err)
			http.Error(w, "internal check error", http.StatusInternalServerError)
			return
		}
		if exists {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"errors": [{"code": "TAG_INVALID", "message": "tag is immutable"}]}`))
			return
		}
	}

	manifestPath := path.Join("manifests", repoName, reference)
	writer, err := h.Storage.Writer(r.Context(), manifestPath)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	
	n, err := writer.Write(body)
	if err != nil {
		writer.Close()
		fmt.Printf("Failed to write manifest to storage: %v\n", err)
		http.Error(w, "storage write error", http.StatusInternalServerError)
		return
	}
	if n != len(body) {
		writer.Close()
		fmt.Printf("Incomplete write: wrote %d bytes, expected %d\n", n, len(body))
		http.Error(w, "storage write incomplete", http.StatusInternalServerError)
		return
	}
	
	if err := writer.Close(); err != nil {
		fmt.Printf("Failed to close writer: %v\n", err)
		http.Error(w, "storage close error", http.StatusInternalServerError)
		return
	}
	
	hash := sha256.Sum256(body)
	digest := "sha256:" + hex.EncodeToString(hash[:])
	
	digestPath := path.Join("manifests", repoName, digest)
	if digestPath != manifestPath {
		dWriter, err := h.Storage.Writer(r.Context(), digestPath)
		if err == nil {
			dWriter.Write(body)
			dWriter.Close()
		}
	}

	// --- Media Type Detection ---
	var manifestMap map[string]interface{}
	mediaType := "application/vnd.docker.distribution.manifest.v2+json" // Default
	
	if err := json.Unmarshal(body, &manifestMap); err == nil {
		if mt, ok := manifestMap["mediaType"].(string); ok && mt != "" {
			mediaType = mt
		} else if schemaVer, ok := manifestMap["schemaVersion"].(float64); ok && schemaVer == 1 {
			mediaType = "application/vnd.docker.distribution.manifest.v1+json"
		}
	}

	// --- Parsing for Stats ---
	var totalSize int64 = 0
	type Descriptor struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	}
	// V2 Struct
	type ManifestV2 struct {
		Config Descriptor   `json:"config"`
		Layers []Descriptor `json:"layers"`
	}
	
	isV2OrOCI := (mediaType == "application/vnd.docker.distribution.manifest.v2+json" || mediaType == "application/vnd.oci.image.manifest.v1+json")

	if isV2OrOCI {
		var m ManifestV2
		if err := json.Unmarshal(body, &m); err == nil {
			fmt.Printf("[DEBUG] PutManifest V2/OCI: Config Size=%d, Layers=%d\n", m.Config.Size, len(m.Layers))
			h.Metadata.RegisterBlob(r.Context(), m.Config.Digest, m.Config.Size, m.Config.MediaType)
			totalSize += m.Config.Size
			for _, layer := range m.Layers {
				h.Metadata.RegisterBlob(r.Context(), layer.Digest, layer.Size, layer.MediaType)
				totalSize += layer.Size
			}
		} else {
			fmt.Printf("[DEBUG] PutManifest V2/OCI Unmarshal Failed: %v\n", err)
		}
	} else {
		fmt.Printf("[DEBUG] PutManifest Media Type Mismatch: %s\n", mediaType)
		// V1 or Other - Fallback
		totalSize = int64(len(body)) 
	}
	
	if totalSize == 0 {
		totalSize = int64(len(body))
	}

	// --- Quota Check ---
	parts := strings.SplitN(repoName, "/", 2)
	nsName := "library" 
	if len(parts) == 2 {
		nsName = parts[0]
	}
	if err := h.Metadata.CheckQuota(r.Context(), nsName, totalSize); err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf(`{"errors": [{"code": "DENIED", "message": "quota exceeded: %v"}]}`, err)))
		return
	}

	// Extract User ID from context for namespace ownership
	var userID uuid.UUID
	userIDStr := getUserFromContext(r)
	fmt.Printf("[DEBUG] PutManifest: userIDStr=%s\n", userIDStr)
	if userIDStr != "anonymous" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			userID = uid
			fmt.Printf("[DEBUG] PutManifest: Parsed userID=%s\n", userID)
		} else {
			fmt.Printf("[DEBUG] PutManifest: Failed to parse userID: %v\n", err)
		}
	}

	manifestID, err := h.Metadata.RegisterManifest(r.Context(), repoName, reference, digest, totalSize, mediaType, userID)
	if err != nil {
		fmt.Printf("[ERROR] RegisterManifest failed: %v\n", err)
		http.Error(w, fmt.Sprintf("Metadata registration failed: %v", err), http.StatusInternalServerError)
		return
	}

	// --- Dependency Detection (V2/OCI Only) ---
	if isV2OrOCI {
		var m ManifestV2
		if err := json.Unmarshal(body, &m); err == nil && len(m.Layers) > 0 {
			layerDigests := make([]string, len(m.Layers))
			for i, l := range m.Layers {
				layerDigests[i] = l.Digest
			}
			h.Metadata.RegisterManifestLayers(r.Context(), manifestID, layerDigests)
			h.Metadata.DetectAndStoreDependencies(r.Context(), manifestID)
		}
	} else {
		fmt.Printf("Skipping dependency detection for %s (MediaType: %s)\n", manifestID, mediaType)
	}
	
	if h.Queue != nil {
		h.Queue.EnqueueScan(r.Context(), manifestID, repoName, reference)
	}

	if h.Webhook != nil {
		go h.Webhook.Notify(context.Background(), webhook.Event{
			Action: "push", Repository: repoName, Tag: reference, Digest: digest, Timestamp: time.Now(), User: getUserFromContext(r),
		})
	}

	if h.Audit != nil {
		userIDStr := getUserFromContext(r)
		if userIDStr != "anonymous" {
			if uid, err := uuid.Parse(userIDStr); err == nil {
				h.Audit.Log(r.Context(), uid, "PUSH", nil, map[string]interface{}{"repository": repoName, "tag": reference, "digest": digest, "size": totalSize})
			}
		}
	}
	
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}

// GetManifest implements GET /v2/<name>/manifests/<reference>
func (h *Handler) GetManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["name"]
	reference := vars["reference"]
	
	// 1. Resolve Manifest ID & Details to get correct Content-Type
	// We do this FIRST to set headers properly.
	manifestID, err := h.Metadata.GetManifestID(r.Context(), repoName, reference)
	if err != nil || manifestID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	// Fetch metadata
	mediaType := "application/vnd.docker.distribution.manifest.v2+json" // Default
	digest, _, mt, errDet := h.Metadata.GetManifestDetails(r.Context(), manifestID)
	if errDet == nil && mt != "" {
		mediaType = mt
	}

	// Fetch from storage
	manifestPath := path.Join("manifests", repoName, reference)
	// Try resolve path if needed (SmartResolve)
	_, errStat := h.Storage.Stat(r.Context(), manifestPath)
	if errStat != nil {
		// Try alternate
		altName := ""
		if strings.HasPrefix(repoName, "library/") { altName = strings.TrimPrefix(repoName, "library/") } else { altName = "library/" + repoName }
		altPath := path.Join("manifests", altName, reference)
		if _, errAlt := h.Storage.Stat(r.Context(), altPath); errAlt == nil {
			repoName = altName
			manifestPath = altPath
		}
	}
	
	reader, err := h.Storage.Reader(r.Context(), manifestPath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer reader.Close()
	
	w.Header().Set("Content-Type", mediaType)
	if digest != "" {
		w.Header().Set("Docker-Content-Digest", digest)
	}
	
	// --- Policy Enforcement ---
	// 1. Resolve Manifest UUID (Already done above for Content-Type)
	if err == nil {
		// Only enforce if we know the manifest (it exists in DB)
		
		// 2. Fetch Vulnerability Summary
		summary, err := h.Scanner.GetVulnerabilitySummary(r.Context(), manifestID)
		if err == nil {
			// 3. Check Signature (Cosign)
			var isSigned bool
			if digest, err := h.Metadata.GetDigest(r.Context(), manifestID); err == nil {
				// We have a digest, let's look for the .sig tag
				signed, _ := h.Metadata.HasSignature(r.Context(), repoName, digest)
				isSigned = signed
			}

			// 4. Evaluate Policy
			// Construct Input
			user := getUserFromContext(r)
			
			input := policy.EvaluationInput{
				Repository: repoName,
				Tag:        reference,
				User:       user,
				Environment: h.Config.PolicyEnvironment,
				Vulnerabilities: policy.VulnerabilitySummary{
					Critical: summary.Critical,
					High:     summary.High,
				},
				IsSigned: isSigned,
			}
			
			allowed, violations, err := h.Policy.Evaluate(r.Context(), input)
			if err != nil {
				log.Printf("Policy eval error: %v\n", err)
				// Open fail? or Fail closed? Let's fail open for errors to avoid blocking prod on bug.
			} else if !allowed {
				log.Printf("Policy DENIED pull for %s:%s. Violations: %v\n", repoName, reference, violations)
				
				// Return 403 Forbidden with OCI Error
				w.WriteHeader(http.StatusForbidden)
				jsonErrors := fmt.Sprintf(`{"errors": [{"code": "DENIED", "message": "policy violation: %s"}]}`, strings.Join(violations, "; "))
				w.Write([]byte(jsonErrors))
				return
			}
			
			// Policy passed (or fail-open on error) - Track Pull (Only on GET/Download)
			if r.Method == http.MethodGet {
				if err := h.Metadata.TrackPull(r.Context(), manifestID); err != nil {
					fmt.Printf("Failed to track pull for %s: %v\n", manifestID, err)
				}
			}
		}
	}

	manifestBytes, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, "Failed to read manifest", http.StatusInternalServerError)
		return
	}
	w.Write(manifestBytes)
}

// Tags implements GET /v2/<name>/tags/list
func (h *Handler) Tags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["name"]

	tags, err := h.Metadata.GetTags(r.Context(), repoName)
	if err != nil {
		// If repo not found, return 404
		if strings.Contains(err.Error(), "repository not found") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors": [{"code": "NAME_UNKNOWN", "message": "repository name not known to registry"}]}`))
			return
		}
		http.Error(w, "Failed to list tags", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}{
		Name: repoName,
		Tags: tags,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(resp)
}
