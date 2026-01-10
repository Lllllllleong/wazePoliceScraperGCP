package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
	"golang.org/x/time/rate"
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

// TestCorsMiddleware tests CORS header handling
func TestCorsMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectAllowed  bool
		expectedOrigin string
	}{
		{
			name:           "production domain allowed",
			origin:         "https://wazepolicescrapergcp.web.app",
			expectAllowed:  true,
			expectedOrigin: "https://wazepolicescrapergcp.web.app",
		},
		{
			name:           "firebase app domain allowed",
			origin:         "https://wazepolicescrapergcp.firebaseapp.com",
			expectAllowed:  true,
			expectedOrigin: "https://wazepolicescrapergcp.firebaseapp.com",
		},
		{
			name:           "dashboard domain allowed",
			origin:         "https://dashboard.whyhireleong.com",
			expectAllowed:  true,
			expectedOrigin: "https://dashboard.whyhireleong.com",
		},
		{
			name:           "policealert domain allowed",
			origin:         "https://policealert.whyhireleong.com",
			expectAllowed:  true,
			expectedOrigin: "https://policealert.whyhireleong.com",
		},
		{
			name:           "localhost allowed for development",
			origin:         "http://localhost:3000",
			expectAllowed:  true,
			expectedOrigin: "http://localhost:3000",
		},
		{
			name:           "127.0.0.1 allowed for development",
			origin:         "http://127.0.0.1:8080",
			expectAllowed:  true,
			expectedOrigin: "http://127.0.0.1:8080",
		},
		{
			name:          "unknown domain not allowed",
			origin:        "https://malicious-site.com",
			expectAllowed: false,
		},
		{
			name:          "no origin header",
			origin:        "",
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple handler that the middleware wraps
			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with CORS middleware
			handler := corsMiddleware(innerHandler)

			req, err := http.NewRequest("GET", "/police_alerts", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")

			if tt.expectAllowed {
				if allowOrigin != tt.expectedOrigin {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, allowOrigin)
				}
			} else {
				if allowOrigin != "" {
					t.Errorf("expected no Access-Control-Allow-Origin, got %q", allowOrigin)
				}
			}

			// Check other CORS headers
			if rr.Header().Get("Access-Control-Allow-Methods") != "GET, OPTIONS" {
				t.Errorf("unexpected Allow-Methods header: %s", rr.Header().Get("Access-Control-Allow-Methods"))
			}
		})
	}
}

// TestCorsMiddlewarePreflightRequest tests OPTIONS preflight requests
func TestCorsMiddlewarePreflightRequest(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("should not reach here for OPTIONS"))
	})

	handler := corsMiddleware(innerHandler)

	req, err := http.NewRequest("OPTIONS", "/police_alerts", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Origin", "https://wazepolicescrapergcp.web.app")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d for preflight, got %d", http.StatusOK, rr.Code)
	}

	// Body should be empty for preflight
	if rr.Body.Len() != 0 {
		t.Errorf("expected empty body for preflight, got %q", rr.Body.String())
	}
}

// TestGzipMiddleware tests GZIP compression
func TestGzipMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		expectGzip     bool
	}{
		{
			name:           "client accepts gzip",
			acceptEncoding: "gzip, deflate",
			expectGzip:     true,
		},
		{
			name:           "client accepts only gzip",
			acceptEncoding: "gzip",
			expectGzip:     true,
		},
		{
			name:           "client does not accept gzip",
			acceptEncoding: "deflate",
			expectGzip:     false,
		},
		{
			name:           "no accept-encoding header",
			acceptEncoding: "",
			expectGzip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBody := "This is a test response that should be compressed if gzip is accepted"

			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(testBody))
			})

			handler := gzipMiddleware(innerHandler)

			req, err := http.NewRequest("GET", "/police_alerts", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if tt.expectGzip {
				if rr.Header().Get("Content-Encoding") != "gzip" {
					t.Errorf("expected Content-Encoding 'gzip', got %q", rr.Header().Get("Content-Encoding"))
				}

				// Verify response can be decompressed
				reader, err := gzip.NewReader(rr.Body)
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				defer reader.Close()

				decompressed, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("failed to read gzip body: %v", err)
				}

				if string(decompressed) != testBody {
					t.Errorf("decompressed body mismatch: expected %q, got %q", testBody, string(decompressed))
				}
			} else {
				if rr.Header().Get("Content-Encoding") == "gzip" {
					t.Error("did not expect gzip encoding")
				}

				if rr.Body.String() != testBody {
					t.Errorf("body mismatch: expected %q, got %q", testBody, rr.Body.String())
				}
			}
		})
	}
}

