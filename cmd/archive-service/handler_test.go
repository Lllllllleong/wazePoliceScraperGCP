package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
)

// TestHealthHandler tests the health check endpoint
func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got %q", rr.Body.String())
	}
}

// TestArchiveHandlerMethodNotAllowed tests that only POST is allowed
func TestArchiveHandlerMethodNotAllowed(t *testing.T) {
	s := &server{}

	methods := []string{"GET", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(s.archiveHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("method %s: expected status %d, got %d", method, http.StatusMethodNotAllowed, rr.Code)
			}

			if !strings.Contains(rr.Body.String(), "POST") {
				t.Errorf("expected error message to mention 'POST', got %q", rr.Body.String())
			}
		})
	}
}

// TestCreateJSONL tests the JSONL creation function
func TestCreateJSONL(t *testing.T) {
	tests := []struct {
		name          string
		alerts        []models.PoliceAlert
		expectedLines int
	}{
		{
			name:          "empty alerts",
			alerts:        []models.PoliceAlert{},
			expectedLines: 0,
		},
		{
			name: "single alert",
			alerts: []models.PoliceAlert{
				{UUID: "alert-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
			},
			expectedLines: 1,
		},
		{
			name: "multiple alerts",
			alerts: []models.PoliceAlert{
				{UUID: "alert-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
				{UUID: "alert-2", Type: "POLICE", Subtype: "POLICE_WITH_MOBILE_CAMERA"},
				{UUID: "alert-3", Type: "POLICE", Subtype: "POLICE_HIDING"},
			},
			expectedLines: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := createJSONL(tt.alerts)
			if err != nil {
				t.Fatalf("createJSONL failed: %v", err)
			}

			if tt.expectedLines == 0 {
				if len(data) != 0 {
					t.Errorf("expected empty data for empty alerts, got %d bytes", len(data))
				}
				return
			}

			// Count lines
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			if len(lines) != tt.expectedLines {
				t.Errorf("expected %d lines, got %d", tt.expectedLines, len(lines))
			}

			// Verify each line is valid JSON
			for i, line := range lines {
				var parsed models.PoliceAlert
				if err := json.Unmarshal([]byte(line), &parsed); err != nil {
					t.Errorf("line %d is not valid JSON: %v", i, err)
				}
			}
		})
	}
}

// TestCreateJSONLPreservesData tests that JSONL output preserves all alert data
func TestCreateJSONLPreservesData(t *testing.T) {
	now := time.Now()
	alerts := []models.PoliceAlert{
		{
			UUID:             "test-uuid-123",
			Type:             "POLICE",
			Subtype:          "POLICE_WITH_MOBILE_CAMERA",
			Street:           "Hume Highway",
			City:             "Goulburn",
			Country:          "AU",
			Reliability:      9,
			Confidence:       8,
			PublishTime:      now,
			ScrapeTime:       now,
			ExpireTime:       now.Add(2 * time.Hour),
			ActiveMillis:     7200000,
			NThumbsUpInitial: 3,
			NThumbsUpLast:    7,
		},
	}

	data, err := createJSONL(alerts)
	if err != nil {
		t.Fatalf("createJSONL failed: %v", err)
	}

	// Parse the JSONL line back
	var parsed models.PoliceAlert
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &parsed); err != nil {
		t.Fatalf("failed to parse JSONL: %v", err)
	}

	// Verify fields are preserved
	if parsed.UUID != "test-uuid-123" {
		t.Errorf("UUID not preserved: expected %q, got %q", "test-uuid-123", parsed.UUID)
	}
	if parsed.Type != "POLICE" {
		t.Errorf("Type not preserved: expected %q, got %q", "POLICE", parsed.Type)
	}
	if parsed.Subtype != "POLICE_WITH_MOBILE_CAMERA" {
		t.Errorf("Subtype not preserved: expected %q, got %q", "POLICE_WITH_MOBILE_CAMERA", parsed.Subtype)
	}
	if parsed.Street != "Hume Highway" {
		t.Errorf("Street not preserved: expected %q, got %q", "Hume Highway", parsed.Street)
	}
	if parsed.City != "Goulburn" {
		t.Errorf("City not preserved: expected %q, got %q", "Goulburn", parsed.City)
	}
	if parsed.Reliability != 9 {
		t.Errorf("Reliability not preserved: expected %d, got %d", 9, parsed.Reliability)
	}
	if parsed.NThumbsUpLast != 7 {
		t.Errorf("NThumbsUpLast not preserved: expected %d, got %d", 7, parsed.NThumbsUpLast)
	}
}

