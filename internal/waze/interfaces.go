// Package waze provides a client for interacting with the Waze live traffic API.
package waze

import "github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"

// AlertFetcher defines the interface for fetching alerts from Waze.
// This interface enables dependency injection and mocking for testing.
type AlertFetcher interface {
	// GetAlerts fetches alerts from Waze API for a single bounding box.
	// bbox format: "west,south,east,north" (e.g., "103.6,1.15,104.0,1.45").
	GetAlerts(bbox string) (*models.WazeAPIResponse, error)

	// GetAlertsMultipleBBoxes fetches alerts from multiple bounding boxes and deduplicates.
	GetAlertsMultipleBBoxes(bboxes []string) ([]models.WazeAlert, error)

	// GetStats returns scraping statistics.
	GetStats() *models.ScrapingStats
}

// Ensure Client implements AlertFetcher interface.
var _ AlertFetcher = (*Client)(nil)
