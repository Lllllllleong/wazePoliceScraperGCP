@echo off
REM Deployment script for Waze Scraper Cloud Run service (Windows)

REM Configuration
set SERVICE_NAME=scraper-service
set REGION=us-central1

if "%FIRESTORE_COLLECTION%"=="" (
    set FIRESTORE_COLLECTION=police_alerts
)

if "%GCP_PROJECT_ID%"=="" (
    echo Error: GCP_PROJECT_ID environment variable is required
    exit /b 1
)

echo.
echo 🚀 Deploying Waze Scraper to Cloud Run...
echo Project: %GCP_PROJECT_ID%
echo Collection: %FIRESTORE_COLLECTION%
echo Service: %SERVICE_NAME%
echo Region: %REGION%
echo.

REM Set the project
gcloud config set project %GCP_PROJECT_ID%

REM Deploy to Cloud Run
gcloud run deploy %SERVICE_NAME% ^
  --source . ^
  --platform managed ^
  --region %REGION% ^
  --allow-unauthenticated ^
  --set-env-vars GCP_PROJECT_ID=%GCP_PROJECT_ID%,FIRESTORE_COLLECTION=%FIRESTORE_COLLECTION%

if %ERRORLEVEL% neq 0 (
    echo.
    echo ❌ Deployment failed!
    exit /b 1
)

echo.
echo ✅ Deployment complete!
echo.

REM Get the service URL
for /f "delims=" %%i in ('gcloud run services describe %SERVICE_NAME% --platform managed --region %REGION% --format "value(status.url)"') do set SERVICE_URL=%%i

echo Service URL: %SERVICE_URL%
echo.
echo To test the scraper:
echo curl %SERVICE_URL%
echo.