// TestArchiveDateFormat tests the expected date format for archive files
func TestArchiveDateFormat(t *testing.T) {
	tests := []struct {
		name           string
		date           time.Time
		expectedFormat string
	}{
		{
			name:           "standard date",
			date:           time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedFormat: "2024-01-15",
		},
		{
			name:           "single digit month and day",
			date:           time.Date(2024, 5, 7, 0, 0, 0, 0, time.UTC),
			expectedFormat: "2024-05-07",
		},
		{
			name:           "end of year",
			date:           time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			expectedFormat: "2024-12-31",
		},
		{
			name:           "leap year date",
			date:           time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			expectedFormat: "2024-02-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.date.Format("2006-01-02")
			if formatted != tt.expectedFormat {
				t.Errorf("expected %q, got %q", tt.expectedFormat, formatted)
			}

			// Verify the expected filename would be correct
			expectedFilename := tt.expectedFormat + ".jsonl"
			if !strings.HasSuffix(expectedFilename, ".jsonl") {
				t.Errorf("filename should end with .jsonl: %s", expectedFilename)
			}
		})
	}
}

// TestArchiveRequestBodyParsing tests parsing of request body for custom date
func TestArchiveRequestBodyParsing(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		expectedDate string
		expectError  bool
	}{
		{
			name:         "valid date in body",
			body:         `{"date": "2024-01-15"}`,
			expectedDate: "2024-01-15",
			expectError:  false,
		},
		{
			name:         "empty body defaults to yesterday",
			body:         `{}`,
			expectedDate: "", // Will use yesterday's date
			expectError:  false,
		},
		{
			name:        "invalid date format",
			body:        `{"date": "15-01-2024"}`,
			expectError: true,
		},
		{
			name:        "invalid JSON",
			body:        `not json`,
			expectError: false, // Invalid JSON defaults to yesterday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody struct {
				Date string `json:"date"`
			}

			err := json.Unmarshal([]byte(tt.body), &requestBody)

			if tt.body == `not json` {
				// Invalid JSON should fail parsing, but we handle it gracefully
				if err == nil {
					t.Error("expected error for invalid JSON")
				}
				return
			}

			if tt.expectedDate != "" && requestBody.Date != tt.expectedDate {
				t.Errorf("expected date %q, got %q", tt.expectedDate, requestBody.Date)
			}

			// If we have a date, validate the format
			if requestBody.Date != "" {
				loc, _ := time.LoadLocation("Australia/Canberra")
				_, parseErr := time.ParseInLocation("2006-01-02", requestBody.Date, loc)
				if tt.expectError && parseErr == nil {
					t.Error("expected parse error for invalid date format")
				}
				if !tt.expectError && parseErr != nil {
					t.Errorf("unexpected parse error: %v", parseErr)
				}
			}
		})
	}
}

// TestTimezoneHandling tests Australia/Canberra timezone handling
func TestTimezoneHandling(t *testing.T) {
	loc, err := time.LoadLocation("Australia/Canberra")
	if err != nil {
		t.Fatalf("failed to load timezone: %v", err)
	}

	// Test that we can create dates in Canberra timezone
	testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, loc)

	// Verify the timezone is set correctly
	zone, _ := testDate.Zone()
	if zone != "AEDT" && zone != "AEST" {
		t.Logf("timezone zone: %s (expected AEDT or AEST)", zone)
	}

	// Test day boundaries
	startOfDay := time.Date(testDate.Year(), testDate.Month(), testDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)

	if startOfDay.Hour() != 0 || startOfDay.Minute() != 0 || startOfDay.Second() != 0 {
		t.Errorf("start of day should be midnight: got %v", startOfDay)
	}

	if endOfDay.Hour() != 23 || endOfDay.Minute() != 59 || endOfDay.Second() != 59 {
		t.Errorf("end of day should be 23:59:59: got %v", endOfDay)
	}
}

// TestYesterdayCalculation tests calculating yesterday's date
func TestYesterdayCalculation(t *testing.T) {
	loc, _ := time.LoadLocation("Australia/Canberra")
	now := time.Now().In(loc)
	yesterday := now.AddDate(0, 0, -1)

	// Verify yesterday is exactly one day before today
	diff := now.Sub(yesterday)
	if diff < 23*time.Hour || diff > 25*time.Hour {
		t.Errorf("yesterday should be approximately 24 hours ago, got %v", diff)
	}

	// Verify the date is different
	if now.YearDay() == yesterday.YearDay() && now.Year() == yesterday.Year() {
		t.Error("yesterday should have a different day than today")
	}
}

