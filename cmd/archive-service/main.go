// Package main implements the archive service for long-term data storage.
//
// This service is deployed on Google Cloud Run and triggered daily by
// Cloud Scheduler. It moves completed alert data from Firestore to Google
// Cloud Storage for cost-effective long-term archival.
//
// Key behaviors:
//   - Idempotent: Skips dates that are already archived
//   - JSONL format: Stores alerts as newline-delimited JSON
//   - Timezone-aware: Uses Australia/Canberra timezone for date boundaries
//
// Environment Variables:
//   - GCP_PROJECT_ID: Google Cloud project ID (required)
//   - FIRESTORE_COLLECTION: Firestore collection name (default: "police_alerts")
//   - GCS_BUCKET_NAME: GCS bucket for archives (required)
//   - PORT: HTTP server port (default: "8080")
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "time/tzdata"

	gcs "cloud.google.com/go/storage"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
)

type server struct {
	alertStore    storage.AlertStore
	gcsClient     storage.GCSClient
	bucketName    string
	loadLocation  func(name string) (*time.Location, error)
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
		alertStore:    firestoreClient,
		gcsClient:     &storage.GCSClientAdapter{Client: storageClient},
		bucketName:    bucketName,
		loadLocation:  time.LoadLocation,
	}

	log.Printf("Starting Archive Service on port %s", port)

	http.HandleFunc("/", s.archiveHandler)
	http.HandleFunc("/health", healthHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (s *server) archiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// Get Canberra location
	loc, err := s.loadLocation("Australia/Canberra")
	if err != nil {
		log.Printf("Error loading location: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check for a date in the request body
	var requestBody struct {
		Date string `json:"date"`
	}

	var targetDate time.Time
	if r.Body != nil {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestBody); err == nil && requestBody.Date != "" {
			targetDate, err = time.ParseInLocation("2006-01-02", requestBody.Date, loc)
			if err != nil {
				http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
				return
			}
		} else {
			// Default to yesterday if no date is provided
			targetDate = time.Now().In(loc).AddDate(0, 0, -1)
		}
	} else {
		// Default to yesterday if no body is provided
		targetDate = time.Now().In(loc).AddDate(0, 0, -1)
	}

	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)

	// Idempotency check
	fileName := fmt.Sprintf("%s.jsonl", targetDate.Format("2006-01-02"))
	obj := s.gcsClient.Bucket(s.bucketName).Object(fileName)
	_, err = obj.Attrs(ctx)
	if err == nil {
		log.Printf("Archive for %s already exists. Skipping.", targetDate.Format("2006-01-02"))
		fmt.Fprintf(w, "Archive for %s already exists. Nothing to do.", targetDate.Format("2006-01-02"))
		return
	}
	if !storage.IsObjectNotExist(err) {
		log.Printf("Error checking for existing archive: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Archiving alerts for %s (from %s to %s)", targetDate.Format("2006-01-02"), startOfDay, endOfDay)

	// Get alerts from Firestore
	alerts, err := s.alertStore.GetPoliceAlertsByDateRange(ctx, startOfDay, endOfDay)
	if err != nil {
		log.Printf("Error getting alerts from Firestore: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(alerts) == 0 {
		log.Println("No alerts to archive")
		fmt.Fprintf(w, "No alerts to archive for %s", targetDate.Format("2006-01-02"))
		return
	}

	// Create JSONL data
	jsonlData, err := createJSONL(alerts)
	if err != nil {
		log.Printf("Error creating JSONL data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Upload to GCS
	wc := obj.NewWriter(ctx)

	if _, err := wc.Write(jsonlData); err != nil {
		log.Printf("Error writing to GCS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := wc.Close(); err != nil {
		log.Printf("Error closing GCS writer: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully uploaded %s to GCS", fileName)

	fmt.Fprintf(w, "Successfully archived %d alerts for %s", len(alerts), targetDate.Format("2006-01-02"))
}

func createJSONL(alerts []models.PoliceAlert) ([]byte, error) {
	var data []byte
	for _, alert := range alerts {
		jsonData, err := json.Marshal(alert)
		if err != nil {
			return nil, err
		}
		data = append(data, jsonData...)
		data = append(data, '\n')
	}
	return data, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
