// Package storage provides data persistence abstractions for Firestore and GCS.
package storage

import (
	"context"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
)

// MockAlertStore is a mock implementation of AlertStore for testing.
type MockAlertStore struct {
	// SavePoliceAlertsFunc is called when SavePoliceAlerts is invoked.
	// If nil, returns no error.
	SavePoliceAlertsFunc func(ctx context.Context, alerts []models.WazeAlert, scrapeTime time.Time) error

	// GetPoliceAlertsByDateRangeFunc is called when GetPoliceAlertsByDateRange is invoked.
	// If nil, returns empty slice with no error.
	GetPoliceAlertsByDateRangeFunc func(ctx context.Context, startDate, endDate time.Time) ([]models.PoliceAlert, error)

	// GetPoliceAlertsByDatesWithFiltersFunc is called when GetPoliceAlertsByDatesWithFilters is invoked.
	// If nil, returns empty slice with no error.
	GetPoliceAlertsByDatesWithFiltersFunc func(ctx context.Context, dates []string, subtypes []string, streets []string) ([]models.PoliceAlert, error)

	// CloseFunc is called when Close is invoked.
	// If nil, returns no error.
	CloseFunc func() error

	// CallLog tracks calls made to the mock for verification.
	CallLog struct {
		SavePoliceAlertsCalls                  int
		GetPoliceAlertsByDateRangeCalls        int
		GetPoliceAlertsByDatesWithFiltersCalls int
		CloseCalls                             int
		LastSaveAlertsCount                    int
		LastGetDateRangeArgs                   []time.Time
		LastGetDatesWithFiltersArgs            []string
	}
}

// SavePoliceAlerts implements AlertStore.SavePoliceAlerts.
func (m *MockAlertStore) SavePoliceAlerts(ctx context.Context, alerts []models.WazeAlert, scrapeTime time.Time) error {
	m.CallLog.SavePoliceAlertsCalls++
	m.CallLog.LastSaveAlertsCount = len(alerts)

	if m.SavePoliceAlertsFunc != nil {
		return m.SavePoliceAlertsFunc(ctx, alerts, scrapeTime)
	}
	return nil
}

// GetPoliceAlertsByDateRange implements AlertStore.GetPoliceAlertsByDateRange.
func (m *MockAlertStore) GetPoliceAlertsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.PoliceAlert, error) {
	m.CallLog.GetPoliceAlertsByDateRangeCalls++
	m.CallLog.LastGetDateRangeArgs = []time.Time{startDate, endDate}

	if m.GetPoliceAlertsByDateRangeFunc != nil {
		return m.GetPoliceAlertsByDateRangeFunc(ctx, startDate, endDate)
	}
	return []models.PoliceAlert{}, nil
}

// GetPoliceAlertsByDatesWithFilters implements AlertStore.GetPoliceAlertsByDatesWithFilters.
func (m *MockAlertStore) GetPoliceAlertsByDatesWithFilters(ctx context.Context, dates []string, subtypes []string, streets []string) ([]models.PoliceAlert, error) {
	m.CallLog.GetPoliceAlertsByDatesWithFiltersCalls++
	m.CallLog.LastGetDatesWithFiltersArgs = dates

	if m.GetPoliceAlertsByDatesWithFiltersFunc != nil {
		return m.GetPoliceAlertsByDatesWithFiltersFunc(ctx, dates, subtypes, streets)
	}
	return []models.PoliceAlert{}, nil
}

// Close implements AlertStore.Close.
func (m *MockAlertStore) Close() error {
	m.CallLog.CloseCalls++

	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Ensure MockAlertStore implements AlertStore.
var _ AlertStore = (*MockAlertStore)(nil)
