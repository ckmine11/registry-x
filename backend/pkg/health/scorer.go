package health

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// HealthScore represents the overall health of an image
type HealthScore struct {
	Overall     int       `json:"overall"`     // 0-100
	Security    int       `json:"security"`    // 0-100
	Freshness   int       `json:"freshness"`   // 0-100
	Efficiency  int       `json:"efficiency"`  // 0-100
	Maintenance int       `json:"maintenance"` // 0-100
	Grade       string    `json:"grade"`       // A+, A, B, C, D, F
	Trend       string    `json:"trend"`       // improving, stable, declining
	LastUpdated time.Time `json:"lastUpdated"`
}

// ImageMetrics contains the raw data needed for health calculation
type ImageMetrics struct {
	ManifestID        uuid.UUID
	TotalVulns        int
	CriticalVulns     int
	HighVulns         int
	MediumVulns       int
	LowVulns          int
	ImageSizeBytes    int64
	CreatedAt         time.Time
	LastPushedAt      time.Time
	PullCount         int
	AverageSizeInRepo int64 // Average size of similar images in the same repo
}

// Scorer is the health scoring engine
type Scorer struct{}

// NewScorer creates a new health scorer
func NewScorer() *Scorer {
	return &Scorer{}
}

// CalculateHealthScore computes the composite health score for an image
func (s *Scorer) CalculateHealthScore(metrics *ImageMetrics) *HealthScore {
	security := s.calculateSecurityScore(metrics)
	freshness := s.calculateFreshnessScore(metrics)
	efficiency := s.calculateEfficiencyScore(metrics)
	maintenance := s.calculateMaintenanceScore(metrics)

	// Weighted average: Security (40%), Freshness (20%), Efficiency (20%), Maintenance (20%)
	overall := int(float64(security)*0.4 + float64(freshness)*0.2 + float64(efficiency)*0.2 + float64(maintenance)*0.2)

	return &HealthScore{
		Overall:     overall,
		Security:    security,
		Freshness:   freshness,
		Efficiency:  efficiency,
		Maintenance: maintenance,
		Grade:       s.calculateGrade(overall),
		Trend:       "stable", // TODO: Implement trend tracking
		LastUpdated: time.Now(),
	}
}

// calculateSecurityScore scores based on vulnerability count and severity using exponential decay
func (s *Scorer) calculateSecurityScore(metrics *ImageMetrics) int {
	// Calculate weighted penalty
	// Critical: 10 points
	// High: 5 points
	// Medium: 1 point
	// Low: 0.1 points
	penalty := float64(metrics.CriticalVulns)*10.0 + 
		float64(metrics.HighVulns)*5.0 + 
		float64(metrics.MediumVulns)*1.0 + 
		float64(metrics.LowVulns)*0.1

	// Use exponential decay function: Score = 100 * e^(-penalty / 60)
	// This ensures:
	// - 0 vulns -> 100
	// - 1 Critical (10 penalty) -> ~84
	// - 1 Critical + 1 High (15 penalty) -> ~77
	// - 5 Criticals (50 penalty) -> ~43
	// - Massive vulns -> approaches 0 but stays non-negative
	
	score := 100.0 * math.Exp(-penalty/60.0)

	// Round to nearest integer
	return int(math.Round(score))
}

// calculateFreshnessScore scores based on image age
func (s *Scorer) calculateFreshnessScore(metrics *ImageMetrics) int {
	now := time.Now()
	daysSinceCreation := now.Sub(metrics.CreatedAt).Hours() / 24
	daysSinceLastPush := now.Sub(metrics.LastPushedAt).Hours() / 24

	// Use the more recent of the two dates
	daysOld := daysSinceCreation
	if daysSinceLastPush < daysOld {
		daysOld = daysSinceLastPush
	}

	// Scoring logic:
	// 0-7 days: 100 points
	// 8-30 days: 90 points
	// 31-90 days: 70 points
	// 91-180 days: 50 points
	// 181-365 days: 30 points
	// 365+ days: 10 points

	switch {
	case daysOld <= 7:
		return 100
	case daysOld <= 30:
		return 90
	case daysOld <= 90:
		return 70
	case daysOld <= 180:
		return 50
	case daysOld <= 365:
		return 30
	default:
		return 10
	}
}

