# Waze Police Scraper GCP

A Google Cloud Platform application that scrapes Waze police alert data and stores it in Firestore with full lifecycle tracking.

## 🏗️ Architecture

This project consists of three main components:

1. **Cloud Run Service** (`cmd/scraper`): HTTP endpoint that scrapes Waze API and stores police alerts in Firestore
2. **Cloud Scheduler**: Triggers the scraper automatically (every 2 minutes by default)
3. **Exporter Tool** (`cmd/exporter`): CLI utility to export Firestore data to JSON/JSONL files

```
Cloud Scheduler (every 2 min)
         ↓
    Cloud Run Service
         ↓
    Waze API → Filter POLICE alerts → Firestore (with lifecycle tracking)
                                           ↓
                                    Exporter Tool → JSON/JSONL files
```

## ✨ Features

- **Automated Scraping**: Scheduled scraping of Waze alerts via Cloud Scheduler
- **Police Alert Tracking**: Filters and tracks only POLICE type alerts with full lifecycle
- **Lifecycle Management**: Tracks first seen, last seen, and expiration times
- **Configurable Regions**: Support for multiple bounding boxes (default: Sydney-Canberra)
- **Data Export**: Export alerts by date range to JSON/JSONL
- **Cloud-Native**: Runs on Google Cloud Run with Firestore storage

## 📁 Project Structure

```
wazePoliceScraperGCP/
├── cmd/
│   ├── scraper/main.go      # Cloud Run HTTP service (scrapes & saves)
│   └── exporter/main.go     # CLI tool to export data
├── internal/
│   ├── waze/client.go       # Waze API client
│   ├── storage/firestore.go # Firestore operations
│   └── models/alert.go      # Data models
├── configs/
│   └── scheduler-setup.md   # Cloud Scheduler setup guide
├── scripts/
│   └── deploy.sh            # Deployment script
├── Dockerfile               # Cloud Run container
├── .env.example             # Configuration template
└── go.mod                   # Go dependencies
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Google Cloud SDK (`gcloud` CLI) - [Install here](https://cloud.google.com/sdk/docs/install)
- GCP Project with billing enabled
- Required GCP APIs (we'll enable these in setup):
  - Firestore (Native mode)
  - Cloud Run
  - Cloud Scheduler
  - Cloud Build

### 1. GCP Setup

```bash
# Authenticate
gcloud auth login

# Set your project ID
export GCP_PROJECT_ID="your-project-id"
gcloud config set project $GCP_PROJECT_ID

# Enable required APIs (takes a few minutes)
gcloud services enable run.googleapis.com \
  cloudscheduler.googleapis.com \
  firestore.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com

# Create Firestore database (Native mode)
gcloud firestore databases create --location=us-central1
```

### 2. Local Development (Optional)

```bash
# Clone repository
git clone https://github.com/Lllllllleong/wazePoliceScraperGCP.git
cd wazePoliceScraperGCP

# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env and set your GCP_PROJECT_ID

# Authenticate for local development
gcloud auth application-default login

# Run locally
go run cmd/scraper/main.go

# Test it (in another terminal)
curl http://localhost:8080
```

### 3. Deploy to Cloud Run

```bash
# Windows PowerShell
$env:GCP_PROJECT_ID = "your-project-id"
.\scripts\deploy.bat

# Linux/Mac/WSL
export GCP_PROJECT_ID="your-project-id"
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

The deployment script will:
- Build a Docker container
- Push to Google Artifact Registry  
- Deploy to Cloud Run
- Configure environment variables
- Return the service URL

### 4. Setup Automated Scraping

```bash
# Get the Cloud Run URL
SERVICE_URL=$(gcloud run services describe waze-scraper \
  --region us-central1 \
  --format 'value(status.url)')

# Create scheduler job (runs every 2 minutes)
gcloud scheduler jobs create http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/2 * * * *" \
  --uri="$SERVICE_URL" \
  --http-method=GET \
  --oidc-service-account-email="${GCP_PROJECT_ID}@appspot.gserviceaccount.com"

# Test the scheduler
gcloud scheduler jobs run waze-scraper-job --location=us-central1
```

#### Common Schedules
- Every 2 minutes: `"*/2 * * * *"`
- Every 5 minutes: `"*/5 * * * *"`
- Every hour: `"0 * * * *"`
- Daily at 8am: `"0 8 * * *"`

## 📊 Export Data

### Export alerts from a date range

```bash
# Set your project (if not already set)
export GCP_PROJECT_ID="your-project-id"

# Export to JSONL (one alert per line)
go run cmd/exporter/main.go \
  --start=2025-10-03 \
  --end=2025-10-05 \
  --output=alerts.jsonl \
  --format=jsonl

# Export to JSON (single array)
go run cmd/exporter/main.go \
  --start=2025-10-03 \
  --end=2025-10-05 \
  --output=alerts.json \
  --format=json
```

### Export Options

