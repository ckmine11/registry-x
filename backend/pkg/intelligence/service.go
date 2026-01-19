package intelligence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/registryx/registryx/backend/pkg/epss"
)

// Service handles vulnerability intelligence operations
type Service struct {
	DB         *sql.DB
	EPSSClient *epss.Client
}

// VulnIntelligence represents enriched vulnerability data
type VulnIntelligence struct {
	ID               uuid.UUID
	CVEID            string
	EPSSScore        float64
	EPSSPercentile   float64
	HasActiveExploit bool
	ExploitMaturity  string
	TrendingScore    int
	LastUpdated      time.Time
}

// VulnPriority represents a prioritized vulnerability for a manifest
type VulnPriority struct {
	ID                 uuid.UUID
	ManifestID         uuid.UUID
	CVEID              string
	BaseSeverity       string
	EPSSScore          float64
	RuntimeExposed     bool
	PriorityScore      int
	RecommendedAction  string
	Created            time.Time
}

// NewService creates a new vulnerability intelligence service
func NewService(db *sql.DB) *Service {
	return &Service{
		DB:         db,
		EPSSClient: epss.NewClient(),
	}
}

// RefreshEPSSData fetches and stores EPSS scores for all known CVEs
func (s *Service) RefreshEPSSData(ctx context.Context) error {
	// Get all unique CVE IDs from vulnerability_reports
	rows, err := s.DB.QueryContext(ctx, `
		SELECT DISTINCT v->>'VulnerabilityID' as cve_id
		FROM vulnerability_reports,
		     jsonb_array_elements(report_json->'Results') as rs,
		     jsonb_array_elements(rs->'Vulnerabilities') as v
		WHERE report_json IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to query CVEs: %w", err)
	}
	defer rows.Close()

	var cveIDs []string
	for rows.Next() {
		var cveID string
		if err := rows.Scan(&cveID); err != nil {
			continue
		}
		cveIDs = append(cveIDs, cveID)
	}

	if len(cveIDs) == 0 {
		fmt.Println("[Intelligence] No CVEs found to refresh")
		return nil
	}

	fmt.Printf("[Intelligence] Refreshing EPSS data for %d CVEs\n", len(cveIDs))

	// Fetch EPSS scores in bulk
	scores, err := s.EPSSClient.GetBulkScores(ctx, cveIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch EPSS scores: %w", err)
	}

	// Store in database
	for cveID, score := range scores {
		err := s.StoreVulnIntelligence(ctx, cveID, score.EPSS, score.Percentile)
		if err != nil {
			fmt.Printf("[Intelligence] Failed to store %s: %v\n", cveID, err)
		}
	}

	fmt.Printf("[Intelligence] Successfully refreshed %d EPSS scores\n", len(scores))
	return nil
}

// StoreVulnIntelligence stores or updates vulnerability intelligence data
func (s *Service) StoreVulnIntelligence(ctx context.Context, cveID string, epssScore, epssPercentile float64) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO vulnerability_intelligence (cve_id, epss_score, epss_percentile, last_updated)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (cve_id) DO UPDATE SET
			epss_score = EXCLUDED.epss_score,
			epss_percentile = EXCLUDED.epss_percentile,
			last_updated = EXCLUDED.last_updated
	`, cveID, epssScore, epssPercentile, time.Now())

	return err
}

// GetVulnIntelligence retrieves intelligence data for a CVE
func (s *Service) GetVulnIntelligence(ctx context.Context, cveID string) (*VulnIntelligence, error) {
	var intel VulnIntelligence
	var lastUpdated sql.NullTime

	err := s.DB.QueryRowContext(ctx, `
		SELECT id, cve_id, epss_score, epss_percentile, has_active_exploit,
		       exploit_maturity, trending_score, last_updated
		FROM vulnerability_intelligence
		WHERE cve_id = $1
	`, cveID).Scan(
		&intel.ID, &intel.CVEID, &intel.EPSSScore, &intel.EPSSPercentile,
		&intel.HasActiveExploit, &intel.ExploitMaturity, &intel.TrendingScore,
		&lastUpdated,
	)

	if err != nil {
		return nil, err
	}

	if lastUpdated.Valid {
		intel.LastUpdated = lastUpdated.Time
	}

	return &intel, nil
}