// calculateEfficiencyScore scores based on image size compared to repo average
func (s *Scorer) calculateEfficiencyScore(metrics *ImageMetrics) int {
	// If no average size available, give neutral score
	if metrics.AverageSizeInRepo == 0 {
		return 75
	}

	// Calculate size ratio
	ratio := float64(metrics.ImageSizeBytes) / float64(metrics.AverageSizeInRepo)

	// Scoring logic:
	// < 50% of average: 100 points (very efficient)
	// 50-75% of average: 90 points
	// 75-100% of average: 80 points
	// 100-125% of average: 70 points
	// 125-150% of average: 50 points
	// 150-200% of average: 30 points
	// > 200% of average: 10 points (bloated)

	switch {
	case ratio < 0.5:
		return 100
	case ratio < 0.75:
		return 90
	case ratio < 1.0:
		return 80
	case ratio < 1.25:
		return 70
	case ratio < 1.5:
		return 50
	case ratio < 2.0:
		return 30
	default:
		return 10
	}
}

// calculateMaintenanceScore scores based on pull activity and recency
func (s *Scorer) calculateMaintenanceScore(metrics *ImageMetrics) int {
	// Pull count scoring (50% of maintenance score)
	pullScore := 0
	switch {
	case metrics.PullCount >= 100:
		pullScore = 50
	case metrics.PullCount >= 50:
		pullScore = 40
	case metrics.PullCount >= 20:
		pullScore = 30
	case metrics.PullCount >= 10:
		pullScore = 20
	case metrics.PullCount >= 5:
		pullScore = 10
	default:
		pullScore = 5
	}

	// Last push recency (50% of maintenance score)
	daysSinceLastPush := time.Now().Sub(metrics.LastPushedAt).Hours() / 24
	recencyScore := 0
	switch {
	case daysSinceLastPush <= 7:
		recencyScore = 50
	case daysSinceLastPush <= 30:
		recencyScore = 40
	case daysSinceLastPush <= 90:
		recencyScore = 30
	case daysSinceLastPush <= 180:
		recencyScore = 20
	default:
		recencyScore = 10
	}

	return pullScore + recencyScore
}

// calculateGrade converts numeric score to letter grade
func (s *Scorer) calculateGrade(score int) string {
	switch {
	case score >= 95:
		return "A+"
	case score >= 90:
		return "A"
	case score >= 85:
		return "A-"
	case score >= 80:
		return "B+"
	case score >= 75:
		return "B"
	case score >= 70:
		return "B-"
	case score >= 65:
		return "C+"
	case score >= 60:
		return "C"
	case score >= 55:
		return "C-"
	case score >= 50:
		return "D"
	default:
		return "F"
	}
}

// GetScoreColor returns the color code for UI display
func (s *Scorer) GetScoreColor(score int) string {
	switch {
	case score >= 80:
		return "green"
	case score >= 60:
		return "yellow"
	case score >= 40:
		return "orange"
	default:
		return "red"
	}
}

// GetScoreDescription returns a human-readable description
func (s *Scorer) GetScoreDescription(score int) string {
	switch {
	case score >= 90:
		return "Excellent - Production ready"
	case score >= 75:
		return "Good - Minor improvements recommended"
	case score >= 60:
		return "Fair - Some concerns to address"
	case score >= 40:
		return "Poor - Significant issues detected"
	default:
		return "Critical - Immediate attention required"
	}
}

// FormatScore returns a formatted string representation
func (score *HealthScore) String() string {
	return fmt.Sprintf("Health Score: %d/100 (Grade: %s) - Security: %d, Freshness: %d, Efficiency: %d, Maintenance: %d",
		score.Overall, score.Grade, score.Security, score.Freshness, score.Efficiency, score.Maintenance)
}
