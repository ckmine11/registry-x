package costs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles cost calculation and optimization
type Service struct {
	DB     *sql.DB
	Config *CostConfig
}

// CostConfig holds pricing configuration
type CostConfig struct {
	StorageCostPerGBMonth float64 // e.g., $0.023 for S3 Standard
	BandwidthCostPerGB    float64 // e.g., $0.09 for S3 egress
	RegistryRegion        string
}

// ImageCost represents the cost breakdown for an image
type ImageCost struct {
	ManifestID       uuid.UUID  `json:"manifest_id"`
	Repository       string     `json:"repository"`
	Tag              string     `json:"tag"`
	SizeBytes        int64      `json:"size_bytes"`
	StorageCostUSD   float64    `json:"storage_cost_usd"`
	BandwidthCostUSD float64    `json:"bandwidth_cost_usd"`
	TotalCostUSD     float64    `json:"total_cost_usd"`
	PullCount30d     int        `json:"pull_count_30d"`
	LastPulledAt     *time.Time `json:"last_pulled_at,omitempty"`
	CostPerPull      float64    `json:"cost_per_pull"`
}

// ZombieImage represents an unused image
type ZombieImage struct {
	ManifestID          uuid.UUID `json:"manifest_id"`
	Repository          string    `json:"repository"`
	Tag                 string    `json:"tag"`
	DaysSinceLastPull   int       `json:"days_since_last_pull"`
	StorageCostUSD      float64   `json:"storage_cost_usd"`
	RecommendedAction   string    `json:"recommended_action"`
}

// CostDashboard represents the overall cost summary
type CostDashboard struct {
	TotalStorageCostUSD   float64     `json:"total_storage_cost_usd"`
	TotalBandwidthCostUSD float64     `json:"total_bandwidth_cost_usd"`
	TotalCostUSD          float64     `json:"total_cost_usd"`
	TotalImages           int         `json:"total_images"`
	ZombieImages          int         `json:"zombie_images"`
	PotentialSavingsUSD   float64     `json:"potential_savings_usd"`
	TopExpensiveImages    []ImageCost `json:"top_expensive_images"`
	CostTrend             string      `json:"cost_trend"`
}

// NewService creates a new cost service
func NewService(db *sql.DB, config *CostConfig) *Service {
	if config == nil {
		// Default to S3 Standard pricing (us-east-1)
		config = &CostConfig{
			StorageCostPerGBMonth: 0.023,
			BandwidthCostPerGB:    0.09,
			RegistryRegion:        "us-east-1",
		}
	}
	return &Service{
		DB:     db,
		Config: config,
	}
}

// CalculateImageCost calculates the cost for a single image
func (s *Service) CalculateImageCost(sizeBytes int64, pullCount int) ImageCost {
	sizeGB := float64(sizeBytes) / 1e9
	
	storageCost := sizeGB * s.Config.StorageCostPerGBMonth
	bandwidthCost := sizeGB * float64(pullCount) * s.Config.BandwidthCostPerGB
	totalCost := storageCost + bandwidthCost
	
	costPerPull := 0.0
	if pullCount > 0 {
		costPerPull = totalCost / float64(pullCount)
	}
	
	return ImageCost{
		SizeBytes:        sizeBytes,
		StorageCostUSD:   storageCost,
		BandwidthCostUSD: bandwidthCost,
		TotalCostUSD:     totalCost,
		PullCount30d:     pullCount,
		CostPerPull:      costPerPull,
	}
}

