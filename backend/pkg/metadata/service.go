package metadata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"github.com/registryx/registryx/backend/pkg/health"
)

type Service struct {
	DB *sql.DB
}

type DependencyNode struct {
	ID     string `json:"id"`
	Type   string `json:"type"` // 'manifest'
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

type DependencyEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"` // 'bases-on'
}

type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

// EnsureRepository creates the namespace and repository if they don't exist.
// repoName is "namespace/repo" or just "repo" (library).
// userID is the owner of the namespace (for isolation).
func (s *Service) EnsureRepository(ctx context.Context, repoName string, userID uuid.UUID) (uuid.UUID, error) {
	parts := strings.SplitN(repoName, "/", 2)
	nsName := "library"
	rName := repoName
	if len(parts) == 2 {
		nsName = parts[0]
		rName = parts[1]
	}

	// 1. Ensure Namespace
	var nsID uuid.UUID
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO namespaces (name) VALUES ($1) 
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, nsName).Scan(&nsID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// 2. Ensure Repository with Owner
	var repoID uuid.UUID
	err = s.DB.QueryRowContext(ctx, `
		INSERT INTO repositories (namespace_id, name, owner_id) VALUES ($1, $2, $3)
		ON CONFLICT (namespace_id, name, owner_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id`, nsID, rName, userID).Scan(&repoID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to ensure repository: %w", err)
	}

	return repoID, nil
}

// RegisterManifest records the manifest and tag in the DB.
func (s *Service) RegisterManifest(ctx context.Context, repoName, reference, digest string, size int64, mediaType string, userID uuid.UUID) (uuid.UUID, error) {
	repoID, err := s.EnsureRepository(ctx, repoName, userID)
	if err != nil {
		return uuid.Nil, err
	}

	// 1. Insert Manifest
	var manifestID uuid.UUID
	err = s.DB.QueryRowContext(ctx, `
		INSERT INTO manifests (repository_id, digest, size, media_type)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (repository_id, digest) DO UPDATE SET digest = EXCLUDED.digest
		RETURNING id`, repoID, digest, size, mediaType).Scan(&manifestID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert manifest: %w", err)
	}

	// 2. If 'reference' is a tag (not a digest), update the Tag table
	if !strings.HasPrefix(reference, "sha256:") {
		_, err = s.DB.ExecContext(ctx, `
			INSERT INTO tags (repository_id, manifest_id, name)
			VALUES ($1, $2, $3)
			ON CONFLICT (repository_id, name) DO UPDATE SET manifest_id = EXCLUDED.manifest_id, updated_at = CURRENT_TIMESTAMP`,
			repoID, manifestID, reference)
		if err != nil {
			return manifestID, fmt.Errorf("failed to update tag: %w", err)
		}
	}

	return manifestID, nil
}

