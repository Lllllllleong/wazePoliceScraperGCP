package waze

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
)

// Client handles API calls to Waze
type Client struct {
	httpClient *http.Client
	stats      *models.ScrapingStats
}

// NewClient creates a new Waze API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		stats: &models.ScrapingStats{},
	}
}

// GetAlerts fetches alerts from Waze API for a single bounding box
// bbox format: "west,south,east,north" (e.g., "103.6,1.15,104.0,1.45")
func (c *Client) GetAlerts(bbox string) (*models.WazeAPIResponse, error) {
	c.stats.TotalRequests++

	// Parse bounding box: "west,south,east,north"
	parts := strings.Split(bbox, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid bounding box format: %s (expected: west,south,east,north)", bbox)
	}

	west, south, east, north := parts[0], parts[1], parts[2], parts[3]

	// Try multiple potential URL formats
	potentialURLs := []string{
		fmt.Sprintf("https://www.waze.com/live-map/api/georss?top=%s&bottom=%s&left=%s&right=%s&env=row&types=alerts",
			north, south, west, east),
		fmt.Sprintf("https://www.waze.com/live-map/api/georss?bbox=%s&types=alerts", bbox),
	}

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Referer":    "https://www.waze.com/live-map",
		"Accept":     "application/json, text/plain, */*",
	}

	for i, url := range potentialURLs {
		log.Printf("Attempting direct API call %d to: %s", i+1, url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			continue
		}

		// Set headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Printf("Direct API call failed: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			log.Printf("Successful API call: %d", resp.StatusCode)
			c.stats.SuccessfulCalls++

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read response body: %v", err)
				continue
			}

			var apiResponse models.WazeGeoRSSResponse
			if err := json.Unmarshal(body, &apiResponse); err != nil {
				log.Printf("Failed to parse JSON response: %v", err)
				log.Printf("Raw response (first 500 chars): %s", string(body[:min(500, len(body))]))
				continue
			}

			c.stats.TotalAlerts += len(apiResponse.Alerts)
			c.stats.LastSuccessfulRun = time.Now()

			log.Printf("Successfully fetched %d alerts", len(apiResponse.Alerts))
			return &apiResponse, nil
		}

		log.Printf("API call failed with status: %d", resp.StatusCode)
	}

	c.stats.FailedCalls++
	return nil, fmt.Errorf("all API call attempts failed for bbox: %s", bbox)
}

// GetAlertsMultipleBBoxes fetches alerts from multiple bounding boxes and deduplicates
func (c *Client) GetAlertsMultipleBBoxes(bboxes []string) ([]models.WazeAlert, error) {
	uniqueAlerts := make(map[string]models.WazeAlert)
	successfulCalls := 0

	for i, bbox := range bboxes {
		log.Printf("Fetching alerts for bbox %d/%d: %s", i+1, len(bboxes), bbox)

		result, err := c.GetAlerts(bbox)
		if err != nil {
			log.Printf("API call %d failed for bbox: %s, error: %v", i+1, bbox, err)
			continue
		}

		successfulCalls++
		log.Printf("API call %d successful, found %d alerts", i+1, len(result.Alerts))

		// Add alerts to collection, deduplicating by UUID
		for _, alert := range result.Alerts {
			if alert.UUID != "" {
				if _, exists := uniqueAlerts[alert.UUID]; !exists {
					// Add scraped timestamp
					alert.ScrapedAt = time.Now()
					uniqueAlerts[alert.UUID] = alert
				} else {
					log.Printf("Duplicate alert found across bboxes: %s", alert.UUID)
				}
			}
		}
	}

	if successfulCalls == 0 {
		return nil, fmt.Errorf("no successful API calls from %d attempts", len(bboxes))
	}

	// Convert map to slice
	allAlerts := make([]models.WazeAlert, 0, len(uniqueAlerts))
	for _, alert := range uniqueAlerts {
		allAlerts = append(allAlerts, alert)
	}

	c.stats.UniqueAlerts = len(allAlerts)

	log.Printf("Combined results: %d successful calls, %d total alerts, %d unique alerts",
		successfulCalls, c.stats.TotalAlerts, len(allAlerts))

	return allAlerts, nil
}

// GetStats returns scraping statistics
func (c *Client) GetStats() *models.ScrapingStats {
	return c.stats
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
