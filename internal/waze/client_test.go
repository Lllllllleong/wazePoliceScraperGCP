package waze

import (
	"testing"
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

