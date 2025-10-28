@echo off
REM Manual deployment script for the Alerts Service

setlocal

REM --- Configuration ---
REM The GCP_PROJECT_ID is fetched from your gcloud config.
REM Ensure you have run 'gcloud config set project YOUR_PROJECT_ID'
for /f "tokens=*" %%i in ('gcloud config get-value project') do set GCP_PROJECT_ID=%%i
set GAR_LOCATION=us-central1
set SERVICE_NAME=alerts-service
set REGION=us-central1
set DOCKERFILE=Dockerfile.alerts
set IMAGE_TAG=%GAR_LOCATION%-docker.pkg.dev/%GCP_PROJECT_ID%/%SERVICE_NAME%/%SERVICE_NAME%:latest

REM --- Environment Variables for the Service ---
REM These are the variables that will be set on the Cloud Run service itself.
set FIRESTORE_COLLECTION=police_alerts
set GCS_BUCKET_NAME=wazepolicescrapergcp-archive

echo --- Starting Deployment for %SERVICE_NAME% ---
echo Project: %GCP_PROJECT_ID%
echo Region: %REGION%
echo Image: %IMAGE_TAG%
echo -------------------------------------------

REM 1. Build the Docker image
echo STEP 1: Building Docker image...
docker build -t "%IMAGE_TAG%" -f "%DOCKERFILE%" .
if %errorlevel% neq 0 (
    echo Docker build failed.
    exit /b %errorlevel%
)

REM 2. Push the image to Google Artifact Registry
REM Ensure you have authenticated Docker with 'gcloud auth configure-docker %GAR_LOCATION%-docker.pkg.dev'
echo STEP 2: Pushing image to Artifact Registry...
docker push "%IMAGE_TAG%"
if %errorlevel% neq 0 (
    echo Docker push failed.
    exit /b %errorlevel%
)

REM 3. Deploy to Cloud Run
echo STEP 3: Deploying to Cloud Run...
gcloud run deploy "%SERVICE_NAME%" ^
  --image "%IMAGE_TAG%" ^
  --region "%REGION%" ^
  --platform "managed" ^
  --allow-unauthenticated ^
  --max-instances "1" ^
  --min-instances "0" ^
  --memory "512Mi" ^
  --cpu "1" ^
  --set-env-vars "GCP_PROJECT_ID=%GCP_PROJECT_ID%,FIRESTORE_COLLECTION=%FIRESTORE_COLLECTION%,GCS_BUCKET_NAME=%GCS_BUCKET_NAME%"

echo --- ✅ Deployment for %SERVICE_NAME% Complete ---
endlocal
