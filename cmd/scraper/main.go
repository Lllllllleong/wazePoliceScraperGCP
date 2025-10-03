package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/waze"
)

var (
	projectID      string
	collectionName string
	// Default bounding boxes - Sydney to Canberra (Hume Highway)
	// Format: "west,south,east,north"
	defaultBBoxes = []string{
		"150.38822599217056,-34.254577954626086,151.00867887302994,-33.937977044844004", // Hume Highway - Sydney
		"149.58926145838367,-34.76915040190209,150.83016722010242,-34.138639582841435",  // Hume Highway - Middle
		"149.09281124417694,-35.21080621952668,150.3337170058957,-34.583661538587855",   // Hume Highway - Canberra
		"148.80885598970738,-35.4530012424677,149.42930887056676,-35.14096097196958",    // Canberra
	}
)

func main() {
	// Get configuration from environment
	projectID = os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is required")
	}

	collectionName = os.Getenv("FIRESTORE_COLLECTION")
	if collectionName == "" {
		collectionName = "police_alerts" // Default collection name
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Override bboxes from environment if provided
	bboxesEnv := os.Getenv("WAZE_BBOXES")
	bboxes := defaultBBoxes
	if bboxesEnv != "" {
		bboxes = strings.Split(bboxesEnv, ";")
	}

	log.Printf("Starting Waze Scraper on port %s", port)
	log.Printf("Project ID: %s", projectID)
	log.Printf("Collection: %s", collectionName)
	log.Printf("Bounding boxes: %v", bboxes)

	// Setup HTTP handlers
	http.HandleFunc("/", makeScraperHandler(bboxes))
	http.HandleFunc("/health", healthHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func makeScraperHandler(bboxes []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received scrape request from %s", r.RemoteAddr)

		ctx := context.Background()

		// Step 1: Create Waze client and fetch alerts
		wazeClient := waze.NewClient()
		alerts, err := wazeClient.GetAlertsMultipleBBoxes(bboxes)
		if err != nil {
			log.Printf("Error fetching alerts: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch alerts: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Fetched %d unique alerts from Waze", len(alerts))

		// Step 2: Save police alerts to Firestore
		firestoreClient, err := storage.NewFirestoreClient(ctx, projectID, collectionName)
		if err != nil {
			log.Printf("Error creating Firestore client: %v", err)
			http.Error(w, fmt.Sprintf("Failed to connect to Firestore: %v", err), http.StatusInternalServerError)
			return
		}
		defer firestoreClient.Close()

		scrapeTime := time.Now()
		err = firestoreClient.SavePoliceAlerts(ctx, alerts, scrapeTime)
		if err != nil {
			log.Printf("Error saving police alerts to Firestore: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save alerts: %v", err), http.StatusInternalServerError)
			return
		}

		// Count how many police alerts were actually saved
		policeCount := 0
		for _, alert := range alerts {
			if alert.Type == "POLICE" {
				policeCount++
			}
		}

		// Step 3: Return success response
		stats := wazeClient.GetStats()
		response := map[string]interface{}{
			"status":              "success",
			"alerts_found":        len(alerts),
			"police_alerts_saved": policeCount,
			"stats":               stats,
			"bboxes_used":         len(bboxes),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		log.Printf("✅ Successfully scraped and saved %d police alerts (out of %d total alerts)", policeCount, len(alerts))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
