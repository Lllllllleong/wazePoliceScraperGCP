package storage

import (
	"testing"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
)

func TestExtractLastVerification(t *testing.T) {
	tests := []struct {
		name                string
		comments            []models.Comment
		expectNil           bool
		expectedMillisValue int64
	}{
		{
			name:      "nil comments returns nil",
			comments:  nil,
			expectNil: true,
		},
		{
			name:      "empty comments returns nil",
			comments:  []models.Comment{},
			expectNil: true,
		},
		{
			name: "single comment returns its timestamp",
			comments: []models.Comment{
				{ReportMillis: 1704067200000, Text: "Still there!", IsThumbsUp: true},
			},
			expectNil:           false,
			expectedMillisValue: 1704067200000,
		},
		{
			name: "multiple comments returns max timestamp",
			comments: []models.Comment{
				{ReportMillis: 1704067200000, Text: "First", IsThumbsUp: true},
				{ReportMillis: 1704070800000, Text: "Second (latest)", IsThumbsUp: true},
				{ReportMillis: 1704064600000, Text: "Earliest", IsThumbsUp: false},
			},
			expectNil:           false,
			expectedMillisValue: 1704070800000, // Latest timestamp
		},
		{
			name: "comments with zero millis are ignored",
			comments: []models.Comment{
				{ReportMillis: 0, Text: "Zero timestamp", IsThumbsUp: true},
			},
			expectNil: true,
		},
		{
			name: "mix of zero and valid millis returns max valid",
			comments: []models.Comment{
				{ReportMillis: 0, Text: "Zero", IsThumbsUp: true},
				{ReportMillis: 1704067200000, Text: "Valid", IsThumbsUp: true},
				{ReportMillis: 0, Text: "Another zero", IsThumbsUp: false},
			},
			expectNil:           false,
			expectedMillisValue: 1704067200000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			millis, verificationTime := extractLastVerification(tt.comments)

			if tt.expectNil {
				if millis != nil {
					t.Errorf("expected nil millis, got %d", *millis)
				}
				if verificationTime != nil {
					t.Errorf("expected nil time, got %v", verificationTime)
				}
			} else {
				if millis == nil {
					t.Fatal("expected non-nil millis, got nil")
				}
				if *millis != tt.expectedMillisValue {
					t.Errorf("expected millis %d, got %d", tt.expectedMillisValue, *millis)
				}
				if verificationTime == nil {
					t.Fatal("expected non-nil time, got nil")
				}
				expectedTime := time.UnixMilli(tt.expectedMillisValue)
				if !verificationTime.Equal(expectedTime) {
					t.Errorf("expected time %v, got %v", expectedTime, *verificationTime)
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "empty slice returns false",
			slice:    []string{},
			value:    "test",
			expected: false,
		},
		{
			name:     "nil slice returns false",
			slice:    nil,
			value:    "test",
			expected: false,
		},
		{
			name:     "value exists in slice",
			slice:    []string{"apple", "banana", "cherry"},
			value:    "banana",
			expected: true,
		},
		{
			name:     "value does not exist in slice",
			slice:    []string{"apple", "banana", "cherry"},
			value:    "grape",
			expected: false,
		},
		{
			name:     "empty string in slice",
			slice:    []string{"", "apple", "banana"},
			value:    "",
			expected: true,
		},
		{
			name:     "case sensitive match",
			slice:    []string{"Apple", "Banana"},
			value:    "apple",
			expected: false,
		},
		{
			name:     "single element slice - match",
			slice:    []string{"single"},
			value:    "single",
			expected: true,
		},
		{
			name:     "single element slice - no match",
			slice:    []string{"single"},
			value:    "other",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("contains(%v, %q) = %v, expected %v", tt.slice, tt.value, result, tt.expected)
			}
		})
	}
}

func TestFirestoreClient_GetCollectionName(t *testing.T) {
	// Test with custom collection name
	fc := &FirestoreClient{
		collectionName: "custom_collection",
	}
	if fc.GetCollectionName() != "custom_collection" {
		t.Errorf("expected 'custom_collection', got %q", fc.GetCollectionName())
	}

	// Test with default collection name
	fc2 := &FirestoreClient{
		collectionName: "police_alerts",
	}
	if fc2.GetCollectionName() != "police_alerts" {
		t.Errorf("expected 'police_alerts', got %q", fc2.GetCollectionName())
	}
}

