#!/bin/bash
# Manual deployment script for the Scraper Service

set -e # Exit immediately if a command exits with a non-zero status. 

# --- Configuration ---
# The GCP_PROJECT_ID is fetched from your gcloud config.
# Ensure you have run 'gcloud config set project YOUR_PROJECT_ID'
export GCP_PROJECT_ID=$(gcloud config get-value project)
export GAR_LOCATION="us-central1"
export SERVICE_NAME="scraper-service"
export REGION="us-central1"
export DOCKERFILE="Dockerfile.scraper"
export IMAGE_TAG="${GAR_LOCATION}-docker.pkg.dev/${GCP_PROJECT_ID}/${SERVICE_NAME}/${SERVICE_NAME}:latest"

# --- Environment Variables for the Service ---
# These are the variables that will be set on the Cloud Run service itself.
export FIRESTORE_COLLECTION="police_alerts"

echo "--- Starting Deployment for ${SERVICE_NAME} ---"
echo "Project: ${GCP_PROJECT_ID}"
echo "Region: ${REGION}"
echo "Image: ${IMAGE_TAG}"
echo "-------------------------------------------"

# 1. Build the Docker image
echo "STEP 1: Building Docker image..."
docker build -t "${IMAGE_TAG}" -f "${DOCKERFILE}" .

# 2. Push the image to Google Artifact Registry
# Ensure you have authenticated Docker with 'gcloud auth configure-docker ${GAR_LOCATION}-docker.pkg.dev'
echo "STEP 2: Pushing image to Artifact Registry..."
docker push "${IMAGE_TAG}"

# 3. Deploy to Cloud Run
echo "STEP 3: Deploying to Cloud Run..."
gcloud run deploy "${SERVICE_NAME}" \
  --image "${IMAGE_TAG}" \
  --region "${REGION}" \
  --platform "managed" \
  --allow-unauthenticated \
  --max-instances "1" \
  --min-instances "0" \
  --memory "512Mi" \
  --cpu "1" \
  --set-env-vars "GCP_PROJECT_ID=${GCP_PROJECT_ID},FIRESTORE_COLLECTION=${FIRESTORE_COLLECTION}"

echo "--- ✅ Deployment for ${SERVICE_NAME} Complete ---"
