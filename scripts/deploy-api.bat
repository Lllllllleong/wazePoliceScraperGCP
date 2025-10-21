@echo off
REM Deployment script for Waze Police Alerts API Cloud Run service

REM Load .env file if it exists
if exist .env (
    echo 📋 Loading environment variables from .env...
    for /f "tokens=*" %%a in ('type .env ^| findstr /v "^#"') do set %%a
)

REM Configuration
if "%GCP_PROJECT_ID%"=="" set GCP_PROJECT_ID=your-project-id
if "%FIRESTORE_COLLECTION%"=="" set FIRESTORE_COLLECTION=police_alerts
set SERVICE_NAME=waze-alerts-api
set REGION=us-central1

echo 🚀 Deploying Waze Alerts API to Cloud Run...
echo Project: %GCP_PROJECT_ID%
echo Collection: %FIRESTORE_COLLECTION%
echo Service: %SERVICE_NAME%
echo Region: %REGION%

REM Set the project
gcloud config set project %GCP_PROJECT_ID%

REM Build with Docker and deploy to Cloud Run
echo 🔨 Building Docker image...
gcloud builds submit --tag gcr.io/%GCP_PROJECT_ID%/%SERVICE_NAME% --file Dockerfile.api .

echo 🚀 Deploying to Cloud Run...
gcloud run deploy %SERVICE_NAME% ^
  --image gcr.io/%GCP_PROJECT_ID%/%SERVICE_NAME% ^
  --platform managed ^
  --region %REGION% ^
  --allow-unauthenticated ^
  --set-env-vars GCP_PROJECT_ID=%GCP_PROJECT_ID%,FIRESTORE_COLLECTION=%FIRESTORE_COLLECTION%

echo ✅ Deployment complete!

REM Get the service URL
for /f "tokens=*" %%a in ('gcloud run services describe %SERVICE_NAME% --platform managed --region %REGION% --format "value(status.url)"') do set SERVICE_URL=%%a

echo.
echo Service URL: %SERVICE_URL%
echo.
echo 📝 Update dataAnalysis/public/config.js with:
echo    alertsEndpoint: "%SERVICE_URL%/api/alerts"
echo.
echo To test the API:
echo curl -X POST %SERVICE_URL%/api/alerts ^
echo   -H "Content-Type: application/json" ^
echo   -d "{\"dates\": [\"2025-10-20\"]}"
echo.

pause