// TestRateLimiting tests the rate limiting functionality
func TestRateLimiting(t *testing.T) {
	// Create a server instance for testing
	s := &server{
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 5, // Low limit for testing
	}

	// Test getLimiter creates new limiter for new user
	limiter1 := s.getLimiter("user1")
	if limiter1 == nil {
		t.Fatal("expected non-nil limiter")
	}

	// Test getLimiter returns same limiter for same user
	limiter1Again := s.getLimiter("user1")
	if limiter1 != limiter1Again {
		t.Error("expected same limiter instance for same user")
	}

	// Test different users get different limiters
	limiter2 := s.getLimiter("user2")
	if limiter1 == limiter2 {
		t.Error("expected different limiter instances for different users")
	}
}

// TestRateLimitMiddleware tests rate limit enforcement
func TestRateLimitMiddleware(t *testing.T) {
	s := &server{
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 2, // Very low limit for testing
	}

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	handler := s.rateLimitMiddleware(innerHandler)

	// First request should succeed
	req1, _ := http.NewRequest("GET", "/police_alerts", nil)
	ctx1 := context.WithValue(req1.Context(), uidContextKey, "test-user")
	req1 = req1.WithContext(ctx1)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("first request: expected status %d, got %d", http.StatusOK, rr1.Code)
	}

	// Second request should succeed
	req2, _ := http.NewRequest("GET", "/police_alerts", nil)
	ctx2 := context.WithValue(req2.Context(), uidContextKey, "test-user")
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("second request: expected status %d, got %d", http.StatusOK, rr2.Code)
	}
}

// TestRateLimitMiddlewareNoAuth tests rate limit fails without auth
func TestRateLimitMiddlewareNoAuth(t *testing.T) {
	s := &server{
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 30,
	}

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := s.rateLimitMiddleware(innerHandler)

	req, _ := http.NewRequest("GET", "/police_alerts", nil)
	// No UID in context
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d without auth, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestAlertsHandlerMethodNotAllowed tests that only GET is allowed
func TestAlertsHandlerMethodNotAllowed(t *testing.T) {
	s := &server{}

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/police_alerts", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(s.alertsHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("method %s: expected status %d, got %d", method, http.StatusMethodNotAllowed, rr.Code)
			}
		})
	}
}

// TestAlertsHandlerMissingDates tests handler returns error without dates parameter
func TestAlertsHandlerMissingDates(t *testing.T) {
	s := &server{}

	req, err := http.NewRequest("GET", "/police_alerts", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for missing dates, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "dates") {
		t.Errorf("expected error message to mention 'dates', got %q", rr.Body.String())
	}
}