- `--project`: GCP Project ID (or use `GCP_PROJECT_ID` env var)
- `--collection`: Firestore collection name (default: `police_alerts`)
- `--start`: Start date in YYYY-MM-DD format
- `--end`: End date in YYYY-MM-DD format
- `--output`: Output file path
- `--format`: Output format (`json` or `jsonl`)

### Example Exports

```bash
# Last week's data
go run cmd/exporter/main.go \
  --start=2025-09-26 \
  --end=2025-10-03 \
  --output=last_week.jsonl

# Specific day
go run cmd/exporter/main.go \
  --start=2025-10-03 \
  --end=2025-10-03 \
  --output=today.json \
  --format=json

# Custom collection
go run cmd/exporter/main.go \
  --collection=test_alerts \
  --start=2025-10-01 \
  --end=2025-10-03 \
  --output=test.jsonl
```

## 🗺️ Configure Geographic Coverage

The scraper is pre-configured for the **Sydney to Canberra corridor (Hume Highway)** with 4 overlapping regions for complete coverage.

### Default Coverage Areas

1. **Hume Highway - Sydney Section**: Northern section near Sydney
2. **Hume Highway - Middle Section**: Mid-corridor
3. **Hume Highway - Canberra Section**: Southern approach to Canberra
4. **Canberra Metropolitan Area**: ACT region

These default bounding boxes are defined in `cmd/scraper/main.go`.

### Override for Different Regions

Set the `WAZE_BBOXES` environment variable with semicolon-separated coordinates:

```bash
# Single region
export WAZE_BBOXES="103.6,1.15,104.0,1.45"

# Multiple regions (semicolon-separated)
export WAZE_BBOXES="103.6,1.15,104.0,1.45;1.2,103.6,1.5,104.0"

# Then deploy
./scripts/deploy.sh
```

### Coordinate Format

`west,south,east,north` = `longitude_min,latitude_min,longitude_max,latitude_max`

### Finding Coordinates