// TrackPull updates the pull count and last pulled time for a manifest
func (s *Service) TrackPull(ctx context.Context, manifestID uuid.UUID) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE manifests 
		SET pull_count = COALESCE(pull_count, 0) + 1, 
		    last_pulled_at = CURRENT_TIMESTAMP 
		WHERE id = $1`, manifestID)
	return err
}

// GetManifestID resolves a repository and reference (tag or digest) to a Manifest UUID.
func (s *Service) GetManifestID(ctx context.Context, repoName, reference string) (uuid.UUID, error) {
	// 1. Get Repo ID
	var repoID uuid.UUID
	parts := strings.SplitN(repoName, "/", 2)
	nsName := "library"
	rName := repoName
	if len(parts) == 2 {
		nsName = parts[0]
		rName = parts[1]
	}

	err := s.DB.QueryRowContext(ctx, `
		SELECT r.id FROM repositories r
		JOIN namespaces n ON r.namespace_id = n.id
		WHERE n.name = $1 AND r.name = $2`, nsName, rName).Scan(&repoID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository not found")
	}

	// 2. Get Manifest ID
	var manifestID uuid.UUID
	
	if strings.HasPrefix(reference, "sha256:") {
		// By Digest
		err = s.DB.QueryRowContext(ctx, `
			SELECT id FROM manifests WHERE repository_id = $1 AND digest = $2`, 
			repoID, reference).Scan(&manifestID)
	} else {
		// By Tag
		err = s.DB.QueryRowContext(ctx, `
			SELECT manifest_id FROM tags WHERE repository_id = $1 AND name = $2`, 
			repoID, reference).Scan(&manifestID)
	}

	if err != nil {
		return uuid.Nil, fmt.Errorf("manifest not found")
	}

	return manifestID, nil
}

// GetRepositories returns a list of all repository names, filtered by user.
func (s *Service) GetRepositories(ctx context.Context, userID uuid.UUID, role string) ([]string, error) {
    whereClause := "1=1"
    args := []interface{}{}
    if role != "admin" {
        whereClause = "r.owner_id = $1"
        args = append(args, userID)
    }

	query := fmt.Sprintf(`
		SELECT n.name || '/' || r.name 
		FROM repositories r
		JOIN namespaces n ON r.namespace_id = n.id
        WHERE %s`, whereClause)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		repos = append(repos, name)
	}
	return repos, nil
}

// GetDigest retrieves the digest for a manifest UUID.
func (s *Service) GetDigest(ctx context.Context, manifestID uuid.UUID) (string, error) {
	var digest string
	err := s.DB.QueryRowContext(ctx, "SELECT digest FROM manifests WHERE id = $1", manifestID).Scan(&digest)
	return digest, err
}

// GetManifestDetails retrieves digest, size, and media_type for a manifest UUID.
func (s *Service) GetManifestDetails(ctx context.Context, manifestID uuid.UUID) (string, int64, string, error) {
	var digest, mediaType string
	var size int64
	err := s.DB.QueryRowContext(ctx, "SELECT digest, size, media_type FROM manifests WHERE id = $1", manifestID).Scan(&digest, &size, &mediaType)
	return digest, size, mediaType, err
}

// HasSignature checks if a manifest has a corresponding Cosign signature tag.
// format: sha256-<digest>.sig
func (s *Service) HasSignature(ctx context.Context, repoName string, digest string) (bool, error) {
	if !strings.HasPrefix(digest, "sha256:") {
		return false, nil // Only supporting sha256 for now
	}
	
	// Cosign format: sha256:hash -> sha256-hash.sig
	sigTag := strings.Replace(digest, "sha256:", "sha256-", 1) + ".sig"
	
	return s.TagExists(ctx, repoName, sigTag)
}

// TagExists checks if a specific tag exists for a repository.
func (s *Service) TagExists(ctx context.Context, repoName, tagName string) (bool, error) {
	_, err := s.GetManifestID(ctx, repoName, tagName)
	if err == nil {
		return true, nil
	}
	// Simple error string checking for now. Ideally should use custom error types.
	if strings.Contains(err.Error(), "not found") {
		return false, nil
	}
	return false, err
}

// GetTags returns all tags for a given repository.
func (s *Service) GetTags(ctx context.Context, repoName string) ([]string, error) {
	// 1. Get Repo ID
	var repoID uuid.UUID
	parts := strings.SplitN(repoName, "/", 2)
	nsName := "library"
	rName := repoName
	if len(parts) == 2 {
		nsName = parts[0]
		rName = parts[1]
	}

	err := s.DB.QueryRowContext(ctx, `
		SELECT r.id FROM repositories r
		JOIN namespaces n ON r.namespace_id = n.id
		WHERE n.name = $1 AND r.name = $2`, nsName, rName).Scan(&repoID)
	if err != nil {
		return nil, fmt.Errorf("repository not found")
	}

	// 2. Get Tags
	rows, err := s.DB.QueryContext(ctx, `SELECT name FROM tags WHERE repository_id = $1`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tags = append(tags, name)
	}

	return tags, nil
}

// DeleteRepository deletes a repository and all associated tags and manifests
func (s *Service) DeleteRepository(ctx context.Context, repoName string) error {
	// Parse namespace and repo name
	parts := strings.SplitN(repoName, "/", 2)
	nsName := "library"
	rName := repoName
	if len(parts) == 2 {
		nsName = parts[0]
		rName = parts[1]
	}

	// Get repository ID
	var repoID uuid.UUID
	err := s.DB.QueryRowContext(ctx, `
		SELECT r.id FROM repositories r
		JOIN namespaces n ON r.namespace_id = n.id
		WHERE n.name = $1 AND r.name = $2`, nsName, rName).Scan(&repoID)
	if err != nil {
		fmt.Printf("DeleteRepository: Repo not found for %s/%s\n", nsName, rName)
		return fmt.Errorf("repository not found")
	}

	fmt.Printf("DeleteRepository: Found ID %s for %s/%s. Deleting...\n", repoID, nsName, rName)

	// Delete tags (CASCADE will handle manifests via foreign key)
	_, err = s.DB.ExecContext(ctx, `DELETE FROM tags WHERE repository_id = $1`, repoID)
	if err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// Delete manifests
	_, err = s.DB.ExecContext(ctx, `DELETE FROM manifests WHERE repository_id = $1`, repoID)
	if err != nil {
		return fmt.Errorf("failed to delete manifests: %w", err)
	}

	// Delete repository
	res, err := s.DB.ExecContext(ctx, `DELETE FROM repositories WHERE id = $1`, repoID)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}
	
	rows, _ := res.RowsAffected()
	fmt.Printf("DeleteRepository: Deleted ID %s. Rows affected: %d\n", repoID, rows)

	return nil
}

// DeleteTag deletes a specific tag from a repository
func (s *Service) DeleteTag(ctx context.Context, repoName, tagName string) error {
	// Parse namespace and repo name
	parts := strings.SplitN(repoName, "/", 2)
	nsName := "library"
	rName := repoName
	if len(parts) == 2 {
		nsName = parts[0]
		rName = parts[1]
	}

	// Get repository ID
	var repoID uuid.UUID
	err := s.DB.QueryRowContext(ctx, `
		SELECT r.id FROM repositories r
		JOIN namespaces n ON r.namespace_id = n.id
		WHERE n.name = $1 AND r.name = $2`, nsName, rName).Scan(&repoID)
	if err != nil {
		return fmt.Errorf("repository not found")
	}

	// Delete the tag
	result, err := s.DB.ExecContext(ctx, `DELETE FROM tags WHERE repository_id = $1 AND name = $2`, repoID, tagName)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

// DeleteManifest deletes a manifest by ID
func (s *Service) DeleteManifest(ctx context.Context, id uuid.UUID) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM manifests WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("manifest not found")
	}
	return nil
}

// RegisterBlob records a blob in the DB
func (s *Service) RegisterBlob(ctx context.Context, digest string, size int64, mediaType string) error {
    _, err := s.DB.ExecContext(ctx, `
        INSERT INTO blobs (digest, size, media_type)
        VALUES ($1, $2, $3)
        ON CONFLICT (digest) DO NOTHING`,
        digest, size, mediaType)
    return err
}

// BlobExists checks if a blob is registered in the database
func (s *Service) BlobExists(ctx context.Context, digest string) (bool, error) {
    var exists bool
    err := s.DB.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM blobs WHERE digest = $1)`,
        digest).Scan(&exists)
    return exists, err
}