// TestAlertsHandlerInvalidDateFormat tests handler returns error for invalid date format
func TestAlertsHandlerInvalidDateFormat(t *testing.T) {
	s := &server{}

	req, err := http.NewRequest("GET", "/police_alerts?dates=invalid-date", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid date, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestAlertsHandlerTooManyDates tests handler returns error for too many dates
func TestAlertsHandlerTooManyDates(t *testing.T) {
	s := &server{}

	// Create a request with 8 dates (exceeds the 7 date limit)
	dates := "2024-01-01,2024-01-02,2024-01-03,2024-01-04,2024-01-05,2024-01-06,2024-01-07,2024-01-08"
	req, err := http.NewRequest("GET", "/police_alerts?dates="+dates, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for too many dates, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "7") {
		t.Errorf("expected error message to mention '7' date limit, got %q", rr.Body.String())
	}
}

// TestDateParsing tests date parsing logic
func TestDateParsing(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{"valid date", "2024-01-15", false},
		{"another valid date", "2025-12-31", false},
		{"leap year date", "2024-02-29", false},
		{"invalid format DD-MM-YYYY", "15-01-2024", true},
		{"invalid format MM/DD/YYYY", "01/15/2024", true},
		{"invalid day", "2024-01-32", true},
		{"invalid month", "2024-13-01", true},
		{"text instead of date", "not-a-date", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, _ := time.LoadLocation("Australia/Canberra")
			_, err := time.ParseInLocation("2006-01-02", tt.dateStr, loc)

			if tt.expectError && err == nil {
				t.Errorf("expected error for date %q, got nil", tt.dateStr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for date %q: %v", tt.dateStr, err)
			}
		})
	}
}

// TestContextKey tests the context key type
func TestContextKey(t *testing.T) {
	// Verify context key is properly typed
	ctx := context.Background()
	testUID := "test-user-id"

	// Set value in context
	ctx = context.WithValue(ctx, uidContextKey, testUID)

	// Retrieve value
	uid, ok := ctx.Value(uidContextKey).(string)
	if !ok {
		t.Error("expected string value from context")
	}
	if uid != testUID {
		t.Errorf("expected UID %q, got %q", testUID, uid)
	}

	// Verify wrong key returns nil
	wrongKey := contextKey("wrong")
	val := ctx.Value(wrongKey)
	if val != nil {
		t.Error("expected nil for wrong key")
	}
}

// TestAuthMiddlewareMissingHeader tests auth middleware rejects missing auth header
func TestAuthMiddlewareMissingHeader(t *testing.T) {
	s := &server{}

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(innerHandler)

	req, _ := http.NewRequest("GET", "/police_alerts", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for missing auth header, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestAuthMiddlewareInvalidFormat tests auth middleware rejects invalid auth format
func TestAuthMiddlewareInvalidFormat(t *testing.T) {
	s := &server{}

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(innerHandler)

	// Test cases that should be rejected at the format check stage
	// (before attempting Firebase token verification)
	tests := []struct {
		name       string
		authHeader string
	}{
		{"no Bearer prefix", "some-token"},
		{"Basic auth", "Basic dXNlcjpwYXNz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/police_alerts", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d for auth header %q, got %d", http.StatusUnauthorized, tt.authHeader, rr.Code)
			}
		})
	}
}

// TestGzipResponseWriter tests the gzip response writer implementation
func TestGzipResponseWriter(t *testing.T) {
	// Test that gzipResponseWriter properly implements required interfaces
	rr := httptest.NewRecorder()
	gz := gzip.NewWriter(rr)
	defer gz.Close()

	gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: rr}

	// Test Header method
	gzw.Header().Set("X-Test", "value")
	if gzw.Header().Get("X-Test") != "value" {
		t.Error("Header method not working correctly")
	}

	// Test WriteHeader method
	gzw.WriteHeader(http.StatusCreated)
	if rr.Code != http.StatusCreated {
		t.Errorf("WriteHeader not working: expected %d, got %d", http.StatusCreated, rr.Code)
	}

	// Test Write method
	testData := []byte("test data")
	n, err := gzw.Write(testData)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write returned wrong length: expected %d, got %d", len(testData), n)
	}
}

