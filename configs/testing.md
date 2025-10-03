# Quick Test Guide

This guide helps you test the scraper locally before deploying.

## Test 1: Waze API Client

Test if you can fetch alerts from Waze:

```bash
cd wazePoliceScraperGCP

# Create a simple test file
cat > test_waze.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "github.com/Lllllllleong/wazePoliceScraperGCP/internal/waze"
)

func main() {
    client := waze.NewClient()
    
    // Test with Singapore bounding box
    bbox := "103.6,1.15,104.0,1.45"
    
    fmt.Printf("Testing Waze API with bbox: %s\n", bbox)
    alerts, err := client.GetAlertsMultipleBBoxes([]string{bbox})
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Printf("\n✅ Success! Found %d alerts\n", len(alerts))
    
    if len(alerts) > 0 {
        fmt.Printf("\nFirst alert:\n")
        fmt.Printf("  UUID: %s\n", alerts[0].UUID)
        fmt.Printf("  Type: %s\n", alerts[0].Type)
        fmt.Printf("  Location: %.4f, %.4f\n", 
            alerts[0].Location.Latitude, 
            alerts[0].Location.Longitude)
    }
}
EOF

# Run the test
go run test_waze.go
```

## Test 2: Local Scraper

Run the full scraper locally (requires GCP credentials):

```bash
# Set your project
export GCP_PROJECT_ID="your-project-id"

# Set bounding box (change to your region)
export WAZE_BBOXES="103.6,1.15,104.0,1.45"

# Run the scraper
go run cmd/scraper/main.go
```

In another terminal:

```bash
# Test the endpoint
curl http://localhost:8080

# Expected response:
# {
#   "status": "success",
#   "alerts_found": 123,
#   "alerts_saved": 123,
#   ...
# }
```

## Test 3: Exporter (after scraping)

```bash
# Export today's alerts
go run cmd/exporter/main.go \
  --project=your-project-id \
  --start=$(date +%Y-%m-%d) \
  --end=$(date +%Y-%m-%d) \
  --output=test_export.jsonl

# Check the output
head test_export.jsonl
```

## Test 4: Cloud Run (after deployment)

```bash
# Get your service URL
SERVICE_URL=$(gcloud run services describe waze-scraper \
  --region us-central1 \
  --format 'value(status.url)')

# Test it
curl $SERVICE_URL

# View logs
gcloud run services logs read waze-scraper \
  --region=us-central1 \
  --limit=20
```

## Troubleshooting

### "No alerts found"
- Check your bounding box coordinates are correct
- Try a larger area
- Check https://www.waze.com/live-map to see if there are alerts in your area

### "Failed to connect to Firestore"
- Make sure Firestore is enabled in your GCP project
- Check you have proper authentication (`gcloud auth application-default login`)

### "Permission denied"
- Run: `gcloud auth application-default login`
- Make sure your service account has Firestore write permissions