// RefreshAllCosts recalculates costs for all images
func (s *Service) RefreshAllCosts(ctx context.Context) error {
	fmt.Println("[Costs] Refreshing cost data for all images...")
	
	rows, err := s.DB.QueryContext(ctx, `
		SELECT m.id, m.size, COALESCE(m.pull_count, 0), m.last_pulled_at
		FROM manifests m
	`)
	if err != nil {
		return fmt.Errorf("failed to query manifests: %w", err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var manifestID uuid.UUID
		var size int64
		var pullCount int
		var lastPulled sql.NullTime
		
		if err := rows.Scan(&manifestID, &size, &pullCount, &lastPulled); err != nil {
			continue
		}
		
		cost := s.CalculateImageCost(size, pullCount)
		
		// Store in database
		_, err := s.DB.ExecContext(ctx, `
			INSERT INTO storage_costs (
				manifest_id, blob_size_bytes, storage_cost_usd, 
				bandwidth_cost_usd, total_cost_usd, pull_count_30d,
				last_pulled_at, cost_per_pull, calculated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (manifest_id) DO UPDATE SET
				blob_size_bytes = EXCLUDED.blob_size_bytes,
				storage_cost_usd = EXCLUDED.storage_cost_usd,
				bandwidth_cost_usd = EXCLUDED.bandwidth_cost_usd,
				total_cost_usd = EXCLUDED.total_cost_usd,
				pull_count_30d = EXCLUDED.pull_count_30d,
				last_pulled_at = EXCLUDED.last_pulled_at,
				cost_per_pull = EXCLUDED.cost_per_pull,
				calculated_at = EXCLUDED.calculated_at
		`, manifestID, cost.SizeBytes, cost.StorageCostUSD, cost.BandwidthCostUSD,
			cost.TotalCostUSD, cost.PullCount30d, lastPulled, cost.CostPerPull, time.Now())
		
		if err != nil {
			fmt.Printf("[Costs] Failed to store cost for %s: %v\n", manifestID, err)
		}
		count++
	}
	
	fmt.Printf("[Costs] Refreshed costs for %d images\n", count)
	return nil
}

// GetDashboard returns the cost dashboard summary, filtered by user permission
func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID, role string) (*CostDashboard, error) {
	dashboard := &CostDashboard{}
	
	// Base WHERE clause for isolation
	// If admin, show all (1=1). If user, show only their namespace.
	whereClause := "1=1"
	args := []interface{}{}
	if role != "admin" {
		whereClause = "r.owner_id = $1"
		args = append(args, userID)
	}

	// 1. Get Total Costs
	queryTotal := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(sc.storage_cost_usd), 0),
			COALESCE(SUM(sc.bandwidth_cost_usd), 0),
			COALESCE(SUM(sc.total_cost_usd), 0),
			COUNT(*)
		FROM storage_costs sc
		JOIN manifests m ON sc.manifest_id = m.id
		JOIN repositories r ON m.repository_id = r.id
		JOIN namespaces n ON r.namespace_id = n.id
		WHERE %s
	`, whereClause)

	err := s.DB.QueryRowContext(ctx, queryTotal, args...).Scan(
		&dashboard.TotalStorageCostUSD, &dashboard.TotalBandwidthCostUSD,
		&dashboard.TotalCostUSD, &dashboard.TotalImages)
	
	if err != nil {
		return nil, err
	}
	
	// 2. Get Zombie Count (Filtered)
	// ZombieImage table captures manifest, so we need to join back to verify ownership for display
	queryZombie := fmt.Sprintf(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(zi.storage_cost_usd), 0)
		FROM zombie_images zi
		JOIN manifests m ON zi.manifest_id = m.id
		JOIN repositories r ON m.repository_id = r.id
		JOIN namespaces n ON r.namespace_id = n.id
		WHERE zi.recommended_action IN ('delete', 'archive')
		AND %s
	`, whereClause)

	err = s.DB.QueryRowContext(ctx, queryZombie, args...).Scan(&dashboard.ZombieImages, &dashboard.PotentialSavingsUSD)
	if err != nil {
		fmt.Printf("[Costs] Failed to get zombie stats: %v\n", err)
	}
	
	// 3. Get Top 10 Expensive Images (Filtered)
	queryTop := fmt.Sprintf(`
		SELECT 
			sc.manifest_id,
			r.name,
			COALESCE(t.name, 'untagged'),
			sc.blob_size_bytes,
			sc.storage_cost_usd,
			sc.bandwidth_cost_usd,
			sc.total_cost_usd,
			sc.pull_count_30d,
			sc.last_pulled_at,
			sc.cost_per_pull
		FROM storage_costs sc
		JOIN manifests m ON sc.manifest_id = m.id
		JOIN repositories r ON m.repository_id = r.id
		JOIN namespaces n ON r.namespace_id = n.id
		LEFT JOIN tags t ON t.manifest_id = m.id
		WHERE %s
		ORDER BY sc.total_cost_usd DESC
		LIMIT 10
	`, whereClause)
	
	rows, err := s.DB.QueryContext(ctx, queryTop, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cost ImageCost
			var lastPulled sql.NullTime
			if err := rows.Scan(&cost.ManifestID, &cost.Repository, &cost.Tag,
				&cost.SizeBytes, &cost.StorageCostUSD, &cost.BandwidthCostUSD,
				&cost.TotalCostUSD, &cost.PullCount30d, &lastPulled, &cost.CostPerPull); err == nil {
				
				if lastPulled.Valid {
					cost.LastPulledAt = &lastPulled.Time
				}
				dashboard.TopExpensiveImages = append(dashboard.TopExpensiveImages, cost)
			}
		}
	}
	
	dashboard.CostTrend = "stable"
	
	return dashboard, nil
}

