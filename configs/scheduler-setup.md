# Waze Police Scraper - Cloud Scheduler Setup

This directory contains documentation for setting up Cloud Scheduler to trigger the scraper.

## Setup Cloud Scheduler

### 1. Deploy the Cloud Run service first

```bash
./scripts/deploy.sh
```

### 2. Get the Cloud Run service URL

```bash
gcloud run services describe waze-scraper \
  --platform managed \
  --region us-central1 \
  --format 'value(status.url)'
```

### 3. Create a Cloud Scheduler job

**Every 2 minutes:**

```bash
gcloud scheduler jobs create http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/2 * * * *" \
  --uri="https://YOUR-CLOUD-RUN-URL" \
  --http-method=GET \
  --oidc-service-account-email=YOUR-SERVICE-ACCOUNT@PROJECT-ID.iam.gserviceaccount.com
```

Replace:
- `YOUR-CLOUD-RUN-URL` with the URL from step 2
- `YOUR-SERVICE-ACCOUNT` with your service account
- `PROJECT-ID` with your GCP project ID

### Alternative Schedules

**Every 5 minutes:**
```bash
--schedule="*/5 * * * *"
```

**Every hour:**
```bash
--schedule="0 * * * *"
```

**Every day at 8am:**
```bash
--schedule="0 8 * * *"
```

## Verify

```bash
# List scheduler jobs
gcloud scheduler jobs list --location=us-central1

# Manually trigger the job
gcloud scheduler jobs run waze-scraper-job --location=us-central1

# View logs
gcloud scheduler jobs describe waze-scraper-job --location=us-central1
```

## Monitoring

View Cloud Run logs to see scraping activity:

```bash
gcloud run services logs read waze-scraper \
  --region=us-central1 \
  --limit=50
```

Or use the Cloud Console:
https://console.cloud.google.com/run
