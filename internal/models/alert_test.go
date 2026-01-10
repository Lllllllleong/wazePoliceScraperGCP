package models

import (
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/type/latlng"
)

func TestLocation(t *testing.T) {
	loc := Location{
		Latitude:  -33.8688,
		Longitude: 151.2093,
	}

	if loc.Latitude != -33.8688 {
		t.Errorf("Expected latitude -33.8688, got %f", loc.Latitude)
	}
	if loc.Longitude != 151.2093 {
		t.Errorf("Expected longitude 151.2093, got %f", loc.Longitude)
	}
}

func TestWazeAlert(t *testing.T) {
	alert := WazeAlert{
		UUID:    "test-uuid-123",
		Type:    "POLICE",
		Subtype: "POLICE_VISIBLE",
		Location: Location{
			Latitude:  -33.8688,
			Longitude: 151.2093,
		},
		Street:      "George Street",
		City:        "Sydney",
		Country:     "AU",
		Reliability: 8,
		Confidence:  7,
		PubMillis:   1704067200000, // 2024-01-01 00:00:00 UTC
		NThumbsUp:   5,
	}

	if alert.UUID != "test-uuid-123" {
		t.Errorf("Expected UUID test-uuid-123, got %s", alert.UUID)
	}
	if alert.Type != "POLICE" {
		t.Errorf("Expected Type POLICE, got %s", alert.Type)
	}
	if alert.Subtype != "POLICE_VISIBLE" {
		t.Errorf("Expected Subtype POLICE_VISIBLE, got %s", alert.Subtype)
	}
	if alert.Street != "George Street" {
		t.Errorf("Expected Street George Street, got %s", alert.Street)
	}
	if alert.NThumbsUp != 5 {
		t.Errorf("Expected NThumbsUp 5, got %d", alert.NThumbsUp)
	}
}

func TestPoliceAlert(t *testing.T) {
	now := time.Now()
	publishTime := now.Add(-1 * time.Hour)
	expireTime := now.Add(1 * time.Hour)

	alert := PoliceAlert{
		UUID:    "test-police-uuid",
		Type:    "POLICE",
		Subtype: "POLICE_WITH_MOBILE_CAMERA",
		Street:  "Hume Highway",
		City:    "Goulburn",
		Country: "AU",
		LocationGeo: &latlng.LatLng{
			Latitude:  -34.7515,
			Longitude: 149.7209,
		},
		Reliability:      9,
		Confidence:       8,
		PublishTime:      publishTime,
		ScrapeTime:       now,
		ExpireTime:       expireTime,
		ActiveMillis:     7200000, // 2 hours in milliseconds
		NThumbsUpInitial: 3,
		NThumbsUpLast:    7,
	}

	if alert.UUID != "test-police-uuid" {
		t.Errorf("Expected UUID test-police-uuid, got %s", alert.UUID)
	}
	if alert.Subtype != "POLICE_WITH_MOBILE_CAMERA" {
		t.Errorf("Expected Subtype POLICE_WITH_MOBILE_CAMERA, got %s", alert.Subtype)
	}
	if alert.LocationGeo.Latitude != -34.7515 {
		t.Errorf("Expected latitude -34.7515, got %f", alert.LocationGeo.Latitude)
	}
	if alert.ActiveMillis != 7200000 {
		t.Errorf("Expected ActiveMillis 7200000, got %d", alert.ActiveMillis)
	}
	if alert.NThumbsUpLast != 7 {
		t.Errorf("Expected NThumbsUpLast 7, got %d", alert.NThumbsUpLast)
	}
}

func TestComment(t *testing.T) {
	comment := Comment{
		ReportMillis: 1704067200000,
		Text:         "Still there!",
		IsThumbsUp:   true,
	}

	if comment.ReportMillis != 1704067200000 {
		t.Errorf("Expected ReportMillis 1704067200000, got %d", comment.ReportMillis)
	}
	if comment.Text != "Still there!" {
		t.Errorf("Expected Text 'Still there!', got %s", comment.Text)
	}
	if !comment.IsThumbsUp {
		t.Error("Expected IsThumbsUp to be true")
	}
}

func TestScrapingStats(t *testing.T) {
	now := time.Now()
	stats := ScrapingStats{
		TotalRequests:     10,
		SuccessfulCalls:   8,
		FailedCalls:       2,
		TotalAlerts:       100,
		UniqueAlerts:      75,
		LastSuccessfulRun: now,
	}

	if stats.TotalRequests != 10 {
		t.Errorf("Expected TotalRequests 10, got %d", stats.TotalRequests)
	}
	if stats.SuccessfulCalls != 8 {
		t.Errorf("Expected SuccessfulCalls 8, got %d", stats.SuccessfulCalls)
	}
	if stats.FailedCalls != 2 {
		t.Errorf("Expected FailedCalls 2, got %d", stats.FailedCalls)
	}
	if stats.UniqueAlerts != 75 {
		t.Errorf("Expected UniqueAlerts 75, got %d", stats.UniqueAlerts)
	}
}
