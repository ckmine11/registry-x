package scanner

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/registryx/registryx/backend/pkg/config"
)

type Service struct {
	DB     *sql.DB
	Config *config.Config
}

func NewService(db *sql.DB, cfg *config.Config) *Service {
	return &Service{
		DB:     db,
		Config: cfg,
	}
}

// ScanManifest triggers a Trivy scan for the given manifest.
// For MVP, this runs 'trivy' as a subprocess.
// In prod, this would likely enqueue a job to a worker pool.
// ScanManifest triggers a Trivy scan for the given manifest.
// For MVP, this runs 'trivy' as a subprocess.
// In prod, this would likely enqueue a job to a worker pool.
func (s *Service) ScanManifest(ctx context.Context, manifestID uuid.UUID, repoName, reference string) {
	fmt.Printf("Scanning manifest %s (repo: %s, ref: %s)...\n", manifestID, repoName, reference)

	// Update status to 'scanning'
	s.updateStatus(ctx, manifestID, "scanning")

	// Run Trivy
	// Point trivy to the registry URL.
	// URI Format: localhost:5000/library/nginx:latest OR localhost:5000/library/nginx@sha256:...
	
	var imageURI string
	port := strings.TrimPrefix(s.Config.ServerPort, ":")
	if strings.HasPrefix(reference, "sha256:") {
		imageURI = fmt.Sprintf("localhost:%s/%s@%s", port, repoName, reference)
	} else {
		imageURI = fmt.Sprintf("localhost:%s/%s:%s", port, repoName, reference)
	}
	
	// Command: trivy image --format json --output - <imageURI>
	// Note: We might need --insecure if using http/self-signed.
	cmd := exec.CommandContext(ctx, "trivy", "image", "--format", "json", "-q", "--insecure", imageURI)
	
	// Environment for auth if needed
	// cmd.Env = append(os.Environ(), "TRIVY_USERNAME=admin", "TRIVY_PASSWORD=...")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Scanner] Scan failed for manifest %s (repo: %s, ref: %s): %v. Output: %s\n", 
			manifestID, repoName, reference, err, string(output))
		s.updateStatus(ctx, manifestID, "failed")
		return
	}

	// Parse Logic
	_, summary, err := parseTrivyOutput(output)
	if err != nil {
		fmt.Printf("Parse failed: %v\n", err)
		s.updateStatus(ctx, manifestID, "failed")
		return
	}

	// Store Report
	err = s.saveReport(ctx, manifestID, output, summary)
	if err != nil {
		fmt.Printf("Save report failed: %v\n", err)
	} else {
		fmt.Printf("Scan completed for %s\n", reference)
	}
}

func (s *Service) updateStatus(ctx context.Context, manifestID uuid.UUID, status string) {
	// Upsert initial record if not exists?
	// The table `vulnerability_reports` should ideally be 1:1 or 1:Many with manifest.
	// Schema: id, manifest_id, status...
	
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO vulnerability_reports (manifest_id, scanner, status)
		VALUES ($1, 'trivy', $2)
		ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status`, 
		// Wait, id constraint? We need to find by manifest_id.
		// Let's assume we insert on creation or allow duplicate reports (history).
		// For MVP, let's just insert a new report or update the latest one.
		manifestID, status)
	
	if err != nil {
		// Just log
		fmt.Println("Error updating scan status:", err)
	}
}

func (s *Service) saveReport(ctx context.Context, manifestID uuid.UUID, rawJSON []byte, summary ScanSummary) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE vulnerability_reports 
		SET status = 'completed', 
		    report_json = $2,
			critical_count = $3,
			high_count = $4,
			medium_count = $5,
			low_count = $6,
			scanned_at = CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM vulnerability_reports 
			WHERE manifest_id = $1 AND status = 'scanning'
			ORDER BY scanned_at DESC LIMIT 1
		)`,
		manifestID, rawJSON, summary.Critical, summary.High, summary.Medium, summary.Low)
	return err
}

type ScanSummary struct {
	Status       string `json:"status"`
	Critical     int    `json:"critical"`
	High         int    `json:"high"`
	Medium       int    `json:"medium"`
	Low          int    `json:"low"`
	HighPriority int    `json:"high_priority"` // EPSS / Reachable
}

