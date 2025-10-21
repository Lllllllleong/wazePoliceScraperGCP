package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/models"
	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
)

var (
	projectID      string
	collectionName string
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

	log.Printf("Starting Waze Police Alerts API on port %s", port)
	log.Printf("Project ID: %s", projectID)
	log.Printf("Collection: %s", collectionName)

	// Setup HTTP handlers
	http.HandleFunc("/api/alerts", corsMiddleware(alertsHandler))
	http.HandleFunc("/health", healthHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// corsMiddleware adds CORS headers to allow frontend access
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// alertsHandler handles requests for police alerts
func alertsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received alerts request from %s (%s)", r.RemoteAddr, r.Method)

	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// Parse request body
	var req models.AlertsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Dates) == 0 {
		sendErrorResponse(w, "At least one date is required", http.StatusBadRequest)
		return
	}

	log.Printf("Request: %d dates, %d subtypes, %d streets",
		len(req.Dates), len(req.Subtypes), len(req.Streets))

	// Create Firestore client
	firestoreClient, err := storage.NewFirestoreClient(ctx, projectID, collectionName)
	if err != nil {
		log.Printf("Error creating Firestore client: %v", err)
		sendErrorResponse(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer firestoreClient.Close()

	// Query alerts with filters
	alerts, err := firestoreClient.GetPoliceAlertsByDatesWithFilters(
		ctx,
		req.Dates,
		req.Subtypes,
		req.Streets,
	)
	if err != nil {
		log.Printf("Error querying alerts: %v", err)
		sendErrorResponse(w, fmt.Sprintf("Failed to query alerts: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := models.AlertsResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully retrieved %d alerts", len(alerts)),
		Alerts:  alerts,
		Stats: models.ResponseStats{
			TotalAlerts:      len(alerts),
			DatesQueried:     req.Dates,
			SubtypesFiltered: req.Subtypes,
			StreetsFiltered:  req.Streets,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	log.Printf("✅ Successfully returned %d alerts for %d date(s): %s",
		len(alerts), len(req.Dates), strings.Join(req.Dates, ", "))
}

// sendErrorResponse sends a JSON error response
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := models.AlertsResponse{
		Success: false,
		Message: message,
		Alerts:  []models.PoliceAlert{},
		Stats: models.ResponseStats{
			TotalAlerts: 0,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// healthHandler returns OK for health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
