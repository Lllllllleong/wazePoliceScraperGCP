//go:build integration

// Package storage integration tests require a Firestore emulator.
//
// To run these tests:
//
//  1. Start the Firestore emulator:
//     gcloud emulators firestore start --host-port=localhost:8080
//
//  2. Set the environment variable:
//     export FIRESTORE_EMULATOR_HOST=localhost:8080
//
//  3. Run the tests:
//     go test -tags=integration ./internal/storage/... -v
package storage

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
)

const testProjectID = "test-project"

// testHelper provides setup and teardown for integration tests
type testHelper struct {
	client         *FirestoreClient
	collectionName string
	ctx            context.Context
	t              *testing.T
}

// newTestHelper creates a new test helper with a unique collection
func newTestHelper(t *testing.T) *testHelper {
	t.Helper()

	// Verify emulator is configured
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST not set, skipping integration test")
	}

	ctx := context.Background()

	// Use unique collection name to isolate tests
	collectionName := fmt.Sprintf("test_alerts_%d", time.Now().UnixNano())

	client, err := NewFirestoreClient(ctx, testProjectID, collectionName)
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	return &testHelper{
		client:         client,
		collectionName: collectionName,
		ctx:            ctx,
		t:              t,
	}
}

// cleanup deletes all documents in the test collection and closes the client
func (h *testHelper) cleanup() {
	h.t.Helper()

	// Delete all documents in the collection
	docs, err := h.client.client.Collection(h.collectionName).Documents(h.ctx).GetAll()
	if err == nil {
		for _, doc := range docs {
			_, _ = doc.Ref.Delete(h.ctx)
		}
	}

	h.client.Close()
}

// createTestWazeAlert creates a WazeAlert with sensible defaults
func createTestWazeAlert(uuid string, alertType string, overrides map[string]interface{}) models.WazeAlert {
	alert := models.WazeAlert{
		UUID:    uuid,
		Type:    alertType,
		Subtype: "POLICE_VISIBLE",
		Location: models.Location{
			Latitude:  -33.8688,
			Longitude: 151.2093,
		},
		Street:      "George Street",
		City:        "Sydney",
		Country:     "AU",
		Reliability: 8,
		Confidence:  7,
		PubMillis:   time.Now().Add(-1 * time.Hour).UnixMilli(),
		NThumbsUp:   5,
	}

	// Apply overrides
	if v, ok := overrides["Subtype"].(string); ok {
		alert.Subtype = v
	}
	if v, ok := overrides["Street"].(string); ok {
		alert.Street = v
	}
	if v, ok := overrides["City"].(string); ok {
		alert.City = v
	}
	if v, ok := overrides["Country"].(string); ok {
		alert.Country = v
	}
	if v, ok := overrides["PubMillis"].(int64); ok {
		alert.PubMillis = v
	}
	if v, ok := overrides["NThumbsUp"].(int); ok {
		alert.NThumbsUp = v
	}
	if v, ok := overrides["Comments"].([]models.Comment); ok {
		alert.Comments = v
	}

	return alert
}

// =============================================================================
// SavePoliceAlerts Tests
// =============================================================================

func TestIntegration_SavePoliceAlerts_NewAlert(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Create a new POLICE alert
	alerts := []models.WazeAlert{
		createTestWazeAlert("new-alert-001", "POLICE", map[string]interface{}{
			"Subtype":   "POLICE_WITH_MOBILE_CAMERA",
			"Street":    "Hume Highway",
			"NThumbsUp": 3,
		}),
	}

	scrapeTime := time.Now()

	// Save the alert
	err := h.client.SavePoliceAlerts(h.ctx, alerts, scrapeTime)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Verify alert was saved
	doc, err := h.client.client.Collection(h.collectionName).Doc("new-alert-001").Get(h.ctx)
	if err != nil {
		t.Fatalf("Failed to get saved alert: %v", err)
	}

	if !doc.Exists() {
		t.Fatal("Expected alert to exist")
	}

	// Verify fields
	data := doc.Data()
	if data["uuid"] != "new-alert-001" {
		t.Errorf("Expected UUID 'new-alert-001', got %v", data["uuid"])
	}
	if data["type"] != "POLICE" {
		t.Errorf("Expected type 'POLICE', got %v", data["type"])
	}
	if data["subtype"] != "POLICE_WITH_MOBILE_CAMERA" {
		t.Errorf("Expected subtype 'POLICE_WITH_MOBILE_CAMERA', got %v", data["subtype"])
	}
	if data["street"] != "Hume Highway" {
		t.Errorf("Expected street 'Hume Highway', got %v", data["street"])
	}
	// Initial thumbs up should be captured
	if data["n_thumbs_up_initial"] != int64(3) {
		t.Errorf("Expected n_thumbs_up_initial 3, got %v", data["n_thumbs_up_initial"])
	}
	if data["n_thumbs_up_last"] != int64(3) {
		t.Errorf("Expected n_thumbs_up_last 3, got %v", data["n_thumbs_up_last"])
	}
}