type DashboardStats struct {
    Repositories    int
    Images          int
    Vulnerabilities int
    StorageBytes    int64
    Severity        SeverityBreakdown
    RecentPushes    []PushEvent
}

type SeverityBreakdown struct {
    Critical int `json:"critical"`
    High     int `json:"high"`
    Medium   int `json:"medium"`
    Low      int `json:"low"`
}

type PushEvent struct {
    Repository string    `json:"repository"`
    Tag        string    `json:"tag"`
    Digest     string    `json:"digest"`
    PushedAt   time.Time `json:"pushedAt"`
}

// OrphanBlob represents a blob that is not referenced by any manifest.
type OrphanBlob struct {
	Digest string
	Size   int64
}

// GetOrphanedBlobs returns a list of blobs that are not referenced by any manifest or manifest_layers.
// This is used for Garbage Collection.
func (s *Service) GetOrphanedBlobs(ctx context.Context) ([]OrphanBlob, error) {
	// Find blobs that are NOT in manifest_layers AND NOT a config digest in manifests
	query := `
		SELECT b.digest, b.size
		FROM blobs b
		LEFT JOIN manifest_layers ml ON b.digest = ml.blob_digest
		LEFT JOIN manifests m ON (m.config_digest = b.digest)
		WHERE ml.blob_digest IS NULL AND m.config_digest IS NULL
	`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orphaned blobs: %w", err)
	}
	defer rows.Close()

	var orphans []OrphanBlob
	for rows.Next() {
		var o OrphanBlob
		if err := rows.Scan(&o.Digest, &o.Size); err != nil {
			return nil, err
		}
		orphans = append(orphans, o)
	}
	return orphans, nil
}

