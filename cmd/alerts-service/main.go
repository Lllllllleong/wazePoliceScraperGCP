// Package main implements the alerts API service for serving police alert data.
//
// This service is deployed on Google Cloud Run and provides a public HTTPS API
// that serves alert data to the frontend dashboard. It implements:
//   - Firebase Anonymous Authentication for user identification
//   - Per-user rate limiting to prevent abuse
//   - GZIP compression for efficient data transfer
//   - JSONL streaming for large datasets
//   - Intelligent data sourcing from GCS archives or live Firestore
//
// Environment Variables:
//   - GCP_PROJECT_ID: Google Cloud project ID (required)
//   - FIRESTORE_COLLECTION: Firestore collection name (default: "police_alerts")
//   - GCS_BUCKET_NAME: GCS bucket for archived data (required)
//   - RATE_LIMIT_PER_MINUTE: Per-user rate limit (default: 30)
//   - PORT: HTTP server port (default: "8080")
package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	_ "time/tzdata"

	gcs "cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
	"golang.org/x/time/rate"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const uidContextKey contextKey = "uid"

// Metrics for buffer performance testing
type requestMetrics struct {
	bufferGrows    atomic.Int64
	channelBlocks  atomic.Int64
	linesProcessed atomic.Int64
	bytesProcessed atomic.Int64
	maxBufSize     atomic.Int64
	start          time.Time
}

type server struct {
	firestoreClient storage.AlertStore
	storageClient   storage.GCSClient
	bucketName      string
	firebaseAuth    storage.FirebaseAuthClient
	// Rate limiting
	limiters      map[string]*rate.Limiter
	limitersMutex sync.RWMutex
	ratePerMinute int
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable not set")
	}

	collectionName := os.Getenv("FIRESTORE_COLLECTION")
	if collectionName == "" {
		collectionName = "police_alerts"
	}

	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("GCS_BUCKET_NAME environment variable not set")
	}

	// Rate limiting configuration
	rateLimit := os.Getenv("RATE_LIMIT_PER_MINUTE")
	if rateLimit == "" {
		rateLimit = "30"
	}
	ratePerMinute, err := strconv.Atoi(rateLimit)
	if err != nil || ratePerMinute <= 0 {
		log.Fatalf("Invalid RATE_LIMIT_PER_MINUTE: %s", rateLimit)
	}

	ctx := context.Background()
	firestoreClient, err := storage.NewFirestoreClient(ctx, projectID, collectionName)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	storageClient, err := gcs.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Storage client: %v", err)
	}
	defer storageClient.Close()

	// Initialize Firebase Admin SDK
	firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}

	firebaseAuth, err := firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to create Firebase Auth client: %v", err)
	}

	// Log if using emulator (for local testing)
	if os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") != "" {
		log.Printf("Using Firebase Auth Emulator at %s", os.Getenv("FIREBASE_AUTH_EMULATOR_HOST"))
	}

	s := &server{
		firestoreClient: firestoreClient,
		storageClient:   &storage.GCSClientAdapter{Client: storageClient},
		bucketName:      bucketName,
		firebaseAuth:    &storage.FirebaseAuthClientAdapter{Client: firebaseAuth},
		limiters:        make(map[string]*rate.Limiter),
		ratePerMinute:   ratePerMinute,
	}

	// Start cleanup routine for old limiters
	go s.cleanupLimiters()

	log.Printf("Starting Alerts Service on port %s", port)
	log.Printf("Rate limit: %d requests per minute per user", ratePerMinute)
	log.Printf("Firebase Authentication: Enabled")
	http.HandleFunc("/police_alerts", corsMiddleware(s.authMiddleware(s.rateLimitMiddleware(gzipMiddleware(s.alertsHandler)))))
	http.HandleFunc("/health", healthHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type gzipResponseWriter struct {
	*gzip.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Flush() {
	w.Writer.Flush()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func gzipMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next(gzw, r)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	allowedOrigins := []string{
		"https://wazepolicescrapergcp.web.app",
		"https://wazepolicescrapergcp.firebaseapp.com",
		"https://dashboard.whyhireleong.com",
		"https://policealert.whyhireleong.com",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		originAllowed := false

		// Check against allowed production origins
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				originAllowed = true
				break
			}
		}

		// Allow localhost for local development/testing
		if !originAllowed && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("Authentication failed: Missing Authorization header from %s", r.RemoteAddr)
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("Authentication failed: Invalid Authorization header format from %s", r.RemoteAddr)
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify Firebase ID token
		token, err := s.firebaseAuth.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("Authentication failed: Invalid token from %s: %v", r.RemoteAddr, err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user ID to context for use in downstream handlers
		ctx := context.WithValue(r.Context(), uidContextKey, token.UID)
		log.Printf("Authenticated user: %s", token.UID)

		next(w, r.WithContext(ctx))
	}
}

func (s *server) getLimiter(uid string) *rate.Limiter {
	s.limitersMutex.Lock()
	defer s.limitersMutex.Unlock()

	limiter, exists := s.limiters[uid]
	if !exists {
		// Create limiter: rate per minute = events per second
		limiter = rate.NewLimiter(rate.Limit(float64(s.ratePerMinute)/60.0), s.ratePerMinute)
		s.limiters[uid] = limiter
	}
	return limiter
}

func (s *server) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by authMiddleware)
		uid, ok := r.Context().Value(uidContextKey).(string)
		if !ok || uid == "" {
			log.Printf("Rate limiting failed: No UID in context")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		limiter := s.getLimiter(uid)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded. Maximum "+strconv.Itoa(s.ratePerMinute)+" requests per minute.", http.StatusTooManyRequests)
			log.Printf("Rate limit exceeded for user: %s", uid)
			return
		}

		next(w, r)
	}
}