// TestJSONLResponseFormat tests that response is in JSONL format
func TestJSONLResponseFormat(t *testing.T) {
	// Create sample JSONL data
	alerts := []map[string]interface{}{
		{"UUID": "alert-1", "Type": "POLICE"},
		{"UUID": "alert-2", "Type": "POLICE"},
		{"UUID": "alert-3", "Type": "POLICE"},
	}

	var jsonlData strings.Builder
	for _, alert := range alerts {
		line, _ := json.Marshal(alert)
		jsonlData.Write(line)
		jsonlData.WriteByte('\n')
	}

	// Verify JSONL format (each line is valid JSON)
	lines := strings.Split(strings.TrimSpace(jsonlData.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 JSONL lines, got %d", len(lines))
	}

	for i, line := range lines {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(line), &parsed); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

// TestCleanupLimitersCreation tests that cleanup routine can be started
func TestCleanupLimitersCreation(t *testing.T) {
	s := &server{
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 30,
	}

	// Add some limiters
	s.getLimiter("user1")
	s.getLimiter("user2")
	s.getLimiter("user3")

	if len(s.limiters) != 3 {
		t.Errorf("expected 3 limiters, got %d", len(s.limiters))
	}

	// Manually clear limiters (simulating cleanup)
	s.limitersMutex.Lock()
	s.limiters = make(map[string]*rate.Limiter)
	s.limitersMutex.Unlock()

	if len(s.limiters) != 0 {
		t.Errorf("expected 0 limiters after cleanup, got %d", len(s.limiters))
	}
}

// TestVaryHeaderSet tests that Vary header is set correctly for CORS
func TestVaryHeaderSet(t *testing.T) {
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware(innerHandler)

	req, _ := http.NewRequest("GET", "/police_alerts", nil)
	req.Header.Set("Origin", "https://wazepolicescrapergcp.web.app")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	vary := rr.Header().Get("Vary")
	if vary != "Origin" {
		t.Errorf("expected Vary header 'Origin', got %q", vary)
	}
}

// =============================================================================
// Tests using dependency injection with mocks
// =============================================================================

// TestAlertsHandlerWithGCSArchive tests the handler when data is available from GCS archive
func TestAlertsHandlerWithGCSArchive(t *testing.T) {
	// Create mock data that would be in a GCS archive file
	archiveData := `{"UUID":"alert-1","Type":"POLICE_VISIBLE","SubType":"POLICE_VISIBLE","PubMillis":1704067200000}
{"UUID":"alert-2","Type":"POLICE_VISIBLE","SubType":"POLICE_VISIBLE","PubMillis":1704067201000}
`

	// Create mock GCS client that returns archive data
	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return &storage.MockGCSBucketHandle{
				ObjectFunc: func(objName string) storage.GCSObjectHandle {
					return &storage.MockGCSObjectHandle{
						NewReaderFunc: func(ctx context.Context) (io.ReadCloser, error) {
							return io.NopCloser(strings.NewReader(archiveData)), nil
						},
					}
				},
			}
		},
	}

	s := &server{
		firestoreClient: &storage.MockAlertStore{},
		storageClient:   mockGCS,
		bucketName:      "test-bucket",
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   30,
	}

	req, err := http.NewRequest("GET", "/police_alerts?dates=2024-01-01", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify JSONL content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/jsonl" {
		t.Errorf("expected Content-Type 'application/jsonl', got %q", contentType)
	}

	// Verify response contains the archive data
	body := rr.Body.String()
	if !strings.Contains(body, "alert-1") {
		t.Errorf("expected response to contain 'alert-1', got %q", body)
	}
	if !strings.Contains(body, "alert-2") {
		t.Errorf("expected response to contain 'alert-2', got %q", body)
	}
}

