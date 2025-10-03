# Complete GCP Deployment Guide - From Scratch

This guide walks you through deploying the Waze Police Scraper to Google Cloud Platform, starting from a fresh GCP project with nothing set up yet.

## 📋 Prerequisites

### What You Need
- ✅ GCP Project created (you have this!)
- ✅ GCP Project ID (e.g., `wazepolicescrapergcp`)
- ✅ Google Cloud SDK (`gcloud` CLI) installed on your computer
- ✅ Billing enabled on your GCP project

### Check if `gcloud` is installed
```bash
gcloud --version
```

If not installed, download from: https://cloud.google.com/sdk/docs/install

---

## 🚀 Step-by-Step Deployment

### **Step 1: Authenticate with Google Cloud**

Open your terminal and authenticate:

```bash
# Login to Google Cloud
gcloud auth login

# This will open a browser window - sign in with your Google account
```

After signing in, you should see a success message.

---

### **Step 2: Set Your Project**

```bash
# Replace with your actual project ID
export GCP_PROJECT_ID=wazepolicescrapergcp

# Set as default project
gcloud config set project $GCP_PROJECT_ID

# Verify it's set correctly
gcloud config get-value project
```

**Expected output:** `wazepolicescrapergcp` (or your project ID)

---

### **Step 3: Enable Required APIs**

These commands enable all the services you need. This takes a few minutes.

```bash
# Enable Cloud Run (for running the scraper)
gcloud services enable run.googleapis.com

# Enable Cloud Scheduler (for automatic triggering)
gcloud services enable cloudscheduler.googleapis.com

# Enable Firestore (for storing data)
gcloud services enable firestore.googleapis.com

# Enable Cloud Build (for building the container)
gcloud services enable cloudbuild.googleapis.com

# Enable Artifact Registry (for storing container images)
gcloud services enable artifactregistry.googleapis.com
```

**Expected output:** Each command will show `Operation "operations/..." finished successfully.`

💡 **Note:** If you get a billing error, make sure billing is enabled on your project at: https://console.cloud.google.com/billing

---

### **Step 4: Create Firestore Database**

You need to create a Firestore database in Native mode:

```bash
# Create Firestore database in us-central1 region
gcloud firestore databases create --location=us-central1
```

**Expected output:** `Success! Selected Google Cloud Firestore Native database for wazepolicescrapergcp`

💡 **Note:** 
- You can only have ONE Firestore database per project
- `us-central1` is recommended for low latency and cost
- The collection (`police_alerts`) will be created automatically when the scraper runs

**Alternative:** Create via Console
1. Go to: https://console.cloud.google.com/firestore
2. Click "Select Native Mode"
3. Choose location: `us-central1`
4. Click "Create Database"

---

### **Step 5: Configure Your Local Environment**

Navigate to your project directory:

```bash
cd C:\Users\Leong\wazePoliceScraperGCP
```

Edit your `.env` file (already created):

```bash
# Open .env in your editor
notepad .env
```

Make sure it has:
```bash
GCP_PROJECT_ID=wazepolicescrapergcp
FIRESTORE_COLLECTION=police_alerts
PORT=8080
WAZE_BBOXES=
```

Save and close.

---

### **Step 6: Deploy to Cloud Run**

Now deploy your scraper:

**Option A: Using PowerShell (Windows)**
```powershell
# Set environment variables
$env:GCP_PROJECT_ID = "wazepolicescrapergcp"
$env:FIRESTORE_COLLECTION = "police_alerts"

# Run the deployment script
.\scripts\deploy.bat
```

**Option B: Using Bash**
```bash
# Set environment variables
export GCP_PROJECT_ID=wazepolicescrapergcp
export FIRESTORE_COLLECTION=police_alerts

# Make script executable
chmod +x scripts/deploy.sh

# Run deployment
./scripts/deploy.sh
```

**What happens during deployment:**
1. ✅ Uploads your code to Google Cloud
2. ✅ Builds a Docker container
3. ✅ Deploys to Cloud Run
4. ✅ Makes it publicly accessible

**Expected output:**
```
🚀 Deploying Waze Scraper to Cloud Run...
Project: wazepolicescrapergcp
Collection: police_alerts
Service: waze-scraper
Region: us-central1

Building using Dockerfile and deploying container to Cloud Run service [waze-scraper]
✓ Deploying... Done.
  ✓ Creating Revision...
  ✓ Routing traffic...
Done.

✅ Deployment complete!

Service URL: https://waze-scraper-xxxxx-uc.a.run.app

To test the scraper:
curl https://waze-scraper-xxxxx-uc.a.run.app
```

⏱️ **Time:** First deployment takes 3-5 minutes. Save the Service URL!

---

### **Step 7: Test Your Deployment**

Test the scraper manually:

```bash
# Replace with your actual service URL from Step 6
curl https://waze-scraper-xxxxx-uc.a.run.app
```

**Expected output:**
```json
{
  "status": "success",
  "alerts_found": 15,
  "police_alerts_saved": 8,
  "stats": {
    "total_requests": 4,
    "successful_calls": 4,
    "failed_calls": 0,
    "total_alerts": 15,
    "unique_alerts": 15
  },
  "bboxes_used": 4
}
```

🎉 **Success!** Your scraper is working and saving data to Firestore!

---

### **Step 8: Verify Data in Firestore**

