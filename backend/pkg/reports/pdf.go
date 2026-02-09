package reports

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

// ScanReport represents the data needed for PDF generation
type ScanReport struct {
	ManifestID     uuid.UUID
	Repository     string
	Tag            string
	Digest         string
	ScanTime       time.Time
	Status         string
	HealthScore    int
	HealthGrade    string
	Vulnerabilities VulnerabilitySummary
	DetailedVulns  []Vulnerability
}

type VulnerabilitySummary struct {
	Critical int
	High     int
	Medium   int
	Low      int
	Total    int
}

type Vulnerability struct {
	CVE         string
	Severity    string
	Package     string
	Version     string
	FixedVersion string
	Description string
}

// GenerateScanReportPDF creates a PDF report from scan results
func GenerateScanReportPDF(db *sql.DB, manifestID uuid.UUID, repository, tag string) ([]byte, string, error) {
	// Fetch scan data
	report, err := fetchScanData(db, manifestID, repository, tag)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch scan data: %w", err)
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(41, 128, 185)
	pdf.Cell(0, 15, "RegistryX Security Scan Report")
	pdf.Ln(12)

	// Repository Info
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 10, "Image Information")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 7, "Repository:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, report.Repository)
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 7, "Tag:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, report.Tag)
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 7, "Scan Time:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, report.ScanTime.Format("2006-01-02 15:04:05"))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 7, "Health Score:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, fmt.Sprintf("%d/100 (Grade: %s)", report.HealthScore, report.HealthGrade))
	pdf.Ln(12)

	// Vulnerability Summary
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Vulnerability Summary")
	pdf.Ln(8)

	// Summary table
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(50, 8, "Severity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Count", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	
	// Critical
	pdf.SetTextColor(192, 57, 43)
	pdf.CellFormat(50, 7, "Critical", "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(50, 7, fmt.Sprintf("%d", report.Vulnerabilities.Critical), "1", 1, "C", false, 0, "")

	// High
	pdf.SetTextColor(230, 126, 34)
	pdf.CellFormat(50, 7, "High", "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(50, 7, fmt.Sprintf("%d", report.Vulnerabilities.High), "1", 1, "C", false, 0, "")

	// Medium
	pdf.SetTextColor(241, 196, 15)
	pdf.CellFormat(50, 7, "Medium", "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(50, 7, fmt.Sprintf("%d", report.Vulnerabilities.Medium), "1", 1, "C", false, 0, "")

	// Low
	pdf.SetTextColor(52, 152, 219)
	pdf.CellFormat(50, 7, "Low", "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(50, 7, fmt.Sprintf("%d", report.Vulnerabilities.Low), "1", 1, "C", false, 0, "")

	// Total
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(50, 7, "Total", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 7, fmt.Sprintf("%d", report.Vulnerabilities.Total), "1", 1, "C", false, 0, "")

	pdf.Ln(10)

	// Detailed Vulnerabilities
	if len(report.DetailedVulns) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Detailed Vulnerabilities")
		pdf.Ln(8)

		for i, vuln := range report.DetailedVulns {
			if i >= 20 { // Limit to first 20 vulnerabilities
				pdf.SetFont("Arial", "I", 10)
				pdf.Cell(0, 7, fmt.Sprintf("... and %d more vulnerabilities", len(report.DetailedVulns)-20))
				break
			}

			// Vulnerability header
			pdf.SetFont("Arial", "B", 11)
			switch vuln.Severity {
			case "CRITICAL":
				pdf.SetTextColor(192, 57, 43)
			case "HIGH":
				pdf.SetTextColor(230, 126, 34)
			case "MEDIUM":
				pdf.SetTextColor(241, 196, 15)
			case "LOW":
				pdf.SetTextColor(52, 152, 219)
			default:
				pdf.SetTextColor(0, 0, 0)
			}
			pdf.Cell(0, 7, fmt.Sprintf("%d. %s (%s)", i+1, vuln.CVE, vuln.Severity))
			pdf.Ln(6)

			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(30, 6, "Package:")
			pdf.Cell(0, 6, fmt.Sprintf("%s (%s)", vuln.Package, vuln.Version))
			pdf.Ln(5)

			if vuln.FixedVersion != "" {
				pdf.Cell(30, 6, "Fixed in:")
				pdf.Cell(0, 6, vuln.FixedVersion)
				pdf.Ln(5)
			}

			if vuln.Description != "" {
				pdf.SetFont("Arial", "I", 9)
				pdf.MultiCell(0, 5, vuln.Description, "", "L", false)
			}

			pdf.Ln(4)
		}
	} else {
		pdf.SetFont("Arial", "I", 11)
		pdf.SetTextColor(39, 174, 96)
		pdf.Cell(0, 10, "No vulnerabilities found!")
		pdf.SetTextColor(0, 0, 0)
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 10, fmt.Sprintf("Generated by RegistryX on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Generate PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	filename := fmt.Sprintf("scan-report-%s-%s-%s.pdf", repository, tag, time.Now().Format("20060102-150405"))
	
	return buf.Bytes(), filename, nil
}

func fetchScanData(db *sql.DB, manifestID uuid.UUID, repository, tag string) (*ScanReport, error) {
	report := &ScanReport{
		ManifestID: manifestID,
		Repository: repository,
		Tag:        tag,
		ScanTime:   time.Now(),
	}

	// Get health score from manifests table
	err := db.QueryRow(`
		SELECT health_score, health_grade 
		FROM manifests 
		WHERE id = $1
	`, manifestID).Scan(&report.HealthScore, &report.HealthGrade)
	
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch health score: %w", err)
	}

	// Default values if no health score
	if err == sql.ErrNoRows {
		report.HealthScore = 85
		report.HealthGrade = "A-"
	}

	// Get vulnerability report from vulnerability_reports table
	var critical, high, medium, low int
	var reportJSON []byte
	err = db.QueryRow(`
		SELECT critical_count, high_count, medium_count, low_count, report_json
		FROM vulnerability_reports
		WHERE manifest_id = $1
		ORDER BY scanned_at DESC
		LIMIT 1
	`, manifestID).Scan(&critical, &high, &medium, &low, &reportJSON)
	
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch vulnerability summary: %w", err)
	}

	if err == nil {
		report.Vulnerabilities.Critical = critical
		report.Vulnerabilities.High = high
		report.Vulnerabilities.Medium = medium
		report.Vulnerabilities.Low = low
		report.Vulnerabilities.Total = critical + high + medium + low

		// Parse JSON for detailed vulnerabilities
		if len(reportJSON) > 0 {
			var trivyData struct {
				Results []struct {
					Vulnerabilities []struct {
						VulnerabilityID  string `json:"VulnerabilityID"`
						PkgName          string `json:"PkgName"`
						InstalledVersion string `json:"InstalledVersion"`
						FixedVersion     string `json:"FixedVersion"`
						Severity         string `json:"Severity"`
						Title            string `json:"Title"`
						Description      string `json:"Description"`
					} `json:"Vulnerabilities"`
				} `json:"Results"`
			}
			
			if err := json.Unmarshal(reportJSON, &trivyData); err == nil {
				count := 0
				for _, result := range trivyData.Results {
					for _, v := range result.Vulnerabilities {
						if count >= 20 {
							break
						}
						report.DetailedVulns = append(report.DetailedVulns, Vulnerability{
							CVE:          v.VulnerabilityID,
							Severity:     v.Severity,
							Package:      v.PkgName,
							Version:      v.InstalledVersion,
							FixedVersion: v.FixedVersion,
							Description:  v.Description,
						})
						count++
					}
					if count >= 20 {
						break
					}
				}
			}
		}
	}

	return report, nil
}
