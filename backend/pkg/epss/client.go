package epss

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the EPSS API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// EPSSScore represents the EPSS data for a CVE
type EPSSScore struct {
	CVE        string  `json:"cve"`
	EPSS       float64 `json:"epss"`
	Percentile float64 `json:"percentile"`
	Date       string  `json:"date"`
}

// EPSSResponse represents the API response
type EPSSResponse struct {
	Status string `json:"status"`
	Data   []struct {
		CVE        string `json:"cve"`
		EPSS       string `json:"epss"`
		Percentile string `json:"percentile"`
		Date       string `json:"date"`
	} `json:"data"`
}

// NewClient creates a new EPSS API client
func NewClient() *Client {
	return &Client{
		BaseURL: "https://api.first.org/data/v1",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetScore fetches the EPSS score for a specific CVE
func (c *Client) GetScore(ctx context.Context, cveID string) (*EPSSScore, error) {
	url := fmt.Sprintf("%s/epss?cve=%s", c.BaseURL, cveID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch EPSS data: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("EPSS API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var epssResp EPSSResponse
	if err := json.NewDecoder(resp.Body).Decode(&epssResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(epssResp.Data) == 0 {
		return nil, fmt.Errorf("no EPSS data found for CVE %s", cveID)
	}
	
	data := epssResp.Data[0]
	
	// Parse EPSS score
	var epss, percentile float64
	fmt.Sscanf(data.EPSS, "%f", &epss)
	fmt.Sscanf(data.Percentile, "%f", &percentile)
	
	return &EPSSScore{
		CVE:        data.CVE,
		EPSS:       epss,
		Percentile: percentile,
		Date:       data.Date,
	}, nil
}

// GetBulkScores fetches EPSS scores for multiple CVEs
func (c *Client) GetBulkScores(ctx context.Context, cveIDs []string) (map[string]*EPSSScore, error) {
	if len(cveIDs) == 0 {
		return make(map[string]*EPSSScore), nil
	}
	
	// EPSS API supports bulk queries with comma-separated CVEs
	// But we'll batch them to avoid URL length limits
	batchSize := 50
	results := make(map[string]*EPSSScore)
	
	for i := 0; i < len(cveIDs); i += batchSize {
		end := i + batchSize
		if end > len(cveIDs) {
			end = len(cveIDs)
		}
		
		batch := cveIDs[i:end]
		batchResults, err := c.fetchBatch(ctx, batch)
		if err != nil {
			fmt.Printf("Warning: failed to fetch batch %d-%d: %v\n", i, end, err)
			continue
		}
		
		for cve, score := range batchResults {
			results[cve] = score
		}
	}
	
	return results, nil
}

func (c *Client) fetchBatch(ctx context.Context, cveIDs []string) (map[string]*EPSSScore, error) {
	// Build comma-separated CVE list
	cveList := ""
	for i, cve := range cveIDs {
		if i > 0 {
			cveList += ","
		}
		cveList += cve
	}
	
	url := fmt.Sprintf("%s/epss?cve=%s", c.BaseURL, cveList)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	var epssResp EPSSResponse
	if err := json.NewDecoder(resp.Body).Decode(&epssResp); err != nil {
		return nil, err
	}
	
	results := make(map[string]*EPSSScore)
	for _, data := range epssResp.Data {
		var epss, percentile float64
		fmt.Sscanf(data.EPSS, "%f", &epss)
		fmt.Sscanf(data.Percentile, "%f", &percentile)
		
		results[data.CVE] = &EPSSScore{
			CVE:        data.CVE,
			EPSS:       epss,
			Percentile: percentile,
			Date:       data.Date,
		}
	}
	
	return results, nil
}