// TestAlertsHandlerFirestoreFallback tests the handler falls back to Firestore when GCS archive doesn't exist
func TestAlertsHandlerFirestoreFallback(t *testing.T) {
	// Create mock Firestore client that returns alerts
	mockStore := &storage.MockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, startDate, endDate time.Time) ([]models.PoliceAlert, error) {
			return []models.PoliceAlert{
				{
					UUID:        "firestore-alert-1",
					Type:        "POLICE_VISIBLE",
					Subtype:     "POLICE_VISIBLE",
					PublishTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	// Create mock GCS client that returns object not found
	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return &storage.MockGCSBucketHandle{
				ObjectFunc: func(objName string) storage.GCSObjectHandle {
					return &storage.MockGCSObjectHandle{
						NewReaderFunc: func(ctx context.Context) (io.ReadCloser, error) {
							return nil, storage.ErrObjectNotExist
						},
					}
				},
			}
		},
	}

	s := &server{
		firestoreClient: mockStore,
		storageClient:   mockGCS,
		bucketName:      "test-bucket",
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   30,
	}

	req, err := http.NewRequest("GET", "/police_alerts?dates=2024-01-01", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify Firestore was called
	if mockStore.CallLog.GetPoliceAlertsByDateRangeCalls == 0 {
		t.Error("expected Firestore GetPoliceAlertsByDateRange to be called")
	}

	// Verify response contains the Firestore data
	body := rr.Body.String()
	if !strings.Contains(body, "firestore-alert-1") {
		t.Errorf("expected response to contain 'firestore-alert-1', got %q", body)
	}
}

// TestAlertsHandlerMultipleDates tests the handler with multiple dates
func TestAlertsHandlerMultipleDates(t *testing.T) {
	archiveCallCount := 0

	// Create mock GCS client that tracks calls
	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return &storage.MockGCSBucketHandle{
				ObjectFunc: func(objName string) storage.GCSObjectHandle {
					return &storage.MockGCSObjectHandle{
						NewReaderFunc: func(ctx context.Context) (io.ReadCloser, error) {
							archiveCallCount++
							data := fmt.Sprintf(`{"UUID":"alert-for-%s","Type":"POLICE_VISIBLE"}
`, strings.TrimSuffix(objName, ".jsonl"))
							return io.NopCloser(strings.NewReader(data)), nil
						},
					}
				},
			}
		},
	}

	s := &server{
		firestoreClient: &storage.MockAlertStore{},
		storageClient:   mockGCS,
		bucketName:      "test-bucket",
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   30,
	}

	req, err := http.NewRequest("GET", "/police_alerts?dates=2024-01-01,2024-01-02,2024-01-03", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify GCS was called for each date
	if archiveCallCount != 3 {
		t.Errorf("expected 3 GCS calls, got %d", archiveCallCount)
	}
}