// DeleteBlob removes a blob from the database.
func (s *Service) DeleteBlob(ctx context.Context, digest string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM blobs WHERE digest = $1", digest)
	return err
}

// CalculateAndStoreHealthScore calculates the health score for a manifest and stores it
func (s *Service) CalculateAndStoreHealthScore(ctx context.Context, manifestID uuid.UUID) (*health.HealthScore, error) {
	fmt.Printf("[Health] Calculating score for manifest %s\n", manifestID)
	// Gather metrics needed for health calculation
	metrics, err := s.getImageMetrics(ctx, manifestID)
	if err != nil {
		fmt.Printf("[Health] Failed to get metrics for %s: %v\n", manifestID, err)
		return nil, fmt.Errorf("failed to get image metrics: %w", err)
	}

	// Calculate health score
	scorer := health.NewScorer()
	score := scorer.CalculateHealthScore(metrics)
	fmt.Printf("[Health] Score for %s: Overall=%d, Grade=%s\n", manifestID, score.Overall, score.Grade)

	// Store in database
	res, err := s.DB.ExecContext(ctx, `
		UPDATE manifests 
		SET health_score = $1, 
		    health_grade = $2,
		    health_security = $3,
		    health_freshness = $4,
		    health_efficiency = $5,
		    health_maintenance = $6,
		    last_health_check = $7
		WHERE id = $8`,
		score.Overall, score.Grade, score.Security, score.Freshness,
		score.Efficiency, score.Maintenance, score.LastUpdated, manifestID)

	if err != nil {
		fmt.Printf("[Health] Failed to store score for %s: %v\n", manifestID, err)
		return nil, fmt.Errorf("failed to store health score: %w", err)
	}

	rows, _ := res.RowsAffected()
	fmt.Printf("[Health] DB Update for %s: rows affected = %d\n", manifestID, rows)

	return score, nil
}