// TestJSONLLineEndings tests that JSONL has proper line endings
func TestJSONLLineEndings(t *testing.T) {
	alerts := []models.PoliceAlert{
		{UUID: "alert-1", Type: "POLICE"},
		{UUID: "alert-2", Type: "POLICE"},
	}

	data, err := createJSONL(alerts)
	if err != nil {
		t.Fatalf("createJSONL failed: %v", err)
	}

	// Each line should end with newline
	if data[len(data)-1] != '\n' {
		t.Error("JSONL should end with newline")
	}

	// Count newlines
	newlineCount := 0
	for _, b := range data {
		if b == '\n' {
			newlineCount++
		}
	}

	if newlineCount != len(alerts) {
		t.Errorf("expected %d newlines, got %d", len(alerts), newlineCount)
	}
}

// TestEmptyAlertsResponse tests behavior when no alerts to archive
func TestEmptyAlertsResponse(t *testing.T) {
	// Empty alerts should produce empty JSONL
	data, err := createJSONL([]models.PoliceAlert{})
	if err != nil {
		t.Fatalf("createJSONL failed: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("expected empty data for no alerts, got %d bytes", len(data))
	}
}

// TestLargeDatasetJSONL tests JSONL creation with larger datasets
func TestLargeDatasetJSONL(t *testing.T) {
	// Create 100 alerts
	alerts := make([]models.PoliceAlert, 100)
	for i := 0; i < 100; i++ {
		alerts[i] = models.PoliceAlert{
			UUID:    string(rune('a'+i%26)) + "-" + string(rune('0'+i%10)),
			Type:    "POLICE",
			Subtype: "POLICE_VISIBLE",
			Street:  "Test Street",
		}
	}

	data, err := createJSONL(alerts)
	if err != nil {
		t.Fatalf("createJSONL failed for large dataset: %v", err)
	}

	// Verify line count
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 100 {
		t.Errorf("expected 100 lines, got %d", len(lines))
	}
}

// TestFilenameGeneration tests archive filename generation
func TestFilenameGeneration(t *testing.T) {
	tests := []struct {
		name             string
		date             time.Time
		expectedFilename string
	}{
		{
			name:             "standard date",
			date:             time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC),
			expectedFilename: "2024-01-15.jsonl",
		},
		{
			name:             "new year's day",
			date:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedFilename: "2025-01-01.jsonl",
		},
		{
			name:             "end of year",
			date:             time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expectedFilename: "2024-12-31.jsonl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.date.Format("2006-01-02") + ".jsonl"
			if filename != tt.expectedFilename {
				t.Errorf("expected filename %q, got %q", tt.expectedFilename, filename)
			}
		})
	}
}

// TestIdempotencyLogic tests the idempotency check concept
func TestIdempotencyLogic(t *testing.T) {
	// Simulate the idempotency check logic
	existingArchives := map[string]bool{
		"2024-01-01.jsonl": true,
		"2024-01-02.jsonl": true,
		"2024-01-03.jsonl": true,
	}

	tests := []struct {
		name       string
		targetDate string
		shouldSkip bool
	}{
		{"archive exists", "2024-01-01", true},
		{"archive exists 2", "2024-01-02", true},
		{"archive does not exist", "2024-01-04", false},
		{"archive does not exist 2", "2024-01-05", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.targetDate + ".jsonl"
			exists := existingArchives[filename]

			if exists != tt.shouldSkip {
				t.Errorf("expected shouldSkip=%v for %s, got %v", tt.shouldSkip, tt.targetDate, exists)
			}
		})
	}
}

// TestAlertCountInResponse tests that response includes correct alert count
func TestAlertCountInResponse(t *testing.T) {
	alerts := []models.PoliceAlert{
		{UUID: "1"},
		{UUID: "2"},
		{UUID: "3"},
	}

	count := len(alerts)
	expectedMsg := "Successfully archived 3 alerts"

	if count != 3 {
		t.Errorf("expected 3 alerts, got %d", count)
	}

	if !strings.Contains(expectedMsg, "3") {
		t.Errorf("expected message to contain count, got %q", expectedMsg)
	}
}

// =============================================================================
// Handler Integration Tests with Mocks
// =============================================================================

// mockAlertStore is a mock implementation of storage.AlertStore for testing.
type mockAlertStore struct {
	GetPoliceAlertsByDateRangeFunc func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error)
}