// TestAlertsHandlerEmptyArchive tests the handler with an empty archive file
func TestAlertsHandlerEmptyArchive(t *testing.T) {
	// Create mock GCS client that returns empty data
	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return &storage.MockGCSBucketHandle{
				ObjectFunc: func(objName string) storage.GCSObjectHandle {
					return &storage.MockGCSObjectHandle{
						NewReaderFunc: func(ctx context.Context) (io.ReadCloser, error) {
							return io.NopCloser(strings.NewReader("")), nil
						},
					}
				},
			}
		},
	}

	s := &server{
		firestoreClient: &storage.MockAlertStore{},
		storageClient:   mockGCS,
		bucketName:      "test-bucket",
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   30,
	}

	req, err := http.NewRequest("GET", "/police_alerts?dates=2024-01-01", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// TestAlertsHandlerGCSError tests the handler when GCS returns an unexpected error
func TestAlertsHandlerGCSError(t *testing.T) {
	// Create mock GCS client that returns an error
	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return &storage.MockGCSBucketHandle{
				ObjectFunc: func(objName string) storage.GCSObjectHandle {
					return &storage.MockGCSObjectHandle{
						NewReaderFunc: func(ctx context.Context) (io.ReadCloser, error) {
							return nil, errors.New("unexpected GCS error")
						},
					}
				},
			}
		},
	}

	s := &server{
		firestoreClient: &storage.MockAlertStore{},
		storageClient:   mockGCS,
		bucketName:      "test-bucket",
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   30,
	}

	req, err := http.NewRequest("GET", "/police_alerts?dates=2024-01-01", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	// The handler should still return OK but with no data (error is logged)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// TestAlertsHandlerFirestoreError tests the handler when Firestore returns an error
func TestAlertsHandlerFirestoreError(t *testing.T) {
	// Create mock Firestore client that returns an error
	mockStore := &storage.MockAlertStore{
		GetPoliceAlertsByDateRangeFunc: func(ctx context.Context, startDate, endDate time.Time) ([]models.PoliceAlert, error) {
			return nil, errors.New("firestore connection error")
		},
	}

	// Create mock GCS client that returns object not found (to trigger Firestore fallback)
	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return &storage.MockGCSBucketHandle{
				ObjectFunc: func(objName string) storage.GCSObjectHandle {
					return &storage.MockGCSObjectHandle{
						NewReaderFunc: func(ctx context.Context) (io.ReadCloser, error) {
							return nil, storage.ErrObjectNotExist
						},
					}
				},
			}
		},
	}

	s := &server{
		firestoreClient: mockStore,
		storageClient:   mockGCS,
		bucketName:      "test-bucket",
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   30,
	}

	req, err := http.NewRequest("GET", "/police_alerts?dates=2024-01-01", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.alertsHandler)
	handler.ServeHTTP(rr, req)

	// The handler should still return OK but with no data (error is logged)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// =============================================================================
// Auth Middleware Tests with Mocked Firebase Auth
// =============================================================================

