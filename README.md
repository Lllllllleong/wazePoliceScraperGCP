# Waze Police Scraper GCP

A Google Cloud Platform application that scrapes Waze alert data and stores it in Firestore.

## 🏗️ Architecture

This project consists of three main components:

1. **Cloud Run Service** (`cmd/scraper`): HTTP endpoint that scrapes Waze API and stores alerts in Firestore
2. **Cloud Scheduler**: Triggers the scraper every 2 minutes (configurable)
3. **Exporter Tool** (`cmd/exporter`): CLI utility to export Firestore data to JSON/JSONL files

```
Cloud Scheduler (every 2 min)
         ↓
    Cloud Run Service
         ↓
    Waze API → Parse → Firestore
                           ↓
                    Exporter Tool → JSON/JSONL files
```

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
- Google Cloud SDK (`gcloud` CLI)
- GCP Project with:
  - Firestore enabled (Native mode)
  - Cloud Run API enabled
  - Cloud Scheduler API enabled

### 1. GCP Setup

```bash
# Set your project ID
export GCP_PROJECT_ID="your-project-id"
gcloud config set project $GCP_PROJECT_ID

# Enable required APIs
gcloud services enable run.googleapis.com
gcloud services enable cloudscheduler.googleapis.com
gcloud services enable firestore.googleapis.com

# Create Firestore database (if not exists)
gcloud firestore databases create --region=us-central
```

### 2. Local Development

```bash
# Clone and setup
git clone https://github.com/Lllllllleong/wazePoliceScraperGCP.git
cd wazePoliceScraperGCP

# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env with your project ID (bboxes are already configured for Sydney-Canberra)

# Run locally
export GCP_PROJECT_ID="your-project-id"
# Optional: Override bboxes
# export WAZE_BBOXES="150.388,-34.254,151.008,-33.937"
go run cmd/scraper/main.go

# Test it
curl http://localhost:8080
```

### 3. Deploy to Cloud Run

```bash
# Make deploy script executable
chmod +x scripts/deploy.sh

# Deploy
export GCP_PROJECT_ID="your-project-id"
./scripts/deploy.sh
```

### 4. Setup Cloud Scheduler

```bash
# Get the Cloud Run URL
SERVICE_URL=$(gcloud run services describe waze-scraper \
  --platform managed \
  --region us-central1 \
  --format 'value(status.url)')

# Create scheduler job (runs every 2 minutes)
gcloud scheduler jobs create http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/2 * * * *" \
  --uri="$SERVICE_URL" \
  --http-method=GET

# Manually trigger to test
gcloud scheduler jobs run waze-scraper-job --location=us-central1
```

See `configs/scheduler-setup.md` for more scheduling options.

## 📊 Export Data

### Export alerts from a date range:

```bash
go run cmd/exporter/main.go \
  --project=your-project-id \
  --start=2025-10-03 \
  --end=2025-10-05 \
  --output=alerts.jsonl \
  --format=jsonl
```

### Export all alerts:

```bash
go run cmd/exporter/main.go \
  --project=your-project-id \
  --all \
  --output=all_alerts.json \
  --format=json
```

### Options:

- `--project`: GCP Project ID (or set `GCP_PROJECT_ID` env var)
- `--start`: Start date (YYYY-MM-DD)
- `--end`: End date (YYYY-MM-DD)
- `--output`: Output file path
- `--format`: `json` (array) or `jsonl` (one per line)
- `--all`: Export all alerts (ignores date range)

## 🗺️ Configure Bounding Boxes

The scraper is pre-configured for **Sydney to Canberra (Hume Highway)** with 4 overlapping regions.

### Default Coverage

1. **Hume Highway - Sydney Section**
2. **Hume Highway - Middle Section**  
3. **Hume Highway - Canberra Section**
4. **Canberra Metropolitan Area**

These are defined in `cmd/scraper/main.go` as `defaultBBoxes`.

### Override Defaults

You can override via environment variable:

```bash
# Single area
export WAZE_BBOXES="150.388,-34.254,151.008,-33.937"

# Multiple areas (semicolon-separated)
export WAZE_BBOXES="150.388,-34.254,151.008,-33.937;149.589,-34.769,150.830,-34.138"
```

### Format
`west,south,east,north` (longitude, latitude coordinates)

### Find New Coordinates

1. Go to https://boundingbox.klokantech.com/
2. Draw a box around your area of interest
3. Select "CSV" format
4. Copy the coordinates

### Examples

```bash
# Singapore
export WAZE_BBOXES="103.6,1.15,104.0,1.45"

# New York City
export WAZE_BBOXES="-74.05,40.68,-73.90,40.80"

# Multiple cities
export WAZE_BBOXES="103.6,1.15,104.0,1.45;-74.05,40.68,-73.90,40.80"
```

### Update Defaults Permanently

Edit the `defaultBBoxes` variable in `cmd/scraper/main.go`:

```go
defaultBBoxes = []string{
    "your,bbox,coords,here",
    "another,bbox,here",
}
```

Or maintain them in `configs/bboxes.yaml` for documentation.

## 🔍 Monitoring

### View Cloud Run logs:

```bash
gcloud run services logs read waze-scraper \
  --region=us-central1 \
  --limit=50
```

### Check Firestore:

Go to [Cloud Console → Firestore](https://console.cloud.google.com/firestore) and view the `police_alerts` collection.

### Test scraper manually:

```bash
curl https://YOUR-CLOUD-RUN-URL
```

## 📦 Data Model

Each alert stored in Firestore has this structure:

```json
{
  "uuid": "abc123",
  "type": "POLICE",
  "subtype": "POLICE_VISIBLE",
  "location": {
    "latitude": 1.3521,
    "longitude": 103.8198
  },
  "pub_millis": 1696320000000,
  "reliability": 5,
  "street": "Main Street",
  "city": "Singapore",
  "country": "SG",
  "scraped_at": "2025-10-03T10:30:00Z"
}
```

## 🛠️ Development

### Project Layout (Go Standard)

- `cmd/`: Entry points (executables)
- `internal/`: Private application code
- `configs/`: Configuration files
- `scripts/`: Build and deployment scripts

### Adding New Features

1. **New alert types**: Update `internal/models/alert.go`
2. **New API endpoints**: Update `internal/waze/client.go`
3. **New Firestore queries**: Update `internal/storage/firestore.go`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📝 License

MIT