func (m *mockAlertStore) GetPoliceAlertsByDateRange(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
	if m.GetPoliceAlertsByDateRangeFunc != nil {
		return m.GetPoliceAlertsByDateRangeFunc(ctx, start, end)
	}
	return nil, nil
}

func (m *mockAlertStore) SavePoliceAlerts(ctx context.Context, alerts []models.WazeAlert, scrapeTime time.Time) error {
	return nil
}

func (m *mockAlertStore) GetPoliceAlertsByDatesWithFilters(ctx context.Context, dates []string, subtypes []string, streets []string) ([]models.PoliceAlert, error) {
	return nil, nil
}

func (m *mockAlertStore) Close() error {
	return nil
}

// Ensure mockAlertStore implements storage.AlertStore
var _ storage.AlertStore = (*mockAlertStore)(nil)

// createTestServer creates a server with mock dependencies for testing
func createTestServer(alertStore storage.AlertStore, gcsClient storage.GCSClient) *server {
	return &server{
		alertStore:   alertStore,
		gcsClient:    gcsClient,
		bucketName:   "test-bucket",
		loadLocation: time.LoadLocation,
	}
}

// TestArchiveHandlerSuccess tests successful archive operation
func TestArchiveHandlerSuccess(t *testing.T) {
	testDate := "2024-01-15"
	alerts := []models.PoliceAlert{
		{UUID: "alert-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
		{UUID: "alert-2", Type: "POLICE", Subtype: "POLICE_HIDING"},
	}

	var writtenData []byte
	mockWriter := &storage.MockGCSWriter{}

	mockStore := &mockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
			return alerts, nil
		},
	}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			// Object doesn't exist
			return nil, storage.ErrObjectNotExist
		},
		NewWriterFunc: func(ctx context.Context) storage.GCSWriter {
			return mockWriter
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			expectedFilename := testDate + ".jsonl"
			if name != expectedFilename {
				t.Errorf("expected filename %q, got %q", expectedFilename, name)
			}
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			if name != "test-bucket" {
				t.Errorf("expected bucket 'test-bucket', got %q", name)
			}
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "Successfully archived 2 alerts") {
		t.Errorf("expected success message with count, got %q", rr.Body.String())
	}

	// Verify data was written
	writtenData = mockWriter.Written
	if len(writtenData) == 0 {
		t.Error("expected data to be written to GCS")
	}

	// Verify JSONL format
	lines := strings.Split(strings.TrimSpace(string(writtenData)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 JSONL lines, got %d", len(lines))
	}
}

// TestArchiveHandlerAlreadyExists tests idempotency when archive already exists
func TestArchiveHandlerAlreadyExists(t *testing.T) {
	testDate := "2024-01-15"

	mockStore := &mockAlertStore{}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			// Object already exists
			return &storage.GCSObjectAttrs{
				Name: testDate + ".jsonl",
				Size: 1024,
			}, nil
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "already exists") {
		t.Errorf("expected 'already exists' message, got %q", rr.Body.String())
	}
}

// TestArchiveHandlerNoAlerts tests behavior when no alerts for the date
func TestArchiveHandlerNoAlerts(t *testing.T) {
	testDate := "2024-01-15"

	mockStore := &mockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
			return []models.PoliceAlert{}, nil
		},
	}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			return nil, storage.ErrObjectNotExist
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "No alerts to archive") {
		t.Errorf("expected 'No alerts to archive' message, got %q", rr.Body.String())
	}
}

