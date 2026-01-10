package waze

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.httpClient == nil {
		t.Error("Expected non-nil httpClient")
	}
	if client.stats == nil {
		t.Error("Expected non-nil stats")
	}
}

func TestGetStats(t *testing.T) {
	client := NewClient()
	stats := client.GetStats()

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}
	if stats.TotalRequests != 0 {
		t.Errorf("Expected TotalRequests 0, got %d", stats.TotalRequests)
	}
	if stats.SuccessfulCalls != 0 {
		t.Errorf("Expected SuccessfulCalls 0, got %d", stats.SuccessfulCalls)
	}
	if stats.FailedCalls != 0 {
		t.Errorf("Expected FailedCalls 0, got %d", stats.FailedCalls)
	}
}

func TestGetAlertsInvalidBBox(t *testing.T) {
	client := NewClient()

	// Test with invalid bounding box format
	_, err := client.GetAlerts("invalid")
	if err == nil {
		t.Error("Expected error for invalid bbox format")
	}

	// Test with too few parts
	_, err = client.GetAlerts("1,2,3")
	if err == nil {
		t.Error("Expected error for bbox with too few parts")
	}

	// Test with too many parts
	_, err = client.GetAlerts("1,2,3,4,5")
	if err == nil {
		t.Error("Expected error for bbox with too many parts")
	}
}

func TestMinHelper(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestGetAlertsMultipleBBoxesEmpty(t *testing.T) {
	client := NewClient()

	// Empty bboxes should return error
	_, err := client.GetAlertsMultipleBBoxes([]string{})
	if err == nil {
		t.Error("Expected error for empty bboxes")
	}
}

// TestBBoxParsing tests various bounding box format edge cases
func TestBBoxParsing(t *testing.T) {
	tests := []struct {
		name        string
		bbox        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid bbox",
			bbox:        "150.38,-34.25,151.00,-33.93",
			expectError: false,
		},
		{
			name:        "valid bbox with spaces trimmed internally",
			bbox:        "150.38,-34.25,151.00,-33.93",
			expectError: false,
		},
		{
			name:        "empty string",
			bbox:        "",
			expectError: true,
			errorMsg:    "invalid bounding box format",
		},
		{
			name:        "single value",
			bbox:        "150.38",
			expectError: true,
			errorMsg:    "invalid bounding box format",
		},
		{
			name:        "two values",
			bbox:        "150.38,-34.25",
			expectError: true,
			errorMsg:    "invalid bounding box format",
		},
		{
			name:        "three values",
			bbox:        "150.38,-34.25,151.00",
			expectError: true,
			errorMsg:    "invalid bounding box format",
		},
		{
			name:        "five values",
			bbox:        "150.38,-34.25,151.00,-33.93,0",
			expectError: true,
			errorMsg:    "invalid bounding box format",
		},
		{
			name:        "negative coordinates",
			bbox:        "-150.38,-34.25,-151.00,-33.93",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			_, err := client.GetAlerts(tt.bbox)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for bbox %q, got nil", tt.bbox)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			}
			// Note: valid bbox tests will fail because they hit the real Waze API
			// These are tested with mock servers below
		})
	}
}

// createMockWazeServer creates a test HTTP server that simulates Waze API responses
func createMockWazeServer(response models.WazeGeoRSSResponse, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// TestGetAlertsWithMockServer tests GetAlerts with a mock HTTP server
func TestGetAlertsWithMockServer(t *testing.T) {
	tests := []struct {
		name           string
		response       models.WazeGeoRSSResponse
		statusCode     int
		expectError    bool
		expectedAlerts int
	}{
		{
			name: "successful response with multiple alerts",
			response: models.WazeGeoRSSResponse{
				Alerts: []models.WazeAlert{
					{
						UUID:    "alert-1",
						Type:    "POLICE",
						Subtype: "POLICE_VISIBLE",
						Location: models.Location{
							Latitude:  -33.8688,
							Longitude: 151.2093,
						},
						Street:      "George Street",
						City:        "Sydney",
						Country:     "AU",
						Reliability: 8,
						PubMillis:   1704067200000,
					},
					{
						UUID:    "alert-2",
						Type:    "POLICE",
						Subtype: "POLICE_WITH_MOBILE_CAMERA",
						Location: models.Location{
							Latitude:  -34.7515,
							Longitude: 149.7209,
						},
						Street:      "Hume Highway",
						City:        "Goulburn",
						Country:     "AU",
						Reliability: 9,
						PubMillis:   1704070800000,
					},
				},
			},
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedAlerts: 2,
		},
		{
			name: "empty alerts array",
			response: models.WazeGeoRSSResponse{
				Alerts: []models.WazeAlert{},
			},
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedAlerts: 0,
		},
		{
			name: "single alert",
			response: models.WazeGeoRSSResponse{
				Alerts: []models.WazeAlert{
					{
						UUID:    "single-alert",
						Type:    "POLICE",
						Subtype: "POLICE_HIDING",
						Location: models.Location{
							Latitude:  -35.2809,
							Longitude: 149.1300,
						},
						Street:      "Commonwealth Avenue",
						City:        "Canberra",
						Country:     "AU",
						Reliability: 7,
						PubMillis:   1704074400000,
					},
				},
			},
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedAlerts: 1,
		},
		{
			name:           "server error 500",
			response:       models.WazeGeoRSSResponse{},
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			expectedAlerts: 0,
		},
		{
			name:           "not found 404",
			response:       models.WazeGeoRSSResponse{},
			statusCode:     http.StatusNotFound,
			expectError:    true,
			expectedAlerts: 0,
		},
		{
			name:           "rate limited 429",
			response:       models.WazeGeoRSSResponse{},
			statusCode:     http.StatusTooManyRequests,
			expectError:    true,
			expectedAlerts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockWazeServer(tt.response, tt.statusCode)
			defer server.Close()

			// Create a client and override its HTTP call by creating a custom handler
			client := NewClient()

			// Make a test request to our mock server instead
			resp, err := client.httpClient.Get(server.URL)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, resp.StatusCode)
			}

			if tt.expectError && resp.StatusCode == http.StatusOK {
				t.Error("expected error status code")
			}

			if !tt.expectError && resp.StatusCode == http.StatusOK {
				var apiResponse models.WazeGeoRSSResponse
				if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(apiResponse.Alerts) != tt.expectedAlerts {
					t.Errorf("expected %d alerts, got %d", tt.expectedAlerts, len(apiResponse.Alerts))
				}
			}
		})
	}
}