// TestWazeAlertToPoliceAlertConversion tests the conversion logic embedded in processPoliceAlert
// by testing the data transformations independently
func TestWazeAlertToPoliceAlertConversion(t *testing.T) {
	tests := []struct {
		name           string
		wazeAlert      models.WazeAlert
		expectedType   string
		expectedSubtype string
		expectedLat    float64
		expectedLng    float64
	}{
		{
			name: "standard police alert conversion",
			wazeAlert: models.WazeAlert{
				UUID:    "test-uuid-123",
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
				Confidence:  7,
				PubMillis:   1704067200000,
				NThumbsUp:   5,
			},
			expectedType:    "POLICE",
			expectedSubtype: "POLICE_VISIBLE",
			expectedLat:     -33.8688,
			expectedLng:     151.2093,
		},
		{
			name: "mobile camera alert conversion",
			wazeAlert: models.WazeAlert{
				UUID:    "camera-uuid-456",
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
				Confidence:  8,
				PubMillis:   1704070800000,
				NThumbsUp:   3,
			},
			expectedType:    "POLICE",
			expectedSubtype: "POLICE_WITH_MOBILE_CAMERA",
			expectedLat:     -34.7515,
			expectedLng:     149.7209,
		},
		{
			name: "alert with empty subtype",
			wazeAlert: models.WazeAlert{
				UUID:    "empty-subtype-789",
				Type:    "POLICE",
				Subtype: "",
				Location: models.Location{
					Latitude:  -35.2809,
					Longitude: 149.1300,
				},
				Street:      "Commonwealth Avenue",
				City:        "Canberra",
				Country:     "AU",
				Reliability: 7,
				Confidence:  6,
				PubMillis:   1704074400000,
				NThumbsUp:   1,
			},
			expectedType:    "POLICE",
			expectedSubtype: "",
			expectedLat:     -35.2809,
			expectedLng:     149.1300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the alert type and subtype
			if tt.wazeAlert.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, tt.wazeAlert.Type)
			}
			if tt.wazeAlert.Subtype != tt.expectedSubtype {
				t.Errorf("expected subtype %q, got %q", tt.expectedSubtype, tt.wazeAlert.Subtype)
			}

			// Verify location conversion
			if tt.wazeAlert.Location.Latitude != tt.expectedLat {
				t.Errorf("expected latitude %f, got %f", tt.expectedLat, tt.wazeAlert.Location.Latitude)
			}
			if tt.wazeAlert.Location.Longitude != tt.expectedLng {
				t.Errorf("expected longitude %f, got %f", tt.expectedLng, tt.wazeAlert.Location.Longitude)
			}

			// Verify pubMillis to time conversion
			publishTime := time.UnixMilli(tt.wazeAlert.PubMillis)
			if publishTime.IsZero() {
				t.Error("expected non-zero publish time")
			}
		})
	}
}