// TestAuthMiddlewareValidToken tests auth middleware accepts valid token
func TestAuthMiddlewareValidToken(t *testing.T) {
	mockAuth := &storage.MockFirebaseAuthClient{
		VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*storage.FirebaseToken, error) {
			if idToken == "valid-token" {
				return &storage.FirebaseToken{UID: "test-user-123"}, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	s := &server{
		firebaseAuth:  mockAuth,
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 30,
	}

	// Track if inner handler was called and check context
	var capturedUID string
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := r.Context().Value(uidContextKey).(string)
		if ok {
			capturedUID = uid
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(innerHandler)

	req, _ := http.NewRequest("GET", "/police_alerts", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify UID was set in context
	if capturedUID != "test-user-123" {
		t.Errorf("expected UID 'test-user-123' in context, got %q", capturedUID)
	}

	// Verify mock was called
	if mockAuth.CallLog.VerifyIDTokenCalls != 1 {
		t.Errorf("expected 1 VerifyIDToken call, got %d", mockAuth.CallLog.VerifyIDTokenCalls)
	}
	if mockAuth.CallLog.LastToken != "valid-token" {
		t.Errorf("expected last token 'valid-token', got %q", mockAuth.CallLog.LastToken)
	}
}

// TestAuthMiddlewareInvalidToken tests auth middleware rejects invalid token
func TestAuthMiddlewareInvalidToken(t *testing.T) {
	mockAuth := &storage.MockFirebaseAuthClient{
		VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*storage.FirebaseToken, error) {
			return nil, errors.New("token has expired")
		},
	}

	s := &server{
		firebaseAuth:  mockAuth,
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 30,
	}

	innerHandlerCalled := false
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(innerHandler)

	req, _ := http.NewRequest("GET", "/police_alerts", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Verify inner handler was not called
	if innerHandlerCalled {
		t.Error("inner handler should not have been called for invalid token")
	}
}

// TestAuthMiddlewareEmptyBearerToken tests auth middleware rejects empty Bearer token
func TestAuthMiddlewareEmptyBearerToken(t *testing.T) {
	mockAuth := &storage.MockFirebaseAuthClient{
		VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*storage.FirebaseToken, error) {
			if idToken == "" {
				return nil, errors.New("empty token")
			}
			return &storage.FirebaseToken{UID: "user"}, nil
		},
	}

	s := &server{
		firebaseAuth:  mockAuth,
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 30,
	}

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(innerHandler)

	req, _ := http.NewRequest("GET", "/police_alerts", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Empty token should be rejected by Firebase verification
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for empty bearer token, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// =============================================================================
// Integration-style Tests with Full Middleware Chain
// =============================================================================

// TestFullMiddlewareChain tests the complete middleware chain with mocks
func TestFullMiddlewareChain(t *testing.T) {
	// Create mock Firebase Auth
	mockAuth := &storage.MockFirebaseAuthClient{
		VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*storage.FirebaseToken, error) {
			return &storage.FirebaseToken{UID: "integration-test-user"}, nil
		},
	}

	// Create mock GCS client
	archiveData := `{"UUID":"integration-alert","Type":"POLICE_VISIBLE"}
`
	mockGCS := &storage.MockGCSClient{
		BucketFunc: func(name string) storage.GCSBucketHandle {
			return &storage.MockGCSBucketHandle{
				ObjectFunc: func(objName string) storage.GCSObjectHandle {
					return &storage.MockGCSObjectHandle{
						NewReaderFunc: func(ctx context.Context) (io.ReadCloser, error) {
							return io.NopCloser(strings.NewReader(archiveData)), nil
						},
					}
				},
			}
		},
	}

	s := &server{
		firestoreClient: &storage.MockAlertStore{},
		storageClient:   mockGCS,
		bucketName:      "test-bucket",
		firebaseAuth:    mockAuth,
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   30,
	}

	// Build full middleware chain
	handler := corsMiddleware(s.authMiddleware(s.rateLimitMiddleware(gzipMiddleware(s.alertsHandler))))

	req, _ := http.NewRequest("GET", "/police_alerts?dates=2024-01-01", nil)
	req.Header.Set("Origin", "https://wazepolicescrapergcp.web.app")
	req.Header.Set("Authorization", "Bearer valid-test-token")
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify CORS header
	if rr.Header().Get("Access-Control-Allow-Origin") != "https://wazepolicescrapergcp.web.app" {
		t.Errorf("expected CORS origin header, got %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}

	// Verify gzip encoding
	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding, got %q", rr.Header().Get("Content-Encoding"))
	}

	// Decompress and verify content
	reader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read gzip body: %v", err)
	}

	if !strings.Contains(string(body), "integration-alert") {
		t.Errorf("expected response to contain 'integration-alert', got %q", string(body))
	}
}

// TestRateLimitExceeded tests that rate limiting is enforced
func TestRateLimitExceeded(t *testing.T) {
	mockAuth := &storage.MockFirebaseAuthClient{
		VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*storage.FirebaseToken, error) {
			return &storage.FirebaseToken{UID: "rate-limit-test-user"}, nil
		},
	}

	s := &server{
		firebaseAuth:  mockAuth,
		limiters:      make(map[string]*rate.Limiter),
		ratePerMinute: 1, // Very low limit for testing
	}

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(s.rateLimitMiddleware(innerHandler))

	// First request should succeed
	req1, _ := http.NewRequest("GET", "/police_alerts", nil)
	req1.Header.Set("Authorization", "Bearer token")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("first request: expected status %d, got %d", http.StatusOK, rr1.Code)
	}

	// Second request should be rate limited (burst of 1 used up)
	// With rate of 1/minute and burst of 1, the second immediate request should fail
	req2, _ := http.NewRequest("GET", "/police_alerts", nil)
	req2.Header.Set("Authorization", "Bearer token")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	// Note: The rate limiter has burst=ratePerMinute=1, so second request should fail
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected status %d (rate limited), got %d", http.StatusTooManyRequests, rr2.Code)
	}

	// Verify Retry-After header
	if rr2.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After header '60', got %q", rr2.Header().Get("Retry-After"))
	}
}
