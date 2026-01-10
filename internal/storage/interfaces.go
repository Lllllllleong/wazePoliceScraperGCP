// Package storage provides data persistence abstractions for Firestore and GCS.
package storage

import (
	"context"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
)

// AlertStore defines the interface for police alert storage operations.
// This interface enables dependency injection and mocking for testing.
type AlertStore interface {
	// SavePoliceAlerts processes and saves POLICE type alerts with lifecycle tracking.
	// For new alerts: Initializes all tracking fields.
	// For existing alerts: Updates only lifecycle/tracking fields.
	SavePoliceAlerts(ctx context.Context, alerts []models.WazeAlert, scrapeTime time.Time) error

	// GetPoliceAlertsByDateRange retrieves police alerts that were active within a date range.
	// An alert is considered active if: expire_time >= startDate AND publish_time <= endDate.
	GetPoliceAlertsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.PoliceAlert, error)

	// GetPoliceAlertsByDatesWithFilters retrieves police alerts for multiple specific dates with optional filters.
	// Each date should be in YYYY-MM-DD format.
	GetPoliceAlertsByDatesWithFilters(ctx context.Context, dates []string, subtypes []string, streets []string) ([]models.PoliceAlert, error)

	// Close closes the underlying storage client.
	Close() error
}

// Ensure FirestoreClient implements AlertStore interface.
var _ AlertStore = (*FirestoreClient)(nil)
