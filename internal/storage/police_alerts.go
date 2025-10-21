package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
	"google.golang.org/genproto/googleapis/type/latlng"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SavePoliceAlerts processes and saves POLICE type alerts with lifecycle tracking
// For new alerts: Initializes all tracking fields
// For existing alerts: Updates only lifecycle/tracking fields
func (fc *FirestoreClient) SavePoliceAlerts(ctx context.Context, alerts []models.WazeAlert, scrapeTime time.Time) error {
	// Filter for POLICE type only
	policeAlerts := make([]models.WazeAlert, 0)
	for _, alert := range alerts {
		if alert.Type == "POLICE" {
			policeAlerts = append(policeAlerts, alert)
		}
	}

	if len(policeAlerts) == 0 {
		log.Println("No POLICE alerts to save")
		return nil
	}

	log.Printf("Processing %d POLICE alerts", len(policeAlerts))

	// Process each alert
	for _, alert := range policeAlerts {
		if err := fc.processPoliceAlert(ctx, alert, scrapeTime); err != nil {
			log.Printf("Error processing alert %s: %v", alert.UUID, err)
			// Continue processing other alerts
			continue
		}
	}

	log.Printf("Successfully processed %d POLICE alerts", len(policeAlerts))
	return nil
}

// processPoliceAlert handles a single police alert (new or existing)
func (fc *FirestoreClient) processPoliceAlert(ctx context.Context, alert models.WazeAlert, scrapeTime time.Time) error {
	docRef := fc.client.Collection(fc.collectionName).Doc(alert.UUID)

	// Check if alert already exists
	docSnap, err := docRef.Get(ctx)
	if err != nil && status.Code(err) != codes.NotFound {
		return fmt.Errorf("failed to check if alert exists: %w", err)
	}

	// Convert alert to JSON for raw data storage
	rawJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert to JSON: %w", err)
	}
	rawJSONStr := string(rawJSON)

	// Convert pubMillis to time.Time
	publishTime := time.UnixMilli(alert.PubMillis)

	// Calculate lastVerificationMillis from comments
	lastVerificationMillis, lastVerificationTime := extractLastVerification(alert.Comments)

	if !docSnap.Exists() {
		// NEW ALERT - Initialize all fields
		log.Printf("New POLICE alert: %s", alert.UUID)

		policeAlert := models.PoliceAlert{
			// Core data
			UUID:    alert.UUID,
			ID:      alert.ID,
			Type:    alert.Type,
			Subtype: alert.Subtype,
			Street:  alert.Street,
			City:    alert.City,
			Country: alert.Country,

			// Location as LatLng for geospatial queries
			LocationGeo: &latlng.LatLng{
				Latitude:  alert.Location.Latitude,
				Longitude: alert.Location.Longitude,
			},

			// Reliability
			Reliability:  alert.Reliability,
			Confidence:   alert.Confidence,
			ReportRating: alert.ReportRating,

			// Time tracking
			PublishTime: publishTime,
			ScrapeTime:  scrapeTime, // First time seen
			ExpireTime:  scrapeTime, // Initialize as scrape time (will be updated on next scrape)

			// Verification
			LastVerificationTime:   lastVerificationTime,
			LastVerificationMillis: lastVerificationMillis,

			// Duration (initially 0, will be calculated on next update)
			ActiveMillis: 0,

			// Community engagement
			NThumbsUpInitial: alert.NThumbsUp,
			NThumbsUpLast:    alert.NThumbsUp,

			// Raw data
			RawDataInitial: rawJSONStr,
			RawDataLast:    rawJSONStr,
		}

		// Save to Firestore
		_, err = docRef.Set(ctx, policeAlert)
		if err != nil {
			return fmt.Errorf("failed to create new police alert: %w", err)
		}

		log.Printf("Created new alert %s in %s, %s", alert.UUID, alert.City, alert.Country)

	} else {
		// EXISTING ALERT - Update only tracking fields
		log.Printf("Updating existing POLICE alert: %s", alert.UUID)

		// Calculate activeMillis: current scrapeTime - original publishTime
		expireMillis := scrapeTime.UnixMilli()
		activeMillis := expireMillis - alert.PubMillis

		updates := []firestore.Update{
			{Path: "expire_time", Value: scrapeTime},
			{Path: "active_millis", Value: activeMillis},
			{Path: "n_thumbs_up_last", Value: alert.NThumbsUp},
			{Path: "raw_data_last", Value: rawJSONStr},
		}

		// Update verification fields if there are comments
		if lastVerificationMillis != nil {
			updates = append(updates,
				firestore.Update{Path: "last_verification_millis", Value: lastVerificationMillis},
				firestore.Update{Path: "last_verification_time", Value: lastVerificationTime},
			)
		}

		_, err = docRef.Update(ctx, updates)
		if err != nil {
			return fmt.Errorf("failed to update police alert: %w", err)
		}

		log.Printf("Updated alert %s (active for %d ms)", alert.UUID, activeMillis)
	}

	return nil
}

