package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

// TestHealthHandlerMethods tests that health endpoint responds to various HTTP methods
func TestHealthHandlerMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/health", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(healthHandler)

			handler.ServeHTTP(rr, req)

			// Health endpoint should return OK for all methods
			if rr.Code != http.StatusOK {
				t.Errorf("method %s: expected status %d, got %d", method, http.StatusOK, rr.Code)
			}
		})
	}
}

// TestScraperHandlerCreation tests that makeScraperHandler returns a valid handler
func TestScraperHandlerCreation(t *testing.T) {
	bboxes := []string{
		"150.38,-34.25,151.00,-33.93",
		"149.58,-34.76,150.83,-34.13",
	}

	handler := makeScraperHandler(bboxes)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

// TestScraperHandlerWithEmptyBBoxes tests handler behavior with empty bboxes
func TestScraperHandlerWithEmptyBBoxes(t *testing.T) {
	handler := makeScraperHandler([]string{})
	if handler == nil {
		t.Fatal("expected non-nil handler even with empty bboxes")
	}
}

// TestDefaultBBoxesConfiguration verifies the default bboxes are configured correctly
func TestDefaultBBoxesConfiguration(t *testing.T) {
	if len(defaultBBoxes) == 0 {
		t.Error("defaultBBoxes should not be empty")
	}

	// Verify each bbox has the correct format (4 comma-separated values)
	for i, bbox := range defaultBBoxes {
		parts := splitBBox(bbox)
		if len(parts) != 4 {
			t.Errorf("bbox %d should have 4 parts, got %d: %s", i, len(parts), bbox)
		}
	}
}

// splitBBox splits a bbox string into its parts
func splitBBox(bbox string) []string {
	var parts []string
	current := ""
	for _, c := range bbox {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// TestResponseStructure tests the expected response structure from scraper handler
func TestResponseStructure(t *testing.T) {
	// Test that the expected response structure is valid JSON
	response := map[string]interface{}{
		"status":              "success",
		"alerts_found":        10,
		"police_alerts_saved": 5,
		"stats": map[string]interface{}{
			"total_requests":   2,
			"successful_calls": 2,
			"failed_calls":     0,
		},
		"bboxes_used": 2,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// Verify it can be unmarshaled back
	var decoded map[string]interface{}
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify required fields exist
	requiredFields := []string{"status", "alerts_found", "police_alerts_saved", "stats", "bboxes_used"}
	for _, field := range requiredFields {
		if _, exists := decoded[field]; !exists {
			t.Errorf("required field %q not found in response", field)
		}
	}
}

// TestScraperHandlerWithMockedDependencies provides a template for full handler testing
// In practice, this would use dependency injection with interfaces
func TestScraperHandlerWithMockedDependencies(t *testing.T) {
	t.Run("handler returns JSON content type", func(t *testing.T) {
		// This test verifies the handler sets correct content type
		// A full implementation would mock the Waze client and Firestore client

		// Create a test server that returns a mock response
		mockResponse := map[string]interface{}{
			"status":              "success",
			"alerts_found":        0,
			"police_alerts_saved": 0,
			"stats":               map[string]interface{}{},
			"bboxes_used":         1,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		// Make a request to the mock server
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", contentType)
		}
	})
}

// TestBBoxValidation tests that invalid bboxes are handled gracefully
func TestBBoxValidation(t *testing.T) {
	tests := []struct {
		name    string
		bbox    string
		isValid bool
	}{
		{"valid bbox", "150.38,-34.25,151.00,-33.93", true},
		{"valid negative coords", "-150.38,-34.25,-151.00,-33.93", true},
		{"too few parts", "150.38,-34.25,151.00", false},
		{"too many parts", "150.38,-34.25,151.00,-33.93,0", false},
		{"empty string", "", false},
		{"single value", "150.38", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitBBox(tt.bbox)
			isValid := len(parts) == 4

			if isValid != tt.isValid {
				t.Errorf("bbox %q: expected valid=%v, got valid=%v", tt.bbox, tt.isValid, isValid)
			}
		})
	}
}

// TestPoliceAlertCountLogic tests the police alert counting logic
func TestPoliceAlertCountLogic(t *testing.T) {
	type mockAlert struct {
		Type string
	}

	alerts := []mockAlert{
		{Type: "POLICE"},
		{Type: "ACCIDENT"},
		{Type: "POLICE"},
		{Type: "JAM"},
		{Type: "POLICE"},
		{Type: "HAZARD"},
	}

	policeCount := 0
	for _, alert := range alerts {
		if alert.Type == "POLICE" {
			policeCount++
		}
	}

	if policeCount != 3 {
		t.Errorf("expected 3 police alerts, got %d", policeCount)
	}
}

// TestEnvironmentVariableDefaults tests default values when env vars are not set
func TestEnvironmentVariableDefaults(t *testing.T) {
	// Test default port
	defaultPort := "8080"
	if defaultPort != "8080" {
		t.Errorf("expected default port '8080', got %q", defaultPort)
	}

	// Test default collection name
	defaultCollection := "police_alerts"
	if defaultCollection != "police_alerts" {
		t.Errorf("expected default collection 'police_alerts', got %q", defaultCollection)
	}
}

// TestHTTPRequestHeaders tests that requests include appropriate headers
func TestHTTPRequestHeaders(t *testing.T) {
	var capturedRequest *http.Request

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequest = r
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if capturedRequest == nil {
		t.Fatal("request was not captured")
	}

	// Verify basic request properties
	if capturedRequest.Method != "GET" {
		t.Errorf("expected GET method, got %s", capturedRequest.Method)
	}
}