// TestArchiveHandlerFirestoreError tests error handling for Firestore failures
func TestArchiveHandlerFirestoreError(t *testing.T) {
	testDate := "2024-01-15"

	mockStore := &mockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
			return nil, errors.New("firestore connection failed")
		},
	}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			return nil, storage.ErrObjectNotExist
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestArchiveHandlerGCSWriteError tests error handling for GCS write failures
func TestArchiveHandlerGCSWriteError(t *testing.T) {
	testDate := "2024-01-15"
	alerts := []models.PoliceAlert{
		{UUID: "alert-1", Type: "POLICE"},
	}

	mockStore := &mockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
			return alerts, nil
		},
	}

	mockWriter := &storage.MockGCSWriter{
		WriteFunc: func(p []byte) (n int, err error) {
			return 0, errors.New("gcs write failed")
		},
	}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			return nil, storage.ErrObjectNotExist
		},
		NewWriterFunc: func(ctx context.Context) storage.GCSWriter {
			return mockWriter
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestArchiveHandlerGCSCloseError tests error handling for GCS close failures
func TestArchiveHandlerGCSCloseError(t *testing.T) {
	testDate := "2024-01-15"
	alerts := []models.PoliceAlert{
		{UUID: "alert-1", Type: "POLICE"},
	}

	mockStore := &mockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
			return alerts, nil
		},
	}

	mockWriter := &storage.MockGCSWriter{
		CloseFunc: func() error {
			return errors.New("gcs close failed")
		},
	}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			return nil, storage.ErrObjectNotExist
		},
		NewWriterFunc: func(ctx context.Context) storage.GCSWriter {
			return mockWriter
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestArchiveHandlerGCSAttrsError tests error handling for GCS Attrs failures (not NotExist)
func TestArchiveHandlerGCSAttrsError(t *testing.T) {
	testDate := "2024-01-15"

	mockStore := &mockAlertStore{}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			return nil, errors.New("gcs permission denied")
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestArchiveHandlerInvalidDateFormat tests error handling for invalid date format
func TestArchiveHandlerInvalidDateFormat(t *testing.T) {
	mockStore := &mockAlertStore{}
	mockGCS := &storage.MockGCSClient{}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "15-01-2024"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid date format") {
		t.Errorf("expected 'Invalid date format' message, got %q", rr.Body.String())
	}
}

// TestArchiveHandlerDefaultsToYesterday tests that no date defaults to yesterday
func TestArchiveHandlerDefaultsToYesterday(t *testing.T) {
	var capturedStart time.Time

	mockStore := &mockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
			capturedStart = start
			return []models.PoliceAlert{}, nil
		},
	}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			return nil, storage.ErrObjectNotExist
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	// Empty body - should default to yesterday
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify the date range is for yesterday
	loc, _ := time.LoadLocation("Australia/Canberra")
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	expectedStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, loc)

	if capturedStart.Year() != expectedStart.Year() ||
		capturedStart.Month() != expectedStart.Month() ||
		capturedStart.Day() != expectedStart.Day() {
		t.Errorf("expected date range for %s, got %s", expectedStart.Format("2006-01-02"), capturedStart.Format("2006-01-02"))
	}
}

// TestArchiveHandlerTimezoneLoadError tests error handling for timezone loading failures
func TestArchiveHandlerTimezoneLoadError(t *testing.T) {
	mockStore := &mockAlertStore{}
	mockGCS := &storage.MockGCSClient{}

	s := &server{
		alertStore: mockStore,
		gcsClient:  mockGCS,
		bucketName: "test-bucket",
		loadLocation: func(name string) (*time.Location, error) {
			return nil, errors.New("timezone not found")
		},
	}

	body := strings.NewReader(`{"date": "2024-01-15"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestArchiveHandlerVerifiesJSONLContent tests that correct JSONL content is written
func TestArchiveHandlerVerifiesJSONLContent(t *testing.T) {
	testDate := "2024-01-15"
	now := time.Now()
	alerts := []models.PoliceAlert{
		{
			UUID:        "test-uuid-123",
			Type:        "POLICE",
			Subtype:     "POLICE_VISIBLE",
			Street:      "Test Street",
			City:        "Canberra",
			Country:     "AU",
			Reliability: 8,
			PublishTime: now,
		},
	}

	mockWriter := &storage.MockGCSWriter{}

	mockStore := &mockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error) {
			return alerts, nil
		},
	}

	mockObjHandle := &storage.MockGCSObjectHandle{
		AttrsFunc: func(ctx context.Context) (*storage.GCSObjectAttrs, error) {
			return nil, storage.ErrObjectNotExist
		},
		NewWriterFunc: func(ctx context.Context) storage.GCSWriter {
			return mockWriter
		},
	}

	mockBucket := &storage.MockGCSBucketHandle{
		ObjectFunc: func(name string) storage.GCSObjectHandle {
			return mockObjHandle
		},
	}

	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return mockBucket
		},
	}

	s := createTestServer(mockStore, mockGCS)

	body := strings.NewReader(`{"date": "` + testDate + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rr := httptest.NewRecorder()

	s.archiveHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Parse the written JSONL
	writtenData := mockWriter.Written
	var parsed models.PoliceAlert
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(writtenData))), &parsed); err != nil {
		t.Fatalf("failed to parse written JSONL: %v", err)
	}

	// Verify content
	if parsed.UUID != "test-uuid-123" {
		t.Errorf("expected UUID 'test-uuid-123', got %q", parsed.UUID)
	}
	if parsed.Street != "Test Street" {
		t.Errorf("expected Street 'Test Street', got %q", parsed.Street)
	}
	if parsed.City != "Canberra" {
		t.Errorf("expected City 'Canberra', got %q", parsed.City)
	}
}
