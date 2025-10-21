#!/bin/bash

# Deployment script for Waze Police Alerts API Cloud Run service

set -e

# Load .env file if it exists
if [ -f .env ]; then
  echo "📋 Loading environment variables from .env..."
  export $(cat .env | grep -v '^#' | xargs)
fi

# Configuration
PROJECT_ID="${GCP_PROJECT_ID:-your-project-id}"
COLLECTION_NAME="${FIRESTORE_COLLECTION:-police_alerts}"
SERVICE_NAME="waze-alerts-api"
REGION="us-central1"

echo "🚀 Deploying Waze Alerts API to Cloud Run..."
echo "Project: $PROJECT_ID"
echo "Collection: $COLLECTION_NAME"
echo "Service: $SERVICE_NAME"
echo "Region: $REGION"

# Set the project
gcloud config set project $PROJECT_ID

# Build with Docker and deploy to Cloud Run
echo "🔨 Building Docker image with Cloud Build..."
gcloud builds submit --config cloudbuild-api.yaml .

echo "🚀 Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
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
echo "📝 Update dataAnalysis/public/config.js with:"
echo "   alertsEndpoint: \"$SERVICE_URL/api/alerts\""
echo ""
echo "To test the API:"
echo "curl -X POST $SERVICE_URL/api/alerts \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"dates\": [\"2025-10-20\"]}'"
echo ""