// Minimal Trivy JSON structs for parsing
type TrivyReport struct {
	Results []struct {
		Vulnerabilities []struct {
			Severity string `json:"Severity"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

func parseTrivyOutput(data []byte) (*TrivyReport, ScanSummary, error) {
	var report TrivyReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, ScanSummary{}, err
	}

	summary := ScanSummary{Status: "completed"}
	for _, res := range report.Results {
		for _, vuln := range res.Vulnerabilities {
			switch strings.ToUpper(vuln.Severity) {
			case "CRITICAL":
				summary.Critical++
			case "HIGH":
				summary.High++
			case "MEDIUM":
				summary.Medium++
			case "LOW":
				summary.Low++
			}
		}
	}
	// --- Smart Prioritization ---
	// Real implementation would check EPSS scores or exploitability
	// For now, we report 0 high priority until connected to a threat intel source
	summary.HighPriority = 0
	// ----------------------------------------

	return &report, summary, nil
}

// GetVulnerabilitySummary fetches the latest scan summary for a manifest.
func (s *Service) GetVulnerabilitySummary(ctx context.Context, manifestID uuid.UUID) (*ScanSummary, error) {
	var summary ScanSummary
	
	err := s.DB.QueryRowContext(ctx, `
		SELECT status, critical_count, high_count, medium_count, low_count
		FROM vulnerability_reports
		WHERE manifest_id = $1 AND (status = 'completed' OR status = 'scanning')
		ORDER BY scanned_at DESC LIMIT 1`, manifestID).Scan(&summary.Status, &summary.Critical, &summary.High, &summary.Medium, &summary.Low)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// No report yet. Return 0 counts instead of mock to avoid confusion if it's truly empty
			return &ScanSummary{Status: "pending"}, nil
		}
		return nil, err
	}
	
	return &summary, nil
}

// ScanStatus represents the current status of a vulnerability scan
type ScanStatus struct {
	Status      string       `json:"status"` // "pending", "scanning", "completed", "failed"
	ScannedAt   *string      `json:"scanned_at,omitempty"`
	Summary     *ScanSummary `json:"summary,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// GetScanStatus returns the current scan status for a manifest
func (s *Service) GetScanStatus(ctx context.Context, manifestID uuid.UUID) (*ScanStatus, error) {
	var status ScanStatus
	var scannedAt sql.NullTime
	var critical, high, medium, low sql.NullInt64
	
	err := s.DB.QueryRowContext(ctx, `
		SELECT status, scanned_at, critical_count, high_count, medium_count, low_count
		FROM vulnerability_reports
		WHERE manifest_id = $1
		ORDER BY scanned_at DESC LIMIT 1`, manifestID).Scan(
		&status.Status, &scannedAt, &critical, &high, &medium, &low)
	
	if err != nil {
		if err == sql.ErrNoRows {
			status.Status = "pending"
			return &status, nil
		}
		return nil, err
	}
	
	if scannedAt.Valid {
		timeStr := scannedAt.Time.Format("2006-01-02T15:04:05Z")
		status.ScannedAt = &timeStr
	}
	
	if status.Status == "scanning" && scannedAt.Valid {
		// If it's been scanning for more than 5 minutes, consider it failed/stuck
		if time.Since(scannedAt.Time) > 5*time.Minute {
			status.Status = "failed"
			status.Error = "Scan timed out (started > 5m ago)"
		}
	}

	if (status.Status == "completed" || status.Status == "scanning") && critical.Valid {
		status.Summary = &ScanSummary{
			Critical: int(critical.Int64),
			High:     int(high.Int64),
			Medium:   int(medium.Int64),
			Low:      int(low.Int64),
		}
	}
	
	return &status, nil
}

// GetScanReport returns the full Trivy JSON report for a manifest
func (s *Service) GetScanReport(ctx context.Context, manifestID uuid.UUID) ([]byte, error) {
	var reportJSON []byte
	var status string
	
	err := s.DB.QueryRowContext(ctx, `
		SELECT status, report_json
		FROM vulnerability_reports
		WHERE manifest_id = $1 AND status = 'completed'
		ORDER BY scanned_at DESC LIMIT 1`, manifestID).Scan(&status, &reportJSON)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no completed scan report found")
		}
		return nil, err
	}
	
	return reportJSON, nil
}

// ScanHistoryEntry represents a single scan in the history
type ScanHistoryEntry struct {
	ID        uuid.UUID    `json:"id"`
	Status    string       `json:"status"`
	ScannedAt *string      `json:"scanned_at,omitempty"`
	Summary   *ScanSummary `json:"summary,omitempty"`
}

// GetScanHistory returns all scan attempts for a manifest
func (s *Service) GetScanHistory(ctx context.Context, manifestID uuid.UUID) ([]ScanHistoryEntry, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, status, scanned_at, critical_count, high_count, medium_count, low_count
		FROM vulnerability_reports
		WHERE manifest_id = $1
		ORDER BY scanned_at DESC`, manifestID)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var history []ScanHistoryEntry
	for rows.Next() {
		var entry ScanHistoryEntry
		var scannedAt sql.NullTime
		var critical, high, medium, low sql.NullInt64
		
		err := rows.Scan(&entry.ID, &entry.Status, &scannedAt, &critical, &high, &medium, &low)
		if err != nil {
			return nil, err
		}
		
		if scannedAt.Valid {
			timeStr := scannedAt.Time.Format("2006-01-02T15:04:05Z")
			entry.ScannedAt = &timeStr
		}
		
		if entry.Status == "completed" && critical.Valid {
			entry.Summary = &ScanSummary{
				Critical: int(critical.Int64),
				High:     int(high.Int64),
				Medium:   int(medium.Int64),
				Low:      int(low.Int64),
			}
		}
		
		history = append(history, entry)
	}
	
	return history, nil
}