// getImageMetrics gathers all metrics needed for health score calculation
func (s *Service) getImageMetrics(ctx context.Context, manifestID uuid.UUID) (*health.ImageMetrics, error) {
	var metrics health.ImageMetrics
	metrics.ManifestID = manifestID

	// Get basic manifest info and vulnerability counts
	var createdAt, lastPushedAt time.Time
	err := s.DB.QueryRowContext(ctx, `
		WITH latest_report AS (
			SELECT manifest_id, critical_count, high_count, medium_count, low_count
			FROM vulnerability_reports
			WHERE status = 'completed'
			ORDER BY scanned_at DESC
			LIMIT 1
		)
		SELECT 
			m.size,
			m.created_at,
			COALESCE(MAX(t.created_at), m.created_at) as last_pushed,
			COALESCE(vr.critical_count, 0) as critical_vulns,
			COALESCE(vr.high_count, 0) as high_vulns,
			COALESCE(vr.medium_count, 0) as medium_vulns,
			COALESCE(vr.low_count, 0) as low_vulns
		FROM manifests m
		LEFT JOIN tags t ON t.manifest_id = m.id
		LEFT JOIN latest_report vr ON vr.manifest_id = m.id
		WHERE m.id = $1
		GROUP BY m.id, m.size, m.created_at, vr.critical_count, vr.high_count, vr.medium_count, vr.low_count`,
		manifestID).Scan(
		&metrics.ImageSizeBytes,
		&createdAt,
		&lastPushedAt,
		&metrics.CriticalVulns,
		&metrics.HighVulns,
		&metrics.MediumVulns,
		&metrics.LowVulns,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query manifest metrics: %w", err)
	}

	metrics.CreatedAt = createdAt
	metrics.LastPushedAt = lastPushedAt
	metrics.TotalVulns = metrics.CriticalVulns + metrics.HighVulns + metrics.MediumVulns + metrics.LowVulns

	// Get pull count (estimate from tag count for now)
	err = s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tags WHERE manifest_id = $1`,
		manifestID).Scan(&metrics.PullCount)
	if err != nil {
		metrics.PullCount = 0 // Default if query fails
	}

	// Get average size in repository
	err = s.DB.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(m2.size), 0)
		FROM manifests m1
		JOIN manifests m2 ON m2.repository_id = m1.repository_id
		WHERE m1.id = $1`,
		manifestID).Scan(&metrics.AverageSizeInRepo)
	if err != nil {
		metrics.AverageSizeInRepo = metrics.ImageSizeBytes // Use own size as fallback
	}

	return &metrics, nil
}

// GetHealthScore retrieves the stored health score for a manifest
func (s *Service) GetHealthScore(ctx context.Context, manifestID uuid.UUID) (*health.HealthScore, error) {
	var score health.HealthScore
	var lastUpdated sql.NullTime

	err := s.DB.QueryRowContext(ctx, `
		SELECT health_score, health_grade, health_security, health_freshness,
		       health_efficiency, health_maintenance, last_health_check
		FROM manifests
		WHERE id = $1`,
		manifestID).Scan(
		&score.Overall, &score.Grade, &score.Security, &score.Freshness,
		&score.Efficiency, &score.Maintenance, &lastUpdated)

	if err != nil {
		return nil, fmt.Errorf("failed to get health score: %w", err)
	}

	if lastUpdated.Valid {
		score.LastUpdated = lastUpdated.Time
	}
	
	// Calculate trend by comparing with previous score
	previousScore, err := s.GetPreviousHealthScore(ctx, manifestID)
	if err == nil && previousScore != nil {
		if score.Overall > previousScore.Overall {
			score.Trend = "improving"
		} else if score.Overall < previousScore.Overall {
			score.Trend = "declining"
		} else {
			score.Trend = "stable"
		}
	} else {
		score.Trend = "stable" // Default to stable if no previous score
	}

	return &score, nil
}

// GetPreviousHealthScore retrieves the second-most recent health score for a manifest
// This is used for trend calculation by comparing current score with previous score
func (s *Service) GetPreviousHealthScore(ctx context.Context, manifestID uuid.UUID) (*health.HealthScore, error) {
	// Query the health_score_history table for the previous score
	// If history doesn't exist, this will return an error which is handled by the caller
	var score health.HealthScore
	var recordedAt sql.NullTime

	err := s.DB.QueryRowContext(ctx, `
		SELECT health_score, health_grade, health_security, health_freshness,
		       health_efficiency, health_maintenance, recorded_at
		FROM health_score_history
		WHERE manifest_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1 OFFSET 1`,
		manifestID).Scan(
		&score.Overall, &score.Grade, &score.Security, &score.Freshness,
		&score.Efficiency, &score.Maintenance, &recordedAt)

	if err != nil {
		return nil, fmt.Errorf("no previous health score found: %w", err)
	}

	if recordedAt.Valid {
		score.LastUpdated = recordedAt.Time
	}

	return &score, nil
}