1. Visit [BBox Finder](https://boundingbox.klokantech.com/)
2. Draw a rectangle around your area of interest
3. Select **"CSV"** format from dropdown
4. Copy the coordinates (west,south,east,north)

### Example Regions

```bash
# Singapore
WAZE_BBOXES="103.6,1.15,104.0,1.45"

# New York City
WAZE_BBOXES="-74.05,40.68,-73.90,40.80"

# Multiple cities
WAZE_BBOXES="103.6,1.15,104.0,1.45;-74.05,40.68,-73.90,40.80"

# Melbourne to Sydney
WAZE_BBOXES="144.8,-37.9,145.1,-37.7;150.9,-34.0,151.3,-33.7"
```

### Update Default Regions Permanently

Edit `cmd/scraper/main.go`:

```go
defaultBBoxes = []string{
    "your,bbox,coords,here",
    "another,region,here",
}
```

See `configs/bboxes.yaml` for the current default configuration reference.

## 🔍 Monitoring & Troubleshooting

### View Logs

```bash
# View recent Cloud Run logs
gcloud run services logs read waze-scraper \
  --region=us-central1 \
  --limit=50

# Stream logs in real-time
gcloud run services logs tail waze-scraper \
  --region=us-central1

# View Cloud Scheduler logs
gcloud scheduler jobs describe waze-scraper-job \
  --location=us-central1
```

### Check Firestore Data

1. Go to [Cloud Console → Firestore](https://console.cloud.google.com/firestore)
2. Look for the `police_alerts` collection
3. View individual alert documents with lifecycle tracking fields

### Test Scraper Manually

```bash
# Get your Cloud Run URL
SERVICE_URL=$(gcloud run services describe waze-scraper \
  --region us-central1 \
  --format 'value(status.url)')

# Trigger scraping
curl $SERVICE_URL

# Should return JSON response:
# {
#   "status": "success",
#   "alerts_found": 123,
#   "police_alerts_saved": 45,
#   "stats": {...},
#   "bboxes_used": 4
# }
```

### Common Issues

**"Failed to connect to Firestore"**
- Ensure Firestore database is created: `gcloud firestore databases create --location=us-central1`
- Check that the Firestore API is enabled

**"No alerts found"**
- Verify bounding boxes cover active areas (try a known busy region)
- Check Waze API is responding (logs will show fetch errors)

**Scheduler not triggering**
- Verify job exists: `gcloud scheduler jobs list --location=us-central1`
- Check job status and last run time
- Ensure service account has permissions

## 📦 Data Model

Each police alert in Firestore includes:

### Core Alert Data
- `uuid`: Unique identifier from Waze
- `type`: "POLICE"
- `subtype`: e.g., "POLICE_VISIBLE", "POLICE_HIDING"
- `location_geo`: GeoPoint (lat/lng) for geospatial queries
- `street`, `city`, `country`: Location details
- `reliability`, `confidence`: Quality metrics

### Lifecycle Tracking
- `publish_time`: When alert was first published on Waze
- `scrape_time`: First time our scraper detected the alert
- `expire_time`: Last time the alert was seen (updated each scrape)
- `verification_count`: Number of times alert was re-observed
- `is_active`: Boolean flag (true until alert disappears)

### Sample Document

```json
{
  "uuid": "abc123-def456",
  "type": "POLICE",
  "subtype": "POLICE_VISIBLE",
  "location_geo": {
    "latitude": -35.123,
    "longitude": 149.456
  },
  "street": "Hume Highway",
  "city": "Goulburn",
  "country": "AU",
  "reliability": 5,
  "confidence": 8,
  "publish_time": "2025-10-03T10:30:00Z",
  "scrape_time": "2025-10-03T10:32:00Z",
  "expire_time": "2025-10-03T11:45:00Z",
  "verification_count": 37,
  "is_active": false
}
```

This structure enables:
- Tracking how long police remain at locations
- Analyzing alert patterns over time
- Identifying frequently monitored locations
- Calculating alert reliability metrics

## 🛠️ Development

### Project Structure

```
wazePoliceScraperGCP/
├── cmd/
│   ├── scraper/main.go      # Cloud Run HTTP service
│   └── exporter/main.go     # Data export CLI tool
├── internal/
│   ├── models/alert.go      # Data structures
│   ├── waze/client.go       # Waze API client
│   ├── storage/
│   │   ├── firestore.go     # Firestore client
│   │   └── police_alerts.go # Police alert operations
├── configs/
│   └── bboxes.yaml          # Bounding box reference
├── scripts/
│   ├── deploy.sh            # Linux/Mac deployment
│   └── deploy.bat           # Windows deployment
├── Dockerfile               # Container definition
├── .env.example             # Configuration template
└── go.mod                   # Go dependencies
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GCP_PROJECT_ID` | ✅ Yes | - | Your GCP Project ID |
| `FIRESTORE_COLLECTION` | ❌ No | `police_alerts` | Firestore collection name |
| `PORT` | ❌ No | `8080` | HTTP server port |
| `WAZE_BBOXES` | ❌ No | Sydney-Canberra | Semicolon-separated bounding boxes |

### Making Changes

**Add new alert fields:**
1. Update `internal/models/alert.go`
2. Update `internal/storage/police_alerts.go` save logic

**Change scraping logic:**
1. Update `internal/waze/client.go` for API changes
2. Update `cmd/scraper/main.go` for processing changes

**Add new queries:**
1. Add methods to `internal/storage/police_alerts.go`
2. Update exporter in `cmd/exporter/main.go` if needed

### Testing Locally

```bash
# Set environment
export GCP_PROJECT_ID="your-project-id"
export WAZE_BBOXES="103.6,1.15,104.0,1.45"  # Optional

# Authenticate
gcloud auth application-default login

# Run scraper
go run cmd/scraper/main.go

# In another terminal, test
curl http://localhost:8080

# Check logs and Firestore to verify data
```

### Deployment Workflow

```bash
# 1. Make code changes
# 2. Test locally
# 3. Deploy
./scripts/deploy.sh

# 4. Verify
SERVICE_URL=$(gcloud run services describe waze-scraper \
  --region us-central1 --format 'value(status.url)')
curl $SERVICE_URL

# 5. Check logs
gcloud run services logs read waze-scraper --region=us-central1
```

## � Maintenance

### Update Scheduler Frequency

```bash
# Update to every 5 minutes
gcloud scheduler jobs update http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/5 * * * *"

# Update to hourly
gcloud scheduler jobs update http waze-scraper-job \
  --location=us-central1 \
  --schedule="0 * * * *"
```

### Clean Up Old Data

The scraper tracks alert lifecycle automatically. To clean up old data:

```bash
# Use Firestore console to delete old documents
# Or implement a cleanup function in the code
```

### Delete Resources

```bash
# Stop scheduler
gcloud scheduler jobs delete waze-scraper-job --location=us-central1

# Delete Cloud Run service
gcloud run services delete waze-scraper --region=us-central1

# Note: Firestore data persists. Delete manually if needed.
```

## 📚 Additional Documentation

- **Deployment Guide**: See `docs/DEPLOYMENT_GUIDE.md` for detailed step-by-step deployment
- **Quick Reference**: See `docs/QUICK_DEPLOY.md` for deployment checklist
- **Environment Variables**: See `docs/ENVIRONMENT_VARIABLES.md` for complete variable reference

## �🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Test locally
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📝 License

MIT License - See LICENSE file for details

## 🔗 Links

- [Google Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Firestore Documentation](https://cloud.google.com/firestore/docs)
- [Cloud Scheduler Documentation](https://cloud.google.com/scheduler/docs)
- [BBox Finder Tool](https://boundingbox.klokantech.com/)

---

**Note**: This project is for educational and research purposes. Ensure compliance with Waze's terms of service when using their API.