// CalculatePriorityScore calculates a priority score for a vulnerability
func (s *Service) CalculatePriorityScore(baseSeverity string, epssScore float64, runtimeExposed bool) int {
	score := 0

	// Base severity (30%)
	switch baseSeverity {
	case "CRITICAL":
		score += 30
	case "HIGH":
		score += 22
	case "MEDIUM":
		score += 15
	case "LOW":
		score += 7
	}

	// EPSS score (40%)
	score += int(epssScore * 40)

	// Active exploit (20%) - simplified, would check exploit databases
	// For now, high EPSS score (>0.5) suggests likely exploitation
	if epssScore > 0.5 {
		score += 20
	} else if epssScore > 0.2 {
		score += 10
	}

	// Runtime exposure (10%)
	if runtimeExposed {
		score += 10
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// GetRecommendedAction returns the recommended action based on priority score
func (s *Service) GetRecommendedAction(priorityScore int) string {
	switch {
	case priorityScore >= 80:
		return "urgent"
	case priorityScore >= 60:
		return "high"
	case priorityScore >= 40:
		return "medium"
	case priorityScore >= 20:
		return "low"
	default:
		return "monitor"
	}
}

// CalculateManifestPriorities calculates and stores priority scores for all vulnerabilities in a manifest
func (s *Service) CalculateManifestPriorities(ctx context.Context, manifestID uuid.UUID) error {
	// 1. Get the latest completed report
	var reportJSON []byte
	err := s.DB.QueryRowContext(ctx, `
		SELECT report_json FROM vulnerability_reports 
		WHERE manifest_id = $1 AND status = 'completed'
		ORDER BY scanned_at DESC LIMIT 1
	`, manifestID).Scan(&reportJSON)

	if err != nil {
		if err == sql.ErrNoRows { return nil }
		return err
	}

	// 2. Parse Minimal Trivy JSON
	var report struct {
		Results []struct {
			Vulnerabilities []struct {
				VulnerabilityID string `json:"VulnerabilityID"`
				Severity        string `json:"Severity"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}
	if err := json.Unmarshal(reportJSON, &report); err != nil {
		return fmt.Errorf("failed to unmarshal report: %w", err)
	}

	// 3. Clear existing priorities for this manifest
	_, _ = s.DB.ExecContext(ctx, "DELETE FROM manifest_vuln_priority WHERE manifest_id = $1", manifestID)

	highPriorityCount := 0

	// 4. Process each vuln
	for _, res := range report.Results {
		for _, v := range res.Vulnerabilities {
			// Get EPSS score if available
			var epssScore float64
			_ = s.DB.QueryRowContext(ctx, "SELECT COALESCE(epss_score, 0) FROM vulnerability_intelligence WHERE cve_id = $1", v.VulnerabilityID).Scan(&epssScore)

			runtimeExposed := false // Future: Hook into K8s runtime data
			priorityScore := s.CalculatePriorityScore(v.Severity, epssScore, runtimeExposed)
			recommendedAction := s.GetRecommendedAction(priorityScore)

			if priorityScore >= 70 {
				highPriorityCount++
			}

			// Store Priority
			_, err = s.DB.ExecContext(ctx, `
				INSERT INTO manifest_vuln_priority (manifest_id, cve_id, base_severity, epss_score, runtime_exposed, priority_score, recommended_action)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				manifestID, v.VulnerabilityID, v.Severity, epssScore, runtimeExposed, priorityScore, recommendedAction)
			
			if err != nil {
				fmt.Printf("[Intelligence] Failed to store priority for %s: %v\n", v.VulnerabilityID, err)
			}
		}
	}

	fmt.Printf("[Intelligence] Calculated priorities for manifest %s (High Priority: %d)\n", manifestID, highPriorityCount)
	return nil
}

// GetPrioritizedVulnerabilities returns vulnerabilities sorted by priority
func (s *Service) GetPrioritizedVulnerabilities(ctx context.Context, manifestID uuid.UUID) ([]VulnPriority, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, manifest_id, cve_id, base_severity, epss_score,
		       runtime_exposed, priority_score, recommended_action, created_at
		FROM manifest_vuln_priority
		WHERE manifest_id = $1
		ORDER BY priority_score DESC
	`, manifestID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var priorities []VulnPriority
	for rows.Next() {
		var p VulnPriority
		err := rows.Scan(
			&p.ID, &p.ManifestID, &p.CVEID, &p.BaseSeverity, &p.EPSSScore,
			&p.RuntimeExposed, &p.PriorityScore, &p.RecommendedAction, &p.Created,
		)
		if err != nil {
			continue
		}
		priorities = append(priorities, p)
	}

	return priorities, nil
}