// GetDashboardStats calculates real-time stats, filtered by user
func (s *Service) GetDashboardStats(ctx context.Context, userID uuid.UUID, role string) (*DashboardStats, error) {
    stats := &DashboardStats{}

    // Isolation Clause
    whereNamespace := "1=1"
    args := []interface{}{}
    
    if role != "admin" {
        whereNamespace = "r.owner_id = $1"
        args = append(args, userID)
    }

    // 1. Count Repositories
    repoQuery := fmt.Sprintf("SELECT COUNT(*) FROM repositories r JOIN namespaces n ON r.namespace_id = n.id WHERE %s", whereNamespace)
    err := s.DB.QueryRowContext(ctx, repoQuery, args...).Scan(&stats.Repositories)
    if err != nil { return nil, err }

    // 2. Count Images (Manifests)
    manifestQuery := fmt.Sprintf("SELECT COUNT(*) FROM manifests JOIN repositories r ON manifests.repository_id = r.id JOIN namespaces n ON r.namespace_id = n.id WHERE %s", whereNamespace)
    err = s.DB.QueryRowContext(ctx, manifestQuery, args...).Scan(&stats.Images)
    if err != nil { return nil, err }

    // 3. Sum Vulnerabilities & Severity (Only counting latest report per manifest)
    // Filter by manifest ownership
    vulnQuery := fmt.Sprintf(`
        SELECT 
            COALESCE(SUM(critical_count + high_count + medium_count + low_count), 0),
            COALESCE(SUM(critical_count), 0),
            COALESCE(SUM(high_count), 0),
            COALESCE(SUM(medium_count), 0),
            COALESCE(SUM(low_count), 0)
        FROM (
            SELECT DISTINCT ON (vr.manifest_id) 
                vr.critical_count, vr.high_count, vr.medium_count, vr.low_count
            FROM vulnerability_reports vr
            JOIN manifests m ON vr.manifest_id = m.id
            JOIN repositories r ON m.repository_id = r.id
            JOIN namespaces n ON r.namespace_id = n.id
            WHERE vr.status = 'completed' AND %s
            ORDER BY vr.manifest_id, vr.scanned_at DESC
        ) latest_reports`, whereNamespace)

    err = s.DB.QueryRowContext(ctx, vulnQuery, args...).Scan(
            &stats.Vulnerabilities,
            &stats.Severity.Critical,
            &stats.Severity.High,
            &stats.Severity.Medium,
            &stats.Severity.Low,
        )
    if err != nil { return nil, err }

    // 4. Sum Storage (Blobs) - Calculate total image size (Layers)
    // We sum the size of all blobs (layers) associated with the user's manifests.
    // Note: This counts shared blobs multiple times (once per manifest), which is 
    // correct for "Usage" perspective (User A uses 50MB, User B uses 50MB).
    storageQuery := fmt.Sprintf(`
        SELECT COALESCE(SUM(b.size), 0)
        FROM manifests m
        JOIN repositories r ON m.repository_id = r.id
        JOIN namespaces n ON r.namespace_id = n.id
        JOIN manifest_layers ml ON m.id = ml.manifest_id
        JOIN blobs b ON ml.blob_digest = b.digest
        WHERE %s`, whereNamespace)
    err = s.DB.QueryRowContext(ctx, storageQuery, args...).Scan(&stats.StorageBytes)
    if err != nil { return nil, err }

    // 5. Recent Pushes (Last 5 manifests)
    pushesQuery := fmt.Sprintf(`
        SELECT r.name, COALESCE(t.name, 'latest'), m.digest, m.created_at
        FROM manifests m
        JOIN repositories r ON m.repository_id = r.id
        JOIN namespaces n ON r.namespace_id = n.id
        LEFT JOIN tags t ON t.manifest_id = m.id
        WHERE %s
        ORDER BY m.created_at DESC
        LIMIT 5`, whereNamespace)


    rows, err := s.DB.QueryContext(ctx, pushesQuery, args...)
    if err != nil { return nil, err }
    defer rows.Close()

    for rows.Next() {
        var p PushEvent
		var tagName sql.NullString // Handle null tags
        if err := rows.Scan(&p.Repository, &tagName, &p.Digest, &p.PushedAt); err == nil {
			if tagName.Valid {
				p.Tag = tagName.String
			} else {
				p.Tag = "untagged"
			}
            stats.RecentPushes = append(stats.RecentPushes, p)
        }
    }

    return stats, nil
}

