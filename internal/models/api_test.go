package models

import (
	"testing"
)

func TestAlertsRequest(t *testing.T) {
	req := AlertsRequest{
		Dates:    []string{"2024-01-01", "2024-01-02", "2024-01-03"},
		Subtypes: []string{"POLICE_VISIBLE", "POLICE_WITH_MOBILE_CAMERA"},
		Streets:  []string{"Hume Highway", "Federal Highway"},
	}

	if len(req.Dates) != 3 {
		t.Errorf("Expected 3 dates, got %d", len(req.Dates))
	}
	if req.Dates[0] != "2024-01-01" {
		t.Errorf("Expected first date 2024-01-01, got %s", req.Dates[0])
	}
	if len(req.Subtypes) != 2 {
		t.Errorf("Expected 2 subtypes, got %d", len(req.Subtypes))
	}
	if len(req.Streets) != 2 {
		t.Errorf("Expected 2 streets, got %d", len(req.Streets))
	}
}

func TestAlertsRequestEmpty(t *testing.T) {
	req := AlertsRequest{
		Dates: []string{"2024-01-01"},
	}

	if len(req.Dates) != 1 {
		t.Errorf("Expected 1 date, got %d", len(req.Dates))
	}
	if req.Subtypes != nil && len(req.Subtypes) != 0 {
		t.Errorf("Expected empty Subtypes, got %v", req.Subtypes)
	}
	if req.Streets != nil && len(req.Streets) != 0 {
		t.Errorf("Expected empty Streets, got %v", req.Streets)
	}
}

func TestAlertsResponse(t *testing.T) {
	resp := AlertsResponse{
		Success: true,
		Message: "Retrieved 50 alerts",
		Alerts:  []PoliceAlert{},
		Stats: ResponseStats{
			TotalAlerts:  50,
			DatesQueried: []string{"2024-01-01", "2024-01-02"},
		},
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Message != "Retrieved 50 alerts" {
		t.Errorf("Expected message 'Retrieved 50 alerts', got %s", resp.Message)
	}
	if resp.Stats.TotalAlerts != 50 {
		t.Errorf("Expected TotalAlerts 50, got %d", resp.Stats.TotalAlerts)
	}
}

func TestResponseStats(t *testing.T) {
	stats := ResponseStats{
		TotalAlerts:      100,
		DatesQueried:     []string{"2024-01-01", "2024-01-02", "2024-01-03"},
		SubtypesFiltered: []string{"POLICE_VISIBLE"},
		StreetsFiltered:  []string{"Hume Highway"},
	}

	if stats.TotalAlerts != 100 {
		t.Errorf("Expected TotalAlerts 100, got %d", stats.TotalAlerts)
	}
	if len(stats.DatesQueried) != 3 {
		t.Errorf("Expected 3 dates queried, got %d", len(stats.DatesQueried))
	}
	if len(stats.SubtypesFiltered) != 1 {
		t.Errorf("Expected 1 subtype filter, got %d", len(stats.SubtypesFiltered))
	}
	if stats.StreetsFiltered[0] != "Hume Highway" {
		t.Errorf("Expected street filter 'Hume Highway', got %s", stats.StreetsFiltered[0])
	}
}