func TestIntegration_SavePoliceAlerts_FiltersNonPolice(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Create mixed alert types
	alerts := []models.WazeAlert{
		createTestWazeAlert("police-001", "POLICE", nil),
		createTestWazeAlert("accident-001", "ACCIDENT", nil),
		createTestWazeAlert("police-002", "POLICE", nil),
		createTestWazeAlert("jam-001", "JAM", nil),
		createTestWazeAlert("hazard-001", "HAZARD", nil),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, time.Now())
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Count documents in collection
	docs, err := h.client.client.Collection(h.collectionName).Documents(h.ctx).GetAll()
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	// Should only have 2 POLICE alerts
	if len(docs) != 2 {
		t.Errorf("Expected 2 documents (POLICE only), got %d", len(docs))
	}

	// Verify the correct ones were saved
	savedUUIDs := make(map[string]bool)
	for _, doc := range docs {
		savedUUIDs[doc.Ref.ID] = true
	}

	if !savedUUIDs["police-001"] {
		t.Error("Expected police-001 to be saved")
	}
	if !savedUUIDs["police-002"] {
		t.Error("Expected police-002 to be saved")
	}
	if savedUUIDs["accident-001"] {
		t.Error("accident-001 should not be saved")
	}
}

func TestIntegration_SavePoliceAlerts_UpdatesExisting(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	pubTime := time.Now().Add(-2 * time.Hour)

	// First scrape - create the alert
	alerts1 := []models.WazeAlert{
		createTestWazeAlert("update-test-001", "POLICE", map[string]interface{}{
			"PubMillis": pubTime.UnixMilli(),
			"NThumbsUp": 5,
		}),
	}

	scrapeTime1 := time.Now().Add(-1 * time.Hour)
	err := h.client.SavePoliceAlerts(h.ctx, alerts1, scrapeTime1)
	if err != nil {
		t.Fatalf("First SavePoliceAlerts failed: %v", err)
	}

	// Second scrape - update the alert with new thumbs up
	alerts2 := []models.WazeAlert{
		createTestWazeAlert("update-test-001", "POLICE", map[string]interface{}{
			"PubMillis": pubTime.UnixMilli(),
			"NThumbsUp": 12, // Increased from 5
		}),
	}

	scrapeTime2 := time.Now()
	err = h.client.SavePoliceAlerts(h.ctx, alerts2, scrapeTime2)
	if err != nil {
		t.Fatalf("Second SavePoliceAlerts failed: %v", err)
	}

	// Verify the alert was updated, not duplicated
	docs, err := h.client.client.Collection(h.collectionName).Documents(h.ctx).GetAll()
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("Expected 1 document (upsert), got %d", len(docs))
	}

	// Verify update fields changed but initial fields preserved
	doc, _ := h.client.client.Collection(h.collectionName).Doc("update-test-001").Get(h.ctx)
	data := doc.Data()

	// Initial should be preserved
	if data["n_thumbs_up_initial"] != int64(5) {
		t.Errorf("Expected n_thumbs_up_initial to stay 5, got %v", data["n_thumbs_up_initial"])
	}

	// Last should be updated
	if data["n_thumbs_up_last"] != int64(12) {
		t.Errorf("Expected n_thumbs_up_last to be 12, got %v", data["n_thumbs_up_last"])
	}

	// Active millis should be calculated
	activeMillis, ok := data["active_millis"].(int64)
	if !ok || activeMillis <= 0 {
		t.Errorf("Expected positive active_millis, got %v", data["active_millis"])
	}
}