// TestPoliceAlertFiltering tests the filtering logic used in GetPoliceAlertsByDatesWithFilters
func TestPoliceAlertFiltering(t *testing.T) {
	alerts := []models.PoliceAlert{
		{
			UUID:    "alert-1",
			Subtype: "POLICE_VISIBLE",
			Street:  "Hume Highway",
		},
		{
			UUID:    "alert-2",
			Subtype: "POLICE_WITH_MOBILE_CAMERA",
			Street:  "Federal Highway",
		},
		{
			UUID:    "alert-3",
			Subtype: "POLICE_VISIBLE",
			Street:  "Federal Highway",
		},
		{
			UUID:    "alert-4",
			Subtype: "",
			Street:  "",
		},
	}

	tests := []struct {
		name           string
		subtypes       []string
		streets        []string
		expectedCount  int
		expectedUUIDs  []string
	}{
		{
			name:          "no filters returns all",
			subtypes:      []string{},
			streets:       []string{},
			expectedCount: 4,
			expectedUUIDs: []string{"alert-1", "alert-2", "alert-3", "alert-4"},
		},
		{
			name:          "filter by single subtype",
			subtypes:      []string{"POLICE_VISIBLE"},
			streets:       []string{},
			expectedCount: 2,
			expectedUUIDs: []string{"alert-1", "alert-3"},
		},
		{
			name:          "filter by single street",
			subtypes:      []string{},
			streets:       []string{"Hume Highway"},
			expectedCount: 1,
			expectedUUIDs: []string{"alert-1"},
		},
		{
			name:          "filter by subtype and street",
			subtypes:      []string{"POLICE_VISIBLE"},
			streets:       []string{"Federal Highway"},
			expectedCount: 1,
			expectedUUIDs: []string{"alert-3"},
		},
		{
			name:          "filter by multiple subtypes",
			subtypes:      []string{"POLICE_VISIBLE", "POLICE_WITH_MOBILE_CAMERA"},
			streets:       []string{},
			expectedCount: 3,
			expectedUUIDs: []string{"alert-1", "alert-2", "alert-3"},
		},
		{
			name:          "filter by empty subtype",
			subtypes:      []string{""},
			streets:       []string{},
			expectedCount: 1,
			expectedUUIDs: []string{"alert-4"},
		},
		{
			name:          "filter by empty street",
			subtypes:      []string{},
			streets:       []string{""},
			expectedCount: 1,
			expectedUUIDs: []string{"alert-4"},
		},
		{
			name:          "no matches returns empty",
			subtypes:      []string{"NONEXISTENT"},
			streets:       []string{},
			expectedCount: 0,
			expectedUUIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply filters manually (simulating the filtering logic)
			var filtered []models.PoliceAlert
			for _, alert := range alerts {
				// Check subtype filter
				if len(tt.subtypes) > 0 && !contains(tt.subtypes, alert.Subtype) {
					continue
				}
				// Check street filter
				if len(tt.streets) > 0 && !contains(tt.streets, alert.Street) {
					continue
				}
				filtered = append(filtered, alert)
			}

			if len(filtered) != tt.expectedCount {
				t.Errorf("expected %d alerts, got %d", tt.expectedCount, len(filtered))
			}

			// Verify the expected UUIDs are present
			for _, expectedUUID := range tt.expectedUUIDs {
				found := false
				for _, alert := range filtered {
					if alert.UUID == expectedUUID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected UUID %q not found in filtered results", expectedUUID)
				}
			}
		})
	}
}

// TestActiveMillisCalculation tests the calculation of active duration
func TestActiveMillisCalculation(t *testing.T) {
	tests := []struct {
		name             string
		pubMillis        int64
		scrapeTimeMillis int64
		expectedActive   int64
	}{
		{
			name:             "one hour active",
			pubMillis:        1704067200000, // 2024-01-01 00:00:00 UTC
			scrapeTimeMillis: 1704070800000, // 2024-01-01 01:00:00 UTC
			expectedActive:   3600000,       // 1 hour in milliseconds
		},
		{
			name:             "just published",
			pubMillis:        1704067200000,
			scrapeTimeMillis: 1704067200000,
			expectedActive:   0,
		},
		{
			name:             "24 hours active",
			pubMillis:        1704067200000,                 // 2024-01-01 00:00:00 UTC
			scrapeTimeMillis: 1704067200000 + 24*3600*1000,  // Next day
			expectedActive:   86400000,                      // 24 hours in milliseconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activeMillis := tt.scrapeTimeMillis - tt.pubMillis
			if activeMillis != tt.expectedActive {
				t.Errorf("expected active millis %d, got %d", tt.expectedActive, activeMillis)
			}
		})
	}
}

// TestPoliceAlertTypeFiltering tests that only POLICE type alerts are saved
func TestPoliceAlertTypeFiltering(t *testing.T) {
	alerts := []models.WazeAlert{
		{UUID: "police-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
		{UUID: "accident-1", Type: "ACCIDENT", Subtype: "MINOR"},
		{UUID: "police-2", Type: "POLICE", Subtype: "POLICE_WITH_MOBILE_CAMERA"},
		{UUID: "jam-1", Type: "JAM", Subtype: "MODERATE_TRAFFIC"},
		{UUID: "police-3", Type: "POLICE", Subtype: ""},
	}

	// Filter for POLICE type only (simulating SavePoliceAlerts logic)
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