func (s *server) cleanupLimiters() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.limitersMutex.Lock()
		// Clear all limiters - they'll be recreated on next request
		s.limiters = make(map[string]*rate.Limiter)
		s.limitersMutex.Unlock()
		log.Println("Cleaned up rate limiters")
	}
}

func (s *server) alertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use GET", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	datesParam := r.URL.Query().Get("dates")
	if datesParam == "" {
		http.Error(w, "Missing 'dates' query parameter", http.StatusBadRequest)
		return
	}

	dateStrings := strings.Split(datesParam, ",")
	if len(dateStrings) > 7 {
		http.Error(w, "Query limited to a maximum of 7 dates.", http.StatusBadRequest)
		return
	}
	var dates []time.Time
	loc, _ := time.LoadLocation("Australia/Canberra")

	for _, ds := range dateStrings {
		t, err := time.ParseInLocation("2006-01-02", ds, loc)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid date format for '%s', use YYYY-MM-DD", ds), http.StatusBadRequest)
			return
		}
		dates = append(dates, t)
	}

	if len(dates) == 0 {
		w.Header().Set("Content-Type", "application/jsonl")
		w.WriteHeader(http.StatusOK)
		return
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	w.Header().Set("Content-Type", "application/jsonl")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Initialize metrics
	metrics := &requestMetrics{
		start: time.Now(),
	}
	defer func() {
		duration := time.Since(metrics.start)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		log.Printf("[METRICS] Request completed in %v | Lines: %d | Bytes: %d (%.2f MB) | Throughput: %.2f MB/s | Buffer grows: %d | Max buffer: %d bytes | Channel blocks: %d | Memory: Alloc=%d MB, TotalAlloc=%d MB, Sys=%d MB, NumGC=%d",
			duration,
			metrics.linesProcessed.Load(),
			metrics.bytesProcessed.Load(),
			float64(metrics.bytesProcessed.Load())/1024/1024,
			float64(metrics.bytesProcessed.Load())/1024/1024/duration.Seconds(),
			metrics.bufferGrows.Load(),
			metrics.maxBufSize.Load(),
			metrics.channelBlocks.Load(),
			m.Alloc/1024/1024,
			m.TotalAlloc/1024/1024,
			m.Sys/1024/1024,
			m.NumGC,
		)
	}()

	numWorkers := 7
	jobs := make(chan time.Time, len(dates))
	dataChan := make(chan []byte, 100) // Channel for workers to send data to the writer
	var wg sync.WaitGroup

	// Start a single writer goroutine
	writerDone := make(chan struct{})
	go func() {
		for data := range dataChan {
			if _, err := w.Write(data); err != nil {
				log.Printf("Error writing response: %v", err)
				return // Stop writing if there's an error
			}
			flusher.Flush()
		}
		close(writerDone)
	}()

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for date := range jobs {
				fileName := fmt.Sprintf("%s.jsonl", date.Format("2006-01-02"))
				obj := s.storageClient.Bucket(s.bucketName).Object(fileName)

				reader, err := obj.NewReader(ctx)
				if err == nil {
					// Archive exists - read line by line to avoid splitting JSON objects
					buf := make([]byte, 0, 64*1024) // 64KB buffer for accumulating data
					readBuf := make([]byte, 4096)

					for {
						n, readErr := reader.Read(readBuf)
						if n > 0 {
							oldCap := cap(buf)
							buf = append(buf, readBuf[:n]...)
							newCap := cap(buf)
							if newCap > oldCap {
								metrics.bufferGrows.Add(1)
							}
							if newCap > int(metrics.maxBufSize.Load()) {
								metrics.maxBufSize.Store(int64(newCap))
							}
							metrics.bytesProcessed.Add(int64(n))

							// Process complete lines
							for {
								lineEnd := -1
								for i := 0; i < len(buf); i++ {
									if buf[i] == '\n' {
										lineEnd = i
										break
									}
								}

								if lineEnd == -1 {
									// No complete line yet
									break
								}

								// Send complete line including newline
								line := make([]byte, lineEnd+1)
								copy(line, buf[:lineEnd+1])
								metrics.linesProcessed.Add(1)

								// Non-blocking send with metrics
								select {
								case dataChan <- line:
									// Sent without blocking
								default:
									metrics.channelBlocks.Add(1)
									dataChan <- line // Block if necessary
								}

								// Remove processed line from buffer
								buf = buf[lineEnd+1:]
							}
						}
						if readErr != nil {
							// Send any remaining data
							if len(buf) > 0 {
								remaining := make([]byte, len(buf))
								copy(remaining, buf)
								dataChan <- remaining
								if buf[len(buf)-1] != '\n' {
									dataChan <- []byte("\n")
								}
							}
							break
						}
					}
					reader.Close()
				} else if storage.IsObjectNotExist(err) {
					// Archive does not exist, query Firestore
					startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
					endOfDay := startOfDay.Add(24*time.Hour - time.Second)

					var alerts []models.PoliceAlert
					alerts, firestoreErr := s.firestoreClient.GetPoliceAlertsByDateRange(ctx, startOfDay, endOfDay)
					if firestoreErr != nil {
						log.Printf("Error getting alerts from Firestore for %s: %v", date.Format("2006-01-02"), firestoreErr)
						continue
					}
					for _, alert := range alerts {
						jsonData, marshalErr := json.Marshal(alert)
						if marshalErr != nil {
							log.Printf("Error marshaling alert %s: %v", alert.UUID, marshalErr)
							continue
						}
						dataChan <- append(jsonData, '\n')
					}
				} else {
					log.Printf("Error checking for archive %s: %v", fileName, err)
				}
			}
		}()
	}

	// Send jobs
	for _, date := range dates {
		jobs <- date
	}
	close(jobs)

	// Wait for all workers to finish, then close the data channel and wait for the writer
	wg.Wait()
	close(dataChan)
	<-writerDone
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
