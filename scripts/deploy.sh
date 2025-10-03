#!/bin/bash

# Deployment script for Waze Scraper Cloud Run service

set -e

# Configuration
PROJECT_ID="${GCP_PROJECT_ID:-your-project-id}"
COLLECTION_NAME="${FIRESTORE_COLLECTION:-police_alerts}"
SERVICE_NAME="waze-scraper"
REGION="us-central1"

echo "🚀 Deploying Waze Scraper to Cloud Run..."
echo "Project: $PROJECT_ID"
echo "Collection: $COLLECTION_NAME"
echo "Service: $SERVICE_NAME"
echo "Region: $REGION"

# Set the project
gcloud config set project $PROJECT_ID

# Deploy to Cloud Run
gcloud run deploy $SERVICE_NAME \
  --source . \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars GCP_PROJECT_ID=$PROJECT_ID,FIRESTORE_COLLECTION=$COLLECTION_NAME

echo "✅ Deployment complete!"

# Get the service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --format 'value(status.url)')

echo ""
echo "Service URL: $SERVICE_URL"
echo ""
echo "To test the scraper:"
echo "curl $SERVICE_URL"