// DetectZombieImages identifies images not pulled in X days (User Isolated)
func (s *Service) DetectZombieImages(ctx context.Context, daysThreshold int, userID uuid.UUID, role string) ([]ZombieImage, error) {
	if daysThreshold == 0 {
		daysThreshold = 90
	}
	
	// Determine isolation filter
	whereClause := "1=1"
	args := []interface{}{daysThreshold}
	if role != "admin" {
		whereClause = "r.owner_id = $2"
		args = append(args, userID)
	}

	// Note: We used to clear the table here, but with multi-user isolation, 
	// clearing everything breaks other users' data if this table is shared.
	// For MVP, we will just upsert/calculate on the fly and return results.
	// However, if we want to PERSIST only relevant zombies, we might need a better strategy.
	// Let's assume 'zombie_images' is a cache. 
	// For now, let's query raw potential zombies first, return them, and upsert them.
	
	query := fmt.Sprintf(`
		SELECT 
			m.id,
			r.name as repository,
			COALESCE(t.name, 'latest') as tag,
			EXTRACT(DAY FROM (NOW() - COALESCE(m.last_pulled_at, m.created_at))) as days_since_pull,
			COALESCE(sc.storage_cost_usd, 0) as storage_cost
		FROM manifests m
		JOIN repositories r ON m.repository_id = r.id
		JOIN namespaces n ON r.namespace_id = n.id
		LEFT JOIN tags t ON t.manifest_id = m.id
		LEFT JOIN storage_costs sc ON sc.manifest_id = m.id
		WHERE ((m.last_pulled_at IS NULL AND m.created_at < NOW() - INTERVAL '1 day' * $1)
		   OR m.last_pulled_at < NOW() - INTERVAL '1 day' * $1)
		AND %s
		ORDER BY days_since_pull DESC
	`, whereClause)
	
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var zombies []ZombieImage
	for rows.Next() {
		var z ZombieImage
		if err := rows.Scan(&z.ManifestID, &z.Repository, &z.Tag, &z.DaysSinceLastPull, &z.StorageCostUSD); err != nil {
			continue
		}
		
		if z.DaysSinceLastPull > 180 {
			z.RecommendedAction = "delete"
		} else if z.DaysSinceLastPull > 120 {
			z.RecommendedAction = "archive"
		} else {
			z.RecommendedAction = "monitor"
		}
		
		zombies = append(zombies, z)
		
		// Upsert into zombie_images table (Global Table)
		// It's safe to upsert our view.
		_, err := s.DB.ExecContext(ctx, `
			INSERT INTO zombie_images (manifest_id, days_since_last_pull, storage_cost_usd, recommended_action)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (manifest_id) DO UPDATE SET
				days_since_last_pull = EXCLUDED.days_since_last_pull,
				storage_cost_usd = EXCLUDED.storage_cost_usd,
				recommended_action = EXCLUDED.recommended_action,
				detected_at = CURRENT_TIMESTAMP
		`, z.ManifestID, z.DaysSinceLastPull, z.StorageCostUSD, z.RecommendedAction)
		
		if err != nil {
			fmt.Printf("[Costs] Failed to store zombie image %s: %v\n", z.ManifestID, err)
		}
	}
	
	return zombies, nil
}

// CleanupZombies deletes zombie images based on criteria (User Isolated)
func (s *Service) CleanupZombies(ctx context.Context, daysThreshold int, dryRun bool, userID uuid.UUID, role string) (int, error) {
	if daysThreshold == 0 {
		daysThreshold = 180 
	}
	
	fmt.Printf("[Costs] CleanupZombies calling with daysThreshold=%d, dryRun=%v, role=%s\n", daysThreshold, dryRun, role)

	// Refresh list first
	if _, err := s.DetectZombieImages(ctx, daysThreshold, userID, role); err != nil {
		return 0, err
	}

	// Filter deletion candidates by user ownership
	whereClause := "1=1"
	args := []interface{}{daysThreshold}
	if role != "admin" {
		whereClause = "r.owner_id = $2"
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
		SELECT zi.manifest_id, zi.days_since_last_pull, zi.recommended_action
		FROM zombie_images zi
		JOIN manifests m ON zi.manifest_id = m.id
		JOIN repositories r ON m.repository_id = r.id
		JOIN namespaces n ON r.namespace_id = n.id
		WHERE zi.days_since_last_pull > $1
		AND zi.recommended_action = 'delete'
		AND %s
	`, whereClause)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var manifestID uuid.UUID
		var days int
		var action string
		if err := rows.Scan(&manifestID, &days, &action); err != nil {
			continue
		}
		
		fmt.Printf("[Costs] Found zombie to delete: %s (days=%d)\n", manifestID, days)
		
		if !dryRun {
			// Delete manifest
			_, err := s.DB.ExecContext(ctx, `DELETE FROM manifests WHERE id = $1`, manifestID)
			if err != nil {
				fmt.Printf("[Costs] Failed to delete manifest %s: %v\n", manifestID, err)
				continue
			}
			
			// Also remove from zombie_images
			s.DB.ExecContext(ctx, `DELETE FROM zombie_images WHERE manifest_id = $1`, manifestID)
		}
		count++
	}
	
	if dryRun {
		fmt.Printf("[Costs] DRY RUN: Would delete %d zombie images\n", count)
	} else {
		fmt.Printf("[Costs] Deleted %d zombie images\n", count)
	}
	
	return count, nil
}