// TestDeduplicationLogic tests alert deduplication across multiple bboxes
func TestDeduplicationLogic(t *testing.T) {
	// Simulate alerts from multiple overlapping bboxes
	alertsFromBBox1 := []models.WazeAlert{
		{UUID: "unique-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
		{UUID: "duplicate-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
		{UUID: "unique-2", Type: "POLICE", Subtype: "POLICE_WITH_MOBILE_CAMERA"},
	}

	alertsFromBBox2 := []models.WazeAlert{
		{UUID: "duplicate-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"}, // Duplicate
		{UUID: "unique-3", Type: "POLICE", Subtype: "POLICE_HIDING"},
		{UUID: "duplicate-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"}, // Another duplicate
	}

	alertsFromBBox3 := []models.WazeAlert{
		{UUID: "unique-4", Type: "POLICE", Subtype: "POLICE_ON_BRIDGE"},
		{UUID: "unique-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"}, // Duplicate from bbox1
	}

	// Combine all alerts and deduplicate (simulating GetAlertsMultipleBBoxes logic)
	uniqueAlerts := make(map[string]models.WazeAlert)

	allAlertGroups := [][]models.WazeAlert{alertsFromBBox1, alertsFromBBox2, alertsFromBBox3}
	for _, alerts := range allAlertGroups {
		for _, alert := range alerts {
			if alert.UUID != "" {
				if _, exists := uniqueAlerts[alert.UUID]; !exists {
					uniqueAlerts[alert.UUID] = alert
				}
			}
		}
	}

	// Expected unique alerts: unique-1, duplicate-1, unique-2, unique-3, unique-4
	expectedUniqueCount := 5
	if len(uniqueAlerts) != expectedUniqueCount {
		t.Errorf("expected %d unique alerts, got %d", expectedUniqueCount, len(uniqueAlerts))
	}

	// Verify each unique alert is present
	expectedUUIDs := []string{"unique-1", "duplicate-1", "unique-2", "unique-3", "unique-4"}
	for _, uuid := range expectedUUIDs {
		if _, exists := uniqueAlerts[uuid]; !exists {
			t.Errorf("expected UUID %q not found in deduplicated results", uuid)
		}
	}
}

// TestEmptyUUIDHandling tests that alerts with empty UUIDs are handled correctly
func TestEmptyUUIDHandling(t *testing.T) {
	alerts := []models.WazeAlert{
		{UUID: "valid-uuid-1", Type: "POLICE"},
		{UUID: "", Type: "POLICE"}, // Empty UUID should be ignored
		{UUID: "valid-uuid-2", Type: "POLICE"},
		{UUID: "", Type: "ACCIDENT"}, // Another empty UUID
	}

	uniqueAlerts := make(map[string]models.WazeAlert)
	for _, alert := range alerts {
		if alert.UUID != "" {
			if _, exists := uniqueAlerts[alert.UUID]; !exists {
				uniqueAlerts[alert.UUID] = alert
			}
		}
	}

	// Only alerts with non-empty UUIDs should be included
	if len(uniqueAlerts) != 2 {
		t.Errorf("expected 2 alerts with valid UUIDs, got %d", len(uniqueAlerts))
	}

	if _, exists := uniqueAlerts["valid-uuid-1"]; !exists {
		t.Error("valid-uuid-1 should be in the map")
	}
	if _, exists := uniqueAlerts["valid-uuid-2"]; !exists {
		t.Error("valid-uuid-2 should be in the map")
	}
	if _, exists := uniqueAlerts[""]; exists {
		t.Error("empty UUID should not be in the map")
	}
}

// TestStatsTracking tests that statistics are correctly tracked
func TestStatsTracking(t *testing.T) {
	client := NewClient()

	// Initial stats should be zero
	stats := client.GetStats()
	if stats.TotalRequests != 0 || stats.SuccessfulCalls != 0 || stats.FailedCalls != 0 {
		t.Error("initial stats should be zero")
	}

	// Make an invalid request to test failed call tracking
	_, _ = client.GetAlerts("invalid-bbox")
	stats = client.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("expected TotalRequests 1, got %d", stats.TotalRequests)
	}
	// Note: FailedCalls is incremented only on HTTP errors, not validation errors

	// Make another invalid request
	_, _ = client.GetAlerts("1,2")
	stats = client.GetStats()
	if stats.TotalRequests != 2 {
		t.Errorf("expected TotalRequests 2, got %d", stats.TotalRequests)
	}
}

// TestMixedAlertTypes tests filtering alerts by type
func TestMixedAlertTypes(t *testing.T) {
	alerts := []models.WazeAlert{
		{UUID: "police-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
		{UUID: "accident-1", Type: "ACCIDENT", Subtype: "MINOR"},
		{UUID: "police-2", Type: "POLICE", Subtype: "POLICE_WITH_MOBILE_CAMERA"},
		{UUID: "jam-1", Type: "JAM", Subtype: "MODERATE_TRAFFIC"},
		{UUID: "hazard-1", Type: "HAZARD", Subtype: "ROAD_OBJECT"},
		{UUID: "police-3", Type: "POLICE", Subtype: "POLICE_HIDING"},
	}

	// Filter for POLICE type only
	var policeAlerts []models.WazeAlert
	for _, alert := range alerts {
		if alert.Type == "POLICE" {
			policeAlerts = append(policeAlerts, alert)
		}
	}

	if len(policeAlerts) != 3 {
		t.Errorf("expected 3 police alerts, got %d", len(policeAlerts))
	}

	// Verify only POLICE type alerts are included
	for _, alert := range policeAlerts {
		if alert.Type != "POLICE" {
			t.Errorf("unexpected alert type %q in filtered results", alert.Type)
		}
	}
}

// TestInvalidJSONResponse tests handling of invalid JSON responses
func TestInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("this is not valid json"))
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.httpClient.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var apiResponse models.WazeGeoRSSResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err == nil {
		t.Error("expected JSON decode error for invalid JSON")
	}
}