func TestIntegration_SavePoliceAlerts_EmptyList(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Save empty list - should not error
	err := h.client.SavePoliceAlerts(h.ctx, []models.WazeAlert{}, time.Now())
	if err != nil {
		t.Fatalf("SavePoliceAlerts with empty list should not error: %v", err)
	}

	// Verify no documents created
	docs, _ := h.client.client.Collection(h.collectionName).Documents(h.ctx).GetAll()
	if len(docs) != 0 {
		t.Errorf("Expected 0 documents, got %d", len(docs))
	}
}

func TestIntegration_SavePoliceAlerts_WithVerificationComments(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	verificationTime := time.Now().Add(-30 * time.Minute).UnixMilli()

	alerts := []models.WazeAlert{
		createTestWazeAlert("verified-001", "POLICE", map[string]interface{}{
			"Comments": []models.Comment{
				{ReportMillis: verificationTime, Text: "Still there!", IsThumbsUp: true},
			},
		}),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, time.Now())
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Verify verification fields were set
	doc, _ := h.client.client.Collection(h.collectionName).Doc("verified-001").Get(h.ctx)
	data := doc.Data()

	if data["last_verification_millis"] != verificationTime {
		t.Errorf("Expected last_verification_millis %d, got %v", verificationTime, data["last_verification_millis"])
	}

	if data["last_verification_time"] == nil {
		t.Error("Expected last_verification_time to be set")
	}
}

// =============================================================================
// GetPoliceAlertsByDateRange Tests
// =============================================================================

func TestIntegration_GetPoliceAlertsByDateRange_ReturnsActiveAlerts(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Set up test data with known times
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	// Create alerts with specific publish times
	alerts := []models.WazeAlert{
		// Alert published yesterday, still active (should be returned)
		createTestWazeAlert("active-yesterday", "POLICE", map[string]interface{}{
			"PubMillis": yesterday.UnixMilli(),
		}),
		// Alert published two days ago, still active (should be returned)
		createTestWazeAlert("active-two-days", "POLICE", map[string]interface{}{
			"PubMillis": twoDaysAgo.UnixMilli(),
		}),
	}

	// Save alerts (sets expire_time to scrape time)
	err := h.client.SavePoliceAlerts(h.ctx, alerts, now)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Query for alerts active yesterday
	startDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(24*time.Hour - time.Second)

	results, err := h.client.GetPoliceAlertsByDateRange(h.ctx, startDate, endDate)
	if err != nil {
		t.Fatalf("GetPoliceAlertsByDateRange failed: %v", err)
	}

	// Both alerts should be returned (both were active yesterday)
	if len(results) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(results))
	}
}