// RegisterManifestLayers links blobs as layers to a manifest
func (s *Service) RegisterManifestLayers(ctx context.Context, manifestID uuid.UUID, layers []string) error {
	// 1. Delete existing layers if any (to handle re-upload)
	_, err := s.DB.ExecContext(ctx, "DELETE FROM manifest_layers WHERE manifest_id = $1", manifestID)
	if err != nil {
		return err
	}

	// 2. Insert new layers
	for i, digest := range layers {
		_, err := s.DB.ExecContext(ctx, `
			INSERT INTO manifest_layers (manifest_id, blob_digest, position)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING`, manifestID, digest, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// DetectAndStoreDependencies finds the parent manifest based on shared layer prefix
func (s *Service) DetectAndStoreDependencies(ctx context.Context, manifestID uuid.UUID) error {
	fmt.Printf("[Dep] Detecting dependencies for manifest %s\n", manifestID)
	// 1. Find parent using the prefix query
	// Potential parent is a manifest that has a subset of this manifest's layers at the exact same positions
	var parentID uuid.UUID
	err := s.DB.QueryRowContext(ctx, `
        SELECT p.id
        FROM manifests p
        JOIN (
            SELECT manifest_id, count(*) as layer_count
            FROM manifest_layers
            GROUP BY manifest_id
        ) p_counts ON p.id = p_counts.manifest_id
        WHERE p.id != $1
        AND p_counts.layer_count < (SELECT count(*) FROM manifest_layers WHERE manifest_id = $1)
        AND NOT EXISTS (
            -- All layers of parent P must exist in child M1 at the same position
            SELECT 1 
            FROM manifest_layers pl
            WHERE pl.manifest_id = p.id
            AND NOT EXISTS (
                SELECT 1 
                FROM manifest_layers cl
                WHERE cl.manifest_id = $1
                AND cl.blob_digest = pl.blob_digest
                AND cl.position = pl.position
            )
        )
        ORDER BY p_counts.layer_count DESC
        LIMIT 1`, manifestID).Scan(&parentID)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("[Dep] No parent found for %s\n", manifestID)
			return nil 
		}
		return fmt.Errorf("failed to detect parent manifest: %w", err)
	}

	fmt.Printf("[Dep] Found parent %s for CHILD %s\n", parentID, manifestID)

	// 2. Store relationship
	_, err = s.DB.ExecContext(ctx, `
        INSERT INTO image_dependencies (manifest_id, parent_manifest_id)
        VALUES ($1, $2)
        ON CONFLICT (manifest_id, parent_manifest_id) DO NOTHING`,
		manifestID, parentID)

	return err
}

// GetDependencyGraph returns a graph representation of image relationships
func (s *Service) GetDependencyGraph(ctx context.Context, repoName string, userID uuid.UUID, role string) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes: []DependencyNode{},
		Edges: []DependencyEdge{},
	}

    // Filter Logic
    whereClause := "1=1"
    args := []interface{}{}
    
    // User Isolation: Users can only see dependencies where THEY own the Child image.
    // They can see parents (base images) even if public, as long as it links to their child.
    // (Or we can restrict entirely, but usually you want to see "My App depends on Alpine")
    if role != "admin" {
        whereClause = "r.owner_id = $1"
        args = append(args, userID)
    }

	// For now, get all dependencies to build a global map
	// In production, we might filter by repoName if provided
    // We add JOIN namespaces n ON r.namespace_id = n.id
	query := fmt.Sprintf(`
        SELECT DISTINCT
            m.id, r.name, COALESCE(t.name, 'latest'), m.digest,
            pm.id, pr.name, COALESCE(pt.name, 'latest'), pm.digest
        FROM image_dependencies id
        JOIN manifests m ON id.manifest_id = m.id
        JOIN repositories r ON m.repository_id = r.id
        JOIN namespaces n ON r.namespace_id = n.id
        LEFT JOIN tags t ON t.manifest_id = m.id
        JOIN manifests pm ON id.parent_manifest_id = pm.id
        JOIN repositories pr ON pm.repository_id = pr.id
        LEFT JOIN tags pt ON pt.manifest_id = pm.id
        WHERE %s
    `, whereClause)
	
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeMap := make(map[string]bool)

	for rows.Next() {
		var mID, rName, tName, mDigest string
		var pmID, prName, ptName, pmDigest string

		if err := rows.Scan(&mID, &rName, &tName, &mDigest, &pmID, &prName, &ptName, &pmDigest); err != nil {
			continue
		}

		// Add child node
		if !nodeMap[mID] {
			graph.Nodes = append(graph.Nodes, DependencyNode{
				ID: mID, Type: "manifest", Name: rName, Tag: tName, Digest: mDigest,
			})
			nodeMap[mID] = true
		}

		// Add parent node
		if !nodeMap[pmID] {
			graph.Nodes = append(graph.Nodes, DependencyNode{
				ID: pmID, Type: "manifest", Name: prName, Tag: ptName, Digest: pmDigest,
			})
			nodeMap[pmID] = true
		}

		// Add edge (Child -> Parent, meaning "Bases On")
		graph.Edges = append(graph.Edges, DependencyEdge{
			Source: mID,
			Target: pmID,
			Label:  "bases-on",
		})
	}

	return graph, nil
}

