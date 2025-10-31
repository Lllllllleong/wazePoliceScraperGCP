package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"

	gcs "cloud.google.com/go/storage"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
)

type server struct {
	firestoreClient *storage.FirestoreClient
	storageClient   *gcs.Client
	bucketName      string
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

	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		log.Fatal("CORS_ALLOWED_ORIGIN environment variable not set")
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

	s := &server{
		firestoreClient: firestoreClient,
		storageClient:   storageClient,
		bucketName:      bucketName,
	}

	log.Printf("Starting Alerts Service on port %s", port)

	http.HandleFunc("/police_alerts", corsMiddleware(s.alertsHandler, allowedOrigin))
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/version", versionHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func corsMiddleware(next http.HandlerFunc, allowedOrigin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Vary", "Origin") // Best practice to prevent caching issues
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
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

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	w.Header().Set("Content-Type", "application/jsonl")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

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
							buf = append(buf, readBuf[:n]...)

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
								dataChan <- line

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
				} else if err == gcs.ErrObjectNotExist {
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

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "v2")
}
