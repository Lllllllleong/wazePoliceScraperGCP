# Quick Deployment Checklist

Use this as a quick reference while deploying. Full details in `DEPLOYMENT_GUIDE.md`.

## ✅ Pre-Deployment (One-Time Setup)

```bash
# 1. Login to Google Cloud
gcloud auth login

# 2. Set your project ID
export GCP_PROJECT_ID=wazepolicescrapergcp
gcloud config set project $GCP_PROJECT_ID

# 3. Enable APIs (copy-paste all at once)
gcloud services enable run.googleapis.com && \
gcloud services enable cloudscheduler.googleapis.com && \
gcloud services enable firestore.googleapis.com && \
gcloud services enable cloudbuild.googleapis.com && \
gcloud services enable artifactregistry.googleapis.com

# 4. Create Firestore database
gcloud firestore databases create --location=us-central1

# 5. Verify .env file
# Should contain:
#   GCP_PROJECT_ID=wazepolicescrapergcp
#   FIRESTORE_COLLECTION=police_alerts
```

## 🚀 Deploy

```bash
# Windows PowerShell
$env:GCP_PROJECT_ID = "wazepolicescrapergcp"
.\scripts\deploy.bat

# OR Bash/Linux/Mac
export GCP_PROJECT_ID=wazepolicescrapergcp
./scripts/deploy.sh
```

## 🧪 Test

```bash
# Get the URL
SERVICE_URL=$(gcloud run services describe waze-scraper --platform managed --region us-central1 --format 'value(status.url)')

# Test it
curl $SERVICE_URL

# Check Firestore
# Go to: https://console.cloud.google.com/firestore
# Look for "police_alerts" collection
```

## ⏰ Schedule (Optional)

```bash
# Get service URL
SERVICE_URL=$(gcloud run services describe waze-scraper --platform managed --region us-central1 --format 'value(status.url)')

# Create scheduler (every 2 minutes)
gcloud scheduler jobs create http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/2 * * * *" \
  --uri="$SERVICE_URL" \
  --http-method=GET \
  --oidc-service-account-email="${GCP_PROJECT_ID}@appspot.gserviceaccount.com"

# Test it
gcloud scheduler jobs run waze-scraper-job --location=us-central1
```

## 📊 Monitor

```bash
# View logs
gcloud run services logs read waze-scraper --region=us-central1 --limit=50

# Or real-time
gcloud run services logs tail waze-scraper --region=us-central1
```

## 🔄 Update

```bash
# After code changes, just redeploy
./scripts/deploy.sh
```

## 🗑️ Cleanup

```bash
# Delete service
gcloud run services delete waze-scraper --region=us-central1

# Delete scheduler
gcloud scheduler jobs delete waze-scraper-job --location=us-central1
```

## 🆘 Troubleshooting

**Build fails?** Check Dockerfile exists and Docker is valid

**No data?** Check logs: `gcloud run services logs read waze-scraper --region=us-central1`

**Permission errors?** Make sure billing is enabled

**API not enabled?** Run the enable commands again