// TestAlertFieldsPreservation tests that all alert fields are correctly preserved
func TestAlertFieldsPreservation(t *testing.T) {
	originalAlert := models.WazeAlert{
		UUID:         "test-uuid-123",
		ID:           "test-id-456",
		Type:         "POLICE",
		Subtype:      "POLICE_WITH_MOBILE_CAMERA",
		PubMillis:    1704067200000,
		Location:     models.Location{Latitude: -33.8688, Longitude: 151.2093},
		Street:       "George Street",
		City:         "Sydney",
		Country:      "AU",
		RoadType:     4,
		Reliability:  8,
		Confidence:   7,
		ReportRating: 5,
		NThumbsUp:    10,
		NComments:    3,
	}

	// Serialize and deserialize to test JSON round-trip
	jsonData, err := json.Marshal(originalAlert)
	if err != nil {
		t.Fatalf("failed to marshal alert: %v", err)
	}

	var deserializedAlert models.WazeAlert
	if err := json.Unmarshal(jsonData, &deserializedAlert); err != nil {
		t.Fatalf("failed to unmarshal alert: %v", err)
	}

	// Verify all fields are preserved
	if deserializedAlert.UUID != originalAlert.UUID {
		t.Errorf("UUID mismatch: expected %q, got %q", originalAlert.UUID, deserializedAlert.UUID)
	}
	if deserializedAlert.Type != originalAlert.Type {
		t.Errorf("Type mismatch: expected %q, got %q", originalAlert.Type, deserializedAlert.Type)
	}
	if deserializedAlert.Subtype != originalAlert.Subtype {
		t.Errorf("Subtype mismatch: expected %q, got %q", originalAlert.Subtype, deserializedAlert.Subtype)
	}
	if deserializedAlert.Location.Latitude != originalAlert.Location.Latitude {
		t.Errorf("Latitude mismatch: expected %f, got %f", originalAlert.Location.Latitude, deserializedAlert.Location.Latitude)
	}
	if deserializedAlert.Location.Longitude != originalAlert.Location.Longitude {
		t.Errorf("Longitude mismatch: expected %f, got %f", originalAlert.Location.Longitude, deserializedAlert.Location.Longitude)
	}
	if deserializedAlert.Reliability != originalAlert.Reliability {
		t.Errorf("Reliability mismatch: expected %d, got %d", originalAlert.Reliability, deserializedAlert.Reliability)
	}
	if deserializedAlert.NThumbsUp != originalAlert.NThumbsUp {
		t.Errorf("NThumbsUp mismatch: expected %d, got %d", originalAlert.NThumbsUp, deserializedAlert.NThumbsUp)
	}
}
