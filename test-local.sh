#!/bin/bash

# Local Testing Script for Firebase Auth Implementation
# This script starts all necessary services for local testing

echo "🚀 Starting Local Test Environment..."
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Firebase emulator is running
echo "📋 Step 1: Starting Firebase Auth Emulator..."
echo "${YELLOW}Run this in a separate terminal:${NC}"
echo "  cd dataAnalysis && firebase emulators:start --only auth"
echo ""
read -p "Press Enter when Firebase emulator is running..."

# Start the Go backend
echo ""
echo "📋 Step 2: Starting alerts-service backend..."
echo "${YELLOW}Run this in a separate terminal:${NC}"
echo "  cd cmd/alerts-service && go run main.go"
echo ""
echo "  OR if you need to set environment variables:"
echo "  export GCP_PROJECT_ID=wazepolicescrapergcp"
echo "  export FIRESTORE_COLLECTION=police_alerts"
echo "  export GCS_BUCKET_NAME=police-alerts-archive-wazepolicescrapergcp"
echo "  export RATE_LIMIT_PER_MINUTE=30"
echo "  cd cmd/alerts-service && go run main.go"
echo ""
read -p "Press Enter when backend is running on port 8080..."

# Start the frontend
echo ""
echo "📋 Step 3: Starting frontend development server..."
echo "${YELLOW}Run this in a separate terminal:${NC}"
echo "  cd dataAnalysis/public && python -m http.server 3000"
echo ""
read -p "Press Enter when frontend is running on port 3000..."

# Instructions
echo ""
echo "${GREEN}✅ All services should now be running!${NC}"
echo ""
echo "🌐 Open your browser to: http://localhost:3000"
echo ""
echo "📊 Firebase Emulator UI: http://localhost:4000"
echo ""
echo "🔍 What to check:"
echo "  1. Open browser console (F12)"
echo "  2. Look for '🔥 Connecting to Firebase Auth Emulator'"
echo "  3. Select dates and click 'Load Data'"
echo "  4. Check for authentication logs"
echo "  5. Backend should log 'Authenticated user: <uid>'"
echo ""
echo "🛑 To stop: Press Ctrl+C in each terminal"
