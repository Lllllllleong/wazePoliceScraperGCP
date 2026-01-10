// Package waze provides a client for interacting with the Waze live traffic API.
package waze

import "github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"

// MockAlertFetcher is a mock implementation of AlertFetcher for testing.
type MockAlertFetcher struct {
	// GetAlertsFunc is called when GetAlerts is invoked.
	// If nil, returns empty response with no error.
	GetAlertsFunc func(bbox string) (*models.WazeAPIResponse, error)

	// GetAlertsMultipleBBoxesFunc is called when GetAlertsMultipleBBoxes is invoked.
	// If nil, returns empty slice with no error.
	GetAlertsMultipleBBoxesFunc func(bboxes []string) ([]models.WazeAlert, error)

	// GetStatsFunc is called when GetStats is invoked.
	// If nil, returns default empty stats.
	GetStatsFunc func() *models.ScrapingStats
}

// GetAlerts implements AlertFetcher.GetAlerts.
func (m *MockAlertFetcher) GetAlerts(bbox string) (*models.WazeAPIResponse, error) {
	if m.GetAlertsFunc != nil {
		return m.GetAlertsFunc(bbox)
	}
	return &models.WazeAPIResponse{Alerts: []models.WazeAlert{}}, nil
}

// GetAlertsMultipleBBoxes implements AlertFetcher.GetAlertsMultipleBBoxes.
func (m *MockAlertFetcher) GetAlertsMultipleBBoxes(bboxes []string) ([]models.WazeAlert, error) {
	if m.GetAlertsMultipleBBoxesFunc != nil {
		return m.GetAlertsMultipleBBoxesFunc(bboxes)
	}
	return []models.WazeAlert{}, nil
}

// GetStats implements AlertFetcher.GetStats.
func (m *MockAlertFetcher) GetStats() *models.ScrapingStats {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc()
	}
	return &models.ScrapingStats{}
}

// Ensure MockAlertFetcher implements AlertFetcher.
var _ AlertFetcher = (*MockAlertFetcher)(nil)