// GetNamespaceUsage calculates current storage usage and returns quota for a namespace
func (s *Service) GetNamespaceUsage(ctx context.Context, nsName string) (int64, int64, error) {
	var nsID uuid.UUID
	var quota int64
	err := s.DB.QueryRowContext(ctx, "SELECT id, quota_bytes FROM namespaces WHERE name = $1", nsName).Scan(&nsID, &quota)
	if err != nil {
		// Namespace doesn't exist yet - return default quota (10GB)
		return 0, 10*1024*1024*1024, nil
	}

	// Calculate Storage Usage (Deduplicated within namespace)
	// Includes Blobs from Manifests + Config Blobs
	query := `
	WITH ns_manifests AS (
		SELECT m.id, m.config_digest 
		FROM manifests m 
		JOIN repositories r ON m.repository_id = r.id 
		WHERE r.namespace_id = $1
	),
	ns_blobs AS (
		SELECT ml.blob_digest AS digest
		FROM manifest_layers ml
		JOIN ns_manifests nm ON ml.manifest_id = nm.id
		UNION
		SELECT config_digest AS digest FROM ns_manifests
	)
	SELECT COALESCE(SUM(b.size), 0)
	FROM blobs b
	JOIN ns_blobs nsb ON b.digest = nsb.digest
	`
	
	var usage int64
	err = s.DB.QueryRowContext(ctx, query, nsID).Scan(&usage)
	if err != nil {
		return 0, quota, err
	}
	
	return usage, quota, nil
}

// CheckQuota checks if adding newBytes would exceed quota
func (s *Service) CheckQuota(ctx context.Context, nsName string, newBytes int64) error {
	usage, quota, err := s.GetNamespaceUsage(ctx, nsName)
	if err != nil {
		return err
	}
	
	if (usage + newBytes) > quota {
		return fmt.Errorf("storage quota exceeded: used %d/%d bytes", usage, quota)
	}
	return nil
}

// DeleteUntaggedManifests deletes manifests that have no tags pointing to them.
func (s *Service) DeleteUntaggedManifests(ctx context.Context) (int64, error) {
	// Delete manifests that are NOT tagged and NOT used as a parent by another image
	query := `
		DELETE FROM manifests 
		WHERE id NOT IN (SELECT manifest_id FROM tags)
		AND id NOT IN (SELECT parent_manifest_id FROM image_dependencies)
	`
	res, err := s.DB.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
