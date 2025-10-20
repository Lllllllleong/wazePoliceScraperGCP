package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Lllllllleong/wazePoliceScraperGCP/internal/storage"
)

func main() {
	// Command-line flags
	var (
		projectID      = flag.String("project", os.Getenv("GCP_PROJECT_ID"), "GCP Project ID")
		collectionName = flag.String("collection", os.Getenv("FIRESTORE_COLLECTION"), "Firestore collection name (default: police_alerts)")
		startDate      = flag.String("start", "", "Start date (YYYY-MM-DD)")
		endDate        = flag.String("end", "", "End date (YYYY-MM-DD)")
		output         = flag.String("output", "police_alerts.jsonl", "Output file path")
		format         = flag.String("format", "jsonl", "Output format: json or jsonl")
	)

	flag.Parse()

	if *projectID == "" {
		log.Fatal("Project ID is required. Set GCP_PROJECT_ID env var or use -project flag")
	}

	if *collectionName == "" {
		*collectionName = "police_alerts" // Default collection name
	}

	if *startDate == "" || *endDate == "" {
		log.Fatal("Start and end dates are required. Format: YYYY-MM-DD")
	}

	ctx := context.Background()

	// Create Firestore client
	log.Printf("Connecting to Firestore (project: %s, collection: %s)", *projectID, *collectionName)
	firestoreClient, err := storage.NewFirestoreClient(ctx, *projectID, *collectionName)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Parse dates
	start, err := time.Parse("2006-01-02", *startDate)
	if err != nil {
		log.Fatalf("Invalid start date format: %v", err)
	}

	end, err := time.Parse("2006-01-02", *endDate)
	if err != nil {
		log.Fatalf("Invalid end date format: %v", err)
	}

	// Set end time to end of day
	end = end.Add(24*time.Hour - time.Second)

	// Fetch police alerts
	log.Printf("Fetching police alerts active from %s to %s (expire_time >= start AND publish_time <= end)", start.Format("2006-01-02"), end.Format("2006-01-02"))
	alerts, err := firestoreClient.GetPoliceAlertsByDateRange(ctx, start, end)
	if err != nil {
		log.Fatalf("Failed to fetch police alerts: %v", err)
	}

	if len(alerts) == 0 {
		log.Println("No police alerts found for the specified criteria")
		return
	}

	log.Printf("Retrieved %d police alerts, writing to %s", len(alerts), *output)

	// Write to file
	file, err := os.Create(*output)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	switch *format {
	case "json":
		// Write as JSON array
		if err := encoder.Encode(alerts); err != nil {
			log.Fatalf("Failed to write JSON: %v", err)
		}

	case "jsonl":
		// Write as JSON Lines (one object per line)
		encoder.SetIndent("", "") // No indentation for JSONL
		for _, alert := range alerts {
			if err := encoder.Encode(alert); err != nil {
				log.Fatalf("Failed to write JSONL: %v", err)
			}
		}

	default:
		log.Fatalf("Invalid format: %s (must be 'json' or 'jsonl')", *format)
	}

	log.Printf("✅ Successfully exported %d police alerts to %s", len(alerts), *output)

	// Print summary
	fmt.Println("\n📊 Export Summary:")
	fmt.Printf("   Total police alerts: %d\n", len(alerts))
	fmt.Printf("   Output file: %s\n", *output)
	fmt.Printf("   Format: %s\n", *format)

	if len(alerts) > 0 {
		fmt.Printf("\n📍 Alert lifecycle range in export:\n")
		fmt.Printf("   First published: %s\n", alerts[0].PublishTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Last expired:    %s\n", alerts[len(alerts)-1].ExpireTime.Format("2006-01-02 15:04:05"))
	}
}
