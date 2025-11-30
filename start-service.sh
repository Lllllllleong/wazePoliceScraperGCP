#!/bin/bash
# Start Alerts Service for Local Testing
# This script helps start the service with required environment variables

echo "Starting Alerts Service for GZIP Testing..."
echo ""

# Check if environment variables are set
if [ -z "$GCP_PROJECT_ID" ]; then
    echo "ERROR: GCP_PROJECT_ID is not set"
    echo "Please set it with: export GCP_PROJECT_ID=your-project-id"
    echo ""
    echo "To find your project ID:"
    echo "  gcloud config get-value project"
    exit 1
fi

if [ -z "$GCS_BUCKET_NAME" ]; then
    echo "ERROR: GCS_BUCKET_NAME is not set"
    echo "Please set it with: export GCS_BUCKET_NAME=your-bucket-name"
    exit 1
fi

# Optional variables with defaults
export PORT="${PORT:-8080}"
export FIRESTORE_COLLECTION="${FIRESTORE_COLLECTION:-police_alerts}"
export RATE_LIMIT_PER_MINUTE="${RATE_LIMIT_PER_MINUTE:-30}"

echo "Configuration:"
echo "  GCP_PROJECT_ID: $GCP_PROJECT_ID"
echo "  GCS_BUCKET_NAME: $GCS_BUCKET_NAME"
echo "  PORT: $PORT"
echo "  FIRESTORE_COLLECTION: $FIRESTORE_COLLECTION"
echo "  RATE_LIMIT_PER_MINUTE: $RATE_LIMIT_PER_MINUTE"
echo ""
echo "Starting service..."
echo "Press Ctrl+C to stop"
echo ""

cd /c/Users/Leong/wazePoliceScraperGCP
./alerts-service.exe