func TestIntegration_GetPoliceAlertsByDateRange_EmptyResult(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Create an alert for today
	alerts := []models.WazeAlert{
		createTestWazeAlert("today-alert", "POLICE", map[string]interface{}{
			"PubMillis": time.Now().UnixMilli(),
		}),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, time.Now())
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Query for alerts from a week ago (should return empty)
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	startDate := time.Date(weekAgo.Year(), weekAgo.Month(), weekAgo.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(24*time.Hour - time.Second)

	results, err := h.client.GetPoliceAlertsByDateRange(h.ctx, startDate, endDate)
	if err != nil {
		t.Fatalf("GetPoliceAlertsByDateRange failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 alerts for old date range, got %d", len(results))
	}
}

// =============================================================================
// GetPoliceAlertsByDatesWithFilters Tests
// =============================================================================

func TestIntegration_GetPoliceAlertsByDatesWithFilters_SubtypeFilter(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	now := time.Now()

	// Create alerts with different subtypes
	alerts := []models.WazeAlert{
		createTestWazeAlert("visible-001", "POLICE", map[string]interface{}{
			"Subtype":   "POLICE_VISIBLE",
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
		createTestWazeAlert("camera-001", "POLICE", map[string]interface{}{
			"Subtype":   "POLICE_WITH_MOBILE_CAMERA",
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
		createTestWazeAlert("hiding-001", "POLICE", map[string]interface{}{
			"Subtype":   "POLICE_HIDING",
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, now)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Query with subtype filter
	today := now.Format("2006-01-02")
	results, err := h.client.GetPoliceAlertsByDatesWithFilters(
		h.ctx,
		[]string{today},
		[]string{"POLICE_VISIBLE", "POLICE_HIDING"}, // Filter to 2 subtypes
		[]string{}, // No street filter
	)
	if err != nil {
		t.Fatalf("GetPoliceAlertsByDatesWithFilters failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 alerts (VISIBLE and HIDING), got %d", len(results))
	}

	// Verify correct subtypes returned
	subtypes := make(map[string]bool)
	for _, alert := range results {
		subtypes[alert.Subtype] = true
	}

	if !subtypes["POLICE_VISIBLE"] {
		t.Error("Expected POLICE_VISIBLE in results")
	}
	if !subtypes["POLICE_HIDING"] {
		t.Error("Expected POLICE_HIDING in results")
	}
	if subtypes["POLICE_WITH_MOBILE_CAMERA"] {
		t.Error("POLICE_WITH_MOBILE_CAMERA should not be in results")
	}
}

func TestIntegration_GetPoliceAlertsByDatesWithFilters_StreetFilter(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	now := time.Now()

	// Create alerts on different streets
	alerts := []models.WazeAlert{
		createTestWazeAlert("hume-001", "POLICE", map[string]interface{}{
			"Street":    "Hume Highway",
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
		createTestWazeAlert("federal-001", "POLICE", map[string]interface{}{
			"Street":    "Federal Highway",
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
		createTestWazeAlert("george-001", "POLICE", map[string]interface{}{
			"Street":    "George Street",
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, now)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Query with street filter
	today := now.Format("2006-01-02")
	results, err := h.client.GetPoliceAlertsByDatesWithFilters(
		h.ctx,
		[]string{today},
		[]string{},               // No subtype filter
		[]string{"Hume Highway"}, // Only Hume Highway
	)
	if err != nil {
		t.Fatalf("GetPoliceAlertsByDatesWithFilters failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 alert (Hume Highway only), got %d", len(results))
	}

	if len(results) > 0 && results[0].Street != "Hume Highway" {
		t.Errorf("Expected street 'Hume Highway', got %q", results[0].Street)
	}
}

func TestIntegration_GetPoliceAlertsByDatesWithFilters_MultipleDates(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Create alert from yesterday
	alertsYesterday := []models.WazeAlert{
		createTestWazeAlert("yesterday-001", "POLICE", map[string]interface{}{
			"PubMillis": yesterday.UnixMilli(),
		}),
	}
	err := h.client.SavePoliceAlerts(h.ctx, alertsYesterday, yesterday.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("SavePoliceAlerts (yesterday) failed: %v", err)
	}

	// Create alert from today
	alertsToday := []models.WazeAlert{
		createTestWazeAlert("today-001", "POLICE", map[string]interface{}{
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
	}
	err = h.client.SavePoliceAlerts(h.ctx, alertsToday, now)
	if err != nil {
		t.Fatalf("SavePoliceAlerts (today) failed: %v", err)
	}

	// Query for both dates
	todayStr := now.Format("2006-01-02")
	yesterdayStr := yesterday.Format("2006-01-02")

	results, err := h.client.GetPoliceAlertsByDatesWithFilters(
		h.ctx,
		[]string{yesterdayStr, todayStr},
		[]string{},
		[]string{},
	)
	if err != nil {
		t.Fatalf("GetPoliceAlertsByDatesWithFilters failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 alerts (one from each day), got %d", len(results))
	}
}

func TestIntegration_GetPoliceAlertsByDatesWithFilters_DeduplicatesAcrossDates(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Create an alert that spans both days
	alerts := []models.WazeAlert{
		createTestWazeAlert("spanning-001", "POLICE", map[string]interface{}{
			"PubMillis": yesterday.UnixMilli(), // Published yesterday
		}),
	}

	// Save with current time (so it's "active" from yesterday to now)
	err := h.client.SavePoliceAlerts(h.ctx, alerts, now)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Query for both dates - alert should appear in both but be deduplicated
	todayStr := now.Format("2006-01-02")
	yesterdayStr := yesterday.Format("2006-01-02")

	results, err := h.client.GetPoliceAlertsByDatesWithFilters(
		h.ctx,
		[]string{yesterdayStr, todayStr},
		[]string{},
		[]string{},
	)
	if err != nil {
		t.Fatalf("GetPoliceAlertsByDatesWithFilters failed: %v", err)
	}

	// Should only get 1 alert (deduplicated by UUID)
	if len(results) != 1 {
		t.Errorf("Expected 1 alert (deduplicated), got %d", len(results))
	}

	if len(results) > 0 && results[0].UUID != "spanning-001" {
		t.Errorf("Expected UUID 'spanning-001', got %q", results[0].UUID)
	}
}

func TestIntegration_GetPoliceAlertsByDatesWithFilters_EmptyDatesError(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	_, err := h.client.GetPoliceAlertsByDatesWithFilters(
		h.ctx,
		[]string{}, // Empty dates
		[]string{},
		[]string{},
	)

	if err == nil {
		t.Error("Expected error for empty dates, got nil")
	}
}

func TestIntegration_GetPoliceAlertsByDatesWithFilters_NoFiltersReturnsAll(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	now := time.Now()

	// Create 3 different alerts
	alerts := []models.WazeAlert{
		createTestWazeAlert("alert-001", "POLICE", map[string]interface{}{
			"PubMillis": now.Add(-1 * time.Hour).UnixMilli(),
		}),
		createTestWazeAlert("alert-002", "POLICE", map[string]interface{}{
			"PubMillis": now.Add(-2 * time.Hour).UnixMilli(),
		}),
		createTestWazeAlert("alert-003", "POLICE", map[string]interface{}{
			"PubMillis": now.Add(-3 * time.Hour).UnixMilli(),
		}),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, now)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Query with no filters
	today := now.Format("2006-01-02")
	results, err := h.client.GetPoliceAlertsByDatesWithFilters(
		h.ctx,
		[]string{today},
		[]string{}, // No subtype filter
		[]string{}, // No street filter
	)
	if err != nil {
		t.Fatalf("GetPoliceAlertsByDatesWithFilters failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 alerts (no filters), got %d", len(results))
	}
}

// =============================================================================
// Edge Cases and Error Handling
// =============================================================================

func TestIntegration_SavePoliceAlerts_ContinuesOnIndividualError(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Create multiple valid alerts - all should be saved even if we had errors
	alerts := []models.WazeAlert{
		createTestWazeAlert("valid-001", "POLICE", nil),
		createTestWazeAlert("valid-002", "POLICE", nil),
		createTestWazeAlert("valid-003", "POLICE", nil),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, time.Now())
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// All 3 should be saved
	docs, _ := h.client.client.Collection(h.collectionName).Documents(h.ctx).GetAll()
	if len(docs) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(docs))
	}
}

func TestIntegration_GetCollectionName(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	if h.client.GetCollectionName() != h.collectionName {
		t.Errorf("Expected collection name %q, got %q", h.collectionName, h.client.GetCollectionName())
	}
}

// =============================================================================
// Data Integrity Tests
// =============================================================================

func TestIntegration_SavePoliceAlerts_PreservesAllFields(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	pubTime := time.Now().Add(-1 * time.Hour)

	alerts := []models.WazeAlert{
		{
			UUID:    "full-fields-001",
			ID:      "waze-id-123",
			Type:    "POLICE",
			Subtype: "POLICE_WITH_MOBILE_CAMERA",
			Location: models.Location{
				Latitude:  -34.7515,
				Longitude: 149.7209,
			},
			Street:       "Hume Highway",
			City:         "Goulburn",
			Country:      "AU",
			Reliability:  9,
			Confidence:   8,
			ReportRating: 5,
			PubMillis:    pubTime.UnixMilli(),
			NThumbsUp:    7,
		},
	}

	scrapeTime := time.Now()
	err := h.client.SavePoliceAlerts(h.ctx, alerts, scrapeTime)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Retrieve and verify all fields
	doc, _ := h.client.client.Collection(h.collectionName).Doc("full-fields-001").Get(h.ctx)
	data := doc.Data()

	// Core fields
	if data["uuid"] != "full-fields-001" {
		t.Errorf("UUID mismatch: got %v", data["uuid"])
	}
	if data["type"] != "POLICE" {
		t.Errorf("Type mismatch: got %v", data["type"])
	}
	if data["subtype"] != "POLICE_WITH_MOBILE_CAMERA" {
		t.Errorf("Subtype mismatch: got %v", data["subtype"])
	}
	if data["street"] != "Hume Highway" {
		t.Errorf("Street mismatch: got %v", data["street"])
	}
	if data["city"] != "Goulburn" {
		t.Errorf("City mismatch: got %v", data["city"])
	}
	if data["country"] != "AU" {
		t.Errorf("Country mismatch: got %v", data["country"])
	}

	// Reliability metrics
	if data["reliability"] != int64(9) {
		t.Errorf("Reliability mismatch: got %v", data["reliability"])
	}
	if data["confidence"] != int64(8) {
		t.Errorf("Confidence mismatch: got %v", data["confidence"])
	}

	// Location should be stored as GeoPoint
	if data["location_geo"] == nil {
		t.Error("location_geo should be set")
	}

	// Raw data should be preserved
	if data["raw_data_initial"] == nil || data["raw_data_initial"] == "" {
		t.Error("raw_data_initial should be set")
	}
}

// ============================================================================
// HIGH PRIORITY: Concurrent Operations & Race Conditions
// ============================================================================

func TestIntegration_SavePoliceAlerts_ConcurrentUpdates(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	pubTime := time.Now().Add(-1 * time.Hour)
	baseAlert := createTestWazeAlert("concurrent-001", "POLICE", map[string]interface{}{
		"PubMillis": pubTime.UnixMilli(),
		"Street":    "Original Street",
		"NThumbsUp": 1,
	})

	// First scrape - create the alert
	scrapeTime1 := time.Now().Add(-30 * time.Minute)
	err := h.client.SavePoliceAlerts(h.ctx, []models.WazeAlert{baseAlert}, scrapeTime1)
	if err != nil {
		t.Fatalf("Initial SavePoliceAlerts failed: %v", err)
	}

	// Launch 10 concurrent updates with different thumbs up counts
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(thumbsUp int) {
			alert := createTestWazeAlert("concurrent-001", "POLICE", map[string]interface{}{
				"PubMillis": pubTime.UnixMilli(),
				"NThumbsUp": thumbsUp,
			})
			scrapeTime := time.Now().Add(time.Duration(thumbsUp) * time.Millisecond)
			err := h.client.SavePoliceAlerts(context.Background(), []models.WazeAlert{alert}, scrapeTime)
			done <- err
		}(i + 2)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent update %d failed: %v", i, err)
		}
	}

	// Verify: Should still have exactly 1 document (no duplication)
	docs, err := h.client.client.Collection(h.collectionName).Documents(h.ctx).GetAll()
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("Expected 1 document after concurrent updates, got %d", len(docs))
	}

	// Verify the document exists and has valid data
	doc, err := h.client.client.Collection(h.collectionName).Doc("concurrent-001").Get(h.ctx)
	if err != nil {
		t.Fatalf("Failed to get final document: %v", err)
	}

	data := doc.Data()

	// Initial thumbs up should be preserved from first write
	if data["n_thumbs_up_initial"] != int64(1) {
		t.Errorf("Expected n_thumbs_up_initial to be 1 (from first write), got %v", data["n_thumbs_up_initial"])
	}

	// Last thumbs up should be from one of the concurrent updates (2-11)
	lastThumbsUp := data["n_thumbs_up_last"].(int64)
	if lastThumbsUp < 2 || lastThumbsUp > 11 {
		t.Errorf("Expected n_thumbs_up_last to be between 2-11 (from concurrent writes), got %v", lastThumbsUp)
	}

	// Active millis should be positive
	activeMillis, ok := data["active_millis"].(int64)
	if !ok || activeMillis <= 0 {
		t.Errorf("Expected positive active_millis, got %v", data["active_millis"])
	}

	t.Logf("Concurrent test passed: final n_thumbs_up_last=%d, active_millis=%d", lastThumbsUp, activeMillis)
}

// ============================================================================
// HIGH PRIORITY: Large Dataset Performance
// ============================================================================

func TestIntegration_SavePoliceAlerts_LargeBatch(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Create 150 alerts in a single batch
	const batchSize = 150
	alerts := make([]models.WazeAlert, batchSize)
	pubTime := time.Now().Add(-30 * time.Minute)

	for i := 0; i < batchSize; i++ {
		alerts[i] = createTestWazeAlert(
			fmt.Sprintf("batch-alert-%03d", i),
			"POLICE",
			map[string]interface{}{
				"PubMillis": pubTime.UnixMilli(),
				"Street":    fmt.Sprintf("Street %d", i%20), // 20 unique streets
				"City":      fmt.Sprintf("City %d", i%5),    // 5 unique cities
				"Subtype":   []string{"POLICE_VISIBLE", "POLICE_HIDING", "POLICE_WITH_MOBILE_CAMERA"}[i%3],
				"NThumbsUp": i % 10,
			},
		)
	}

	// Save the large batch
	scrapeTime := time.Now()
	start := time.Now()
	err := h.client.SavePoliceAlerts(h.ctx, alerts, scrapeTime)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("SavePoliceAlerts with large batch failed: %v", err)
	}

	t.Logf("Saved %d alerts in %v (%.2f alerts/sec)", batchSize, duration, float64(batchSize)/duration.Seconds())

	// Verify all alerts were saved
	docs, err := h.client.client.Collection(h.collectionName).Documents(h.ctx).GetAll()
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	if len(docs) != batchSize {
		t.Errorf("Expected %d documents, got %d", batchSize, len(docs))
	}

	// Verify a sample of alerts
	sampleIDs := []string{"batch-alert-000", "batch-alert-050", "batch-alert-099", "batch-alert-149"}
	for _, id := range sampleIDs {
		doc, err := h.client.client.Collection(h.collectionName).Doc(id).Get(h.ctx)
		if err != nil {
			t.Errorf("Alert %s not found: %v", id, err)
		} else if !doc.Exists() {
			t.Errorf("Alert %s does not exist", id)
		}
	}
}

func TestIntegration_GetPoliceAlertsByDateRange_LargeResult(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Create 500 alerts spread across 7 days
	const numAlerts = 500
	const numDays = 7
	alerts := make([]models.WazeAlert, numAlerts)

	baseTime := time.Now().Add(-7 * 24 * time.Hour)

	for i := 0; i < numAlerts; i++ {
		// Distribute alerts across 7 days
		dayOffset := i % numDays
		pubTime := baseTime.Add(time.Duration(dayOffset) * 24 * time.Hour)

		alerts[i] = createTestWazeAlert(
			fmt.Sprintf("large-result-%03d", i),
			"POLICE",
			map[string]interface{}{
				"PubMillis": pubTime.UnixMilli(),
				"Subtype":   []string{"POLICE_VISIBLE", "POLICE_HIDING"}[i%2],
			},
		)
	}

	// Save all alerts
	scrapeTime := time.Now()
	err := h.client.SavePoliceAlerts(h.ctx, alerts, scrapeTime)
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Query for all 7 days
	startDate := baseTime.Add(-1 * time.Hour) // Just before first alert
	endDate := scrapeTime.Add(1 * time.Hour)  // After last scrape

	start := time.Now()
	results, err := h.client.GetPoliceAlertsByDateRange(h.ctx, startDate, endDate)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("GetPoliceAlertsByDateRange failed: %v", err)
	}

	if len(results) != numAlerts {
		t.Errorf("Expected %d alerts in result, got %d", numAlerts, len(results))
	}

	t.Logf("Retrieved %d alerts in %v (%.2f alerts/sec)", len(results), duration, float64(len(results))/duration.Seconds())

	// Verify no truncation - check for presence of first and last alerts
	foundFirst := false
	foundLast := false
	for _, alert := range results {
		if alert.UUID == "large-result-000" {
			foundFirst = true
		}
		if alert.UUID == "large-result-499" {
			foundLast = true
		}
	}

	if !foundFirst {
		t.Error("First alert (large-result-000) not found in results - possible pagination truncation")
	}
	if !foundLast {
		t.Error("Last alert (large-result-499) not found in results - possible pagination truncation")
	}
}

// ============================================================================
// HIGH PRIORITY: Unicode & International Characters
// ============================================================================

func TestIntegration_SavePoliceAlerts_UnicodeStreets(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Test various international character sets
	testCases := []struct {
		uuid    string
		street  string
		city    string
		country string
	}{
		{"unicode-001", "王府井大街", "北京", "CN"},                              // Chinese
		{"unicode-002", "Friedrichstraße", "Berlin", "DE"},                // German with ß
		{"unicode-003", "улица Арбат", "Москва", "RU"},                     // Russian Cyrillic
		{"unicode-004", "Avenida Paulista", "São Paulo", "BR"},            // Portuguese with tilde
		{"unicode-005", "Rue de la Paix", "Paris", "FR"},                  // French with accents
		{"unicode-006", "شارع الشانزليزيه", "دبي", "AE"},                   // Arabic (RTL)
		{"unicode-007", "Ομόνοια", "Αθήνα", "GR"},                          // Greek
		{"unicode-008", "東京タワー通り", "東京", "JP"},                            // Japanese (mixed)
		{"unicode-009", "Çeşme Caddesi", "İstanbul", "TR"},                // Turkish with special chars
		{"unicode-010", "Västerås Gatan", "Stockholm", "SE"},              // Swedish with å, ä, ö
		{"unicode-011", "Calle ñoño 🚔", "Madrid", "ES"},                   // Spanish with ñ and emoji
		{"unicode-012", "മഹാത്മാഗാന്ധി Road", "കൊച്ചി", "IN"},               // Malayalam
	}

	alerts := make([]models.WazeAlert, len(testCases))
	for i, tc := range testCases {
		alerts[i] = createTestWazeAlert(tc.uuid, "POLICE", map[string]interface{}{
			"Street":  tc.street,
			"City":    tc.city,
			"Country": tc.country,
		})
	}

	// Save all unicode alerts
	err := h.client.SavePoliceAlerts(h.ctx, alerts, time.Now())
	if err != nil {
		t.Fatalf("SavePoliceAlerts with unicode failed: %v", err)
	}

	// Retrieve and verify each alert
	for _, tc := range testCases {
		doc, err := h.client.client.Collection(h.collectionName).Doc(tc.uuid).Get(h.ctx)
		if err != nil {
			t.Errorf("Failed to get alert %s: %v", tc.uuid, err)
			continue
		}

		data := doc.Data()

		// Verify unicode strings preserved exactly
		if data["street"] != tc.street {
			t.Errorf("Alert %s: street encoding mismatch\n  Expected: %s\n  Got: %v", tc.uuid, tc.street, data["street"])
		}
		if data["city"] != tc.city {
			t.Errorf("Alert %s: city encoding mismatch\n  Expected: %s\n  Got: %v", tc.uuid, tc.city, data["city"])
		}
		if data["country"] != tc.country {
			t.Errorf("Alert %s: country encoding mismatch\n  Expected: %s\n  Got: %v", tc.uuid, tc.country, data["country"])
		}
	}

	t.Logf("Successfully stored and retrieved %d alerts with international characters", len(testCases))
}

func TestIntegration_GetPoliceAlertsByDatesWithFilters_UnicodeStreetFilter(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup()

	// Create alerts with unicode streets
	alerts := []models.WazeAlert{
		createTestWazeAlert("unicode-filter-001", "POLICE", map[string]interface{}{
			"Street": "王府井大街",
			"City":   "北京",
		}),
		createTestWazeAlert("unicode-filter-002", "POLICE", map[string]interface{}{
			"Street": "Friedrichstraße",
			"City":   "Berlin",
		}),
		createTestWazeAlert("unicode-filter-003", "POLICE", map[string]interface{}{
			"Street": "улица Арбат",
			"City":   "Москва",
		}),
		createTestWazeAlert("unicode-filter-004", "POLICE", map[string]interface{}{
			"Street": "Main Street", // English for control
			"City":   "Sydney",
		}),
	}

	err := h.client.SavePoliceAlerts(h.ctx, alerts, time.Now())
	if err != nil {
		t.Fatalf("SavePoliceAlerts failed: %v", err)
	}

	// Test filtering by each unicode street
	unicodeStreets := []struct {
		street       string
		expectedUUID string
	}{
		{"王府井大街", "unicode-filter-001"},
		{"Friedrichstraße", "unicode-filter-002"},
		{"улица Арбат", "unicode-filter-003"},
	}

	today := time.Now().Format("2006-01-02")

	for _, tc := range unicodeStreets {
		results, err := h.client.GetPoliceAlertsByDatesWithFilters(
			h.ctx,
			[]string{today},
			[]string{}, // No subtype filter
			[]string{tc.street},
		)

		if err != nil {
			t.Errorf("GetPoliceAlertsByDatesWithFilters failed for street '%s': %v", tc.street, err)
			continue
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 alert for street '%s', got %d", tc.street, len(results))
			continue
		}

		if results[0].UUID != tc.expectedUUID {
			t.Errorf("Expected UUID %s for street '%s', got %s", tc.expectedUUID, tc.street, results[0].UUID)
		}

		if results[0].Street != tc.street {
			t.Errorf("Street field mismatch: expected '%s', got '%s'", tc.street, results[0].Street)
		}
	}

	t.Logf("Unicode street filtering working correctly")
}
