# Quick Deployment Checklist

> **Note**: This is a quick reference guide. For detailed explanations, see [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) or the main [README.md](../README.md).

## ✅ One-Time Setup

```bash
# 1. Authenticate
gcloud auth login

# 2. Set project
export GCP_PROJECT_ID="your-project-id"
gcloud config set project $GCP_PROJECT_ID

# 3. Enable APIs (one command)
gcloud services enable run.googleapis.com \
  cloudscheduler.googleapis.com \
  firestore.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com

# 4. Create Firestore database
gcloud firestore databases create --location=us-central1
```

## 🚀 Deploy Service

```bash
# Set project ID
export GCP_PROJECT_ID="your-project-id"

# Deploy (choose your platform)
# Windows PowerShell:
$env:GCP_PROJECT_ID = "your-project-id"
.\scripts\deploy.bat

# Linux/Mac/WSL:
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

## ⏰ Setup Scheduler

```bash
# Get service URL
SERVICE_URL=$(gcloud run services describe waze-scraper \
  --region us-central1 \
  --format 'value(status.url)')

# Create scheduler (every 2 minutes)
gcloud scheduler jobs create http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/2 * * * *" \
  --uri="$SERVICE_URL" \
  --http-method=GET \
  --oidc-service-account-email="${GCP_PROJECT_ID}@appspot.gserviceaccount.com"

# Test it immediately
gcloud scheduler jobs run waze-scraper-job --location=us-central1
```

### Alternative Schedules
- Every 5 minutes: `--schedule="*/5 * * * *"`
- Every hour: `--schedule="0 * * * *"`
- Daily at 8am: `--schedule="0 8 * * *"`

## 🧪 Verify Deployment

```bash
# Test service directly
curl $SERVICE_URL

# Check logs
gcloud run services logs read waze-scraper \
  --region=us-central1 \
  --limit=20

# View Firestore data
# https://console.cloud.google.com/firestore

# Check scheduler
gcloud scheduler jobs list --location=us-central1
```

## 🔄 Update Deployment

```bash
# After making code changes
./scripts/deploy.sh  # or deploy.bat on Windows

# Scheduler continues to work automatically
```

## 🗑️ Clean Up

```bash
# Delete scheduler
gcloud scheduler jobs delete waze-scraper-job --location=us-central1

# Delete Cloud Run service
gcloud run services delete waze-scraper --region=us-central1

# Note: Firestore data remains. Delete manually if needed.
```

## �️ Configure Different Regions

```bash
# Set custom bounding boxes before deploying
export WAZE_BBOXES="103.6,1.15,104.0,1.45"  # Singapore
./scripts/deploy.sh

# Or edit cmd/scraper/main.go defaultBBoxes for permanent change
```

## � Common Commands

```bash
# View real-time logs
gcloud run services logs tail waze-scraper --region=us-central1

# Manually trigger scraper
gcloud scheduler jobs run waze-scraper-job --location=us-central1

# Export data
go run cmd/exporter/main.go \
  --start=2025-10-03 \
  --end=2025-10-05 \
  --output=alerts.jsonl

# Update scheduler frequency
gcloud scheduler jobs update http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/5 * * * *"
```

---

For more details, see:
- [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) - Complete step-by-step guide
- [ENVIRONMENT_VARIABLES.md](./ENVIRONMENT_VARIABLES.md) - Configuration reference
- [README.md](../README.md) - Project overview and usage

## 🆘 Troubleshooting

**Build fails?** Check Dockerfile exists and Docker is valid

**No data?** Check logs: `gcloud run services logs read waze-scraper --region=us-central1`

**Permission errors?** Make sure billing is enabled

**API not enabled?** Run the enable commands again