Check that data is being saved:

**Option A: Via Console**
1. Go to: https://console.cloud.google.com/firestore
2. You should see a collection called `police_alerts`
3. Click on it to see the saved alerts

**Option B: Via Command Line**
```bash
# Install gcloud firestore
gcloud components install cloud-firestore-emulator

# Query the collection
gcloud firestore documents list police_alerts --limit=5
```

---

### **Step 9: Set Up Automatic Scheduling (Optional but Recommended)**

To automatically scrape every 2 minutes:

```bash
# Get your Cloud Run service URL
SERVICE_URL=$(gcloud run services describe waze-scraper \
  --platform managed \
  --region us-central1 \
  --format 'value(status.url)')

# Create Cloud Scheduler job
gcloud scheduler jobs create http waze-scraper-job \
  --location=us-central1 \
  --schedule="*/2 * * * *" \
  --uri="$SERVICE_URL" \
  --http-method=GET \
  --oidc-service-account-email="${GCP_PROJECT_ID}@appspot.gserviceaccount.com"
```

**Expected output:**
```
Created job [waze-scraper-job].
```

**Test the scheduler:**
```bash
gcloud scheduler jobs run waze-scraper-job --location=us-central1
```

**Check scheduler status:**
```bash
gcloud scheduler jobs describe waze-scraper-job --location=us-central1
```

---

### **Step 10: Monitor Your Scraper**

#### View Logs
```bash
# View recent logs
gcloud run services logs read waze-scraper \
  --region=us-central1 \
  --limit=50

# Follow logs in real-time
gcloud run services logs tail waze-scraper \
  --region=us-central1
```

#### View in Console
- Logs: https://console.cloud.google.com/run
- Firestore: https://console.cloud.google.com/firestore
- Scheduler: https://console.cloud.google.com/cloudscheduler

---

## 📊 Export Data

Once you have data in Firestore, export it:

```bash
# Export alerts from a specific date range
go run cmd/exporter/main.go \
  --project=wazepolicescrapergcp \
  --start=2025-10-03 \
  --end=2025-10-05 \
  --output=police_alerts.jsonl
```

---

## 🔧 Common Issues & Solutions

### Issue 1: "API not enabled"
**Solution:** Run the enable commands from Step 3 again

### Issue 2: "Permission denied"
**Solution:** Make sure billing is enabled and you're the project owner

### Issue 3: "Firestore already exists in Datastore mode"
**Solution:** You can't convert Datastore to Firestore Native. Create a new project.

### Issue 4: "Container build failed"
**Solution:** Check that `Dockerfile` exists in your project root and is valid

### Issue 5: "No data in Firestore"
**Solution:** 
- Check Cloud Run logs for errors
- Verify the bounding boxes in WAZE_BBOXES are correct
- Test the API manually with curl

### Issue 6: "Cloud Scheduler not triggering"
**Solution:**
- Check that the service account has permissions
- Verify the URL is correct
- Check scheduler job status with `gcloud scheduler jobs describe`

---

## 💰 Cost Estimate

With default settings (scraping every 2 minutes):
- **Cloud Run**: ~$5-10/month (mostly free tier)
- **Firestore**: ~$1-5/month (depends on data volume)
- **Cloud Scheduler**: $0.10/month (first 3 jobs free)

**Total: ~$6-15/month** (mostly covered by Google Cloud Free Tier)

💡 **Tip:** Set up billing alerts at https://console.cloud.google.com/billing/budgets

---

## 🎯 Next Steps

After deployment:
1. ✅ Monitor logs for a few hours to ensure stability
2. ✅ Adjust `WAZE_BBOXES` in `.env` if needed
3. ✅ Adjust scheduler frequency (every 2 min vs every 5 min, etc.)
4. ✅ Set up billing alerts
5. ✅ Export data periodically for backups

---

## 📞 Quick Reference Commands

### Redeploy after code changes
```bash
./scripts/deploy.sh
```

### Update environment variables
```bash
gcloud run services update waze-scraper \
  --region=us-central1 \
  --set-env-vars FIRESTORE_COLLECTION=new_collection_name
```

### View service details
```bash
gcloud run services describe waze-scraper --region=us-central1
```

### Delete everything (cleanup)
```bash
# Delete Cloud Run service
gcloud run services delete waze-scraper --region=us-central1

# Delete scheduler job
gcloud scheduler jobs delete waze-scraper-job --location=us-central1

# Note: Firestore data must be deleted manually via console
```

---

## ✅ Deployment Checklist

- [ ] `gcloud` CLI installed
- [ ] Authenticated (`gcloud auth login`)
- [ ] Project ID set
- [ ] APIs enabled (Run, Scheduler, Firestore, Build)
- [ ] Firestore database created (Native mode, us-central1)
- [ ] `.env` file configured
- [ ] Deployed to Cloud Run
- [ ] Tested with curl
- [ ] Data visible in Firestore
- [ ] Cloud Scheduler configured (optional)
- [ ] Monitoring/logs reviewed

---

🎉 **Congratulations!** Your Waze Police Scraper is now running on Google Cloud Platform!

For questions or issues, check:
- Cloud Run logs: `gcloud run services logs read waze-scraper --region=us-central1`
- Firestore console: https://console.cloud.google.com/firestore
- This project's README: `README.md`