// extractLastVerification finds the latest reportMillis from comments
// Returns nil if no comments or empty comments array
func extractLastVerification(comments []models.Comment) (*int64, *time.Time) {
	if len(comments) == 0 {
		return nil, nil
	}

	var maxMillis int64 = 0
	for _, comment := range comments {
		if comment.ReportMillis > maxMillis {
			maxMillis = comment.ReportMillis
		}
	}

	if maxMillis == 0 {
		return nil, nil
	}

	verificationTime := time.UnixMilli(maxMillis)
	return &maxMillis, &verificationTime
}

// GetPoliceAlertsByDateRange retrieves police alerts that were active within a date range
// An alert is considered active if: expire_time >= startDate AND publish_time <= endDate
// This captures all alerts whose lifecycle overlaps with the specified date range
func (fc *FirestoreClient) GetPoliceAlertsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.PoliceAlert, error) {
	log.Printf("Querying police alerts active from %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	query := fc.client.Collection(fc.collectionName).
		Where("expire_time", ">=", startDate).
		Where("publish_time", "<=", endDate).
		OrderBy("expire_time", firestore.Asc).
		OrderBy("publish_time", firestore.Asc)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to query police alerts: %w", err)
	}

	alerts := make([]models.PoliceAlert, 0, len(docs))
	for _, doc := range docs {
		var alert models.PoliceAlert
		if err := doc.DataTo(&alert); err != nil {
			log.Printf("Failed to parse alert %s: %v", doc.Ref.ID, err)
			continue
		}
		alerts = append(alerts, alert)
	}

	log.Printf("Retrieved %d police alerts from Firestore", len(alerts))
	return alerts, nil
}

// GetPoliceAlertsByDatesWithFilters retrieves police alerts for multiple specific dates with optional filters
// Each date should be in YYYY-MM-DD format. The function queries alerts active on each date
// and applies optional subtype and street filters.
func (fc *FirestoreClient) GetPoliceAlertsByDatesWithFilters(ctx context.Context, dates []string, subtypes []string, streets []string) ([]models.PoliceAlert, error) {
	if len(dates) == 0 {
		return nil, fmt.Errorf("at least one date is required")
	}

	log.Printf("Querying police alerts for %d dates with filters (subtypes: %v, streets: %v)", len(dates), subtypes, streets)

	// Use a map to deduplicate alerts by UUID across multiple date queries
	alertsMap := make(map[string]models.PoliceAlert)

	// Query alerts for each date
	for _, dateStr := range dates {
		// Parse the date string (YYYY-MM-DD) explicitly in UTC to avoid timezone issues
		dayStart, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
		if err != nil {
			log.Printf("Invalid date format '%s': %v", dateStr, err)
			continue
		}

		// Set end time to end of day (still in UTC)
		dayEnd := dayStart.Add(24*time.Hour - time.Second)

		log.Printf("Querying alerts for %s (expire_time >= %s AND publish_time <= %s)",
			dateStr, dayStart.Format("2006-01-02 15:04:05"), dayEnd.Format("2006-01-02 15:04:05"))

		// Query alerts where:
		// - expire_time >= start of day (alert is still active at start of day)
		// - publish_time <= end of day (alert was published by end of day)
		query := fc.client.Collection(fc.collectionName).
			Where("expire_time", ">=", dayStart).
			Where("publish_time", "<=", dayEnd)

		docs, err := query.Documents(ctx).GetAll()
		if err != nil {
			log.Printf("Failed to query police alerts for %s: %v", dateStr, err)
			continue
		}

		log.Printf("Retrieved %d documents for %s", len(docs), dateStr)

		// Process documents and deduplicate
		for _, doc := range docs {
			var alert models.PoliceAlert
			if err := doc.DataTo(&alert); err != nil {
				log.Printf("Failed to parse alert %s: %v", doc.Ref.ID, err)
				continue
			}

			// Apply filters
			if len(subtypes) > 0 && !contains(subtypes, alert.Subtype) {
				continue
			}

			if len(streets) > 0 && !contains(streets, alert.Street) {
				continue
			}

			// Add to map (deduplicates by UUID)
			if _, exists := alertsMap[alert.UUID]; !exists {
				alertsMap[alert.UUID] = alert
			}
		}
	}

	// Convert map to slice
	alerts := make([]models.PoliceAlert, 0, len(alertsMap))
	for _, alert := range alertsMap {
		alerts = append(alerts, alert)
	}

	log.Printf("Retrieved %d unique police alerts from Firestore after filtering", len(alerts))
	return alerts, nil
}

// contains checks if a string slice contains a specific value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
