# Police Alert Analysis System

[![CI/CD Status](https://img.shields.io/badge/CI%2FCD-Passing-brightgreen)](https://github.com/Lllllllleong/wazePoliceScraperGCP/actions)
[![codecov](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP/graph/badge.svg)](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP)
[![Go Report Card](https://goreportcard.com/badge/github.com/Lllllllleong/wazePoliceScraperGCP?style=flat)](https://goreportcard.com/report/github.com/Lllllllleong/wazePoliceScraperGCP)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Terraform](https://img.shields.io/badge/Terraform-1.10-7B42BC?logo=terraform&logoColor=white)](https://www.terraform.io/)
[![Branch Protection](https://img.shields.io/badge/Branch%20Protection-Enabled-success)](https://github.com/Lllllllleong/wazePoliceScraperGCP/settings/rules)

A system for scraping, storing, and analyzing police alert data from Waze's live traffic API, built with microservices on Google Cloud Platform.

---

## Project Status: Historical

> **Data collection stopped on March 16, 2026.** The Waze API began returning 403 errors on all requests around March 16, 2026, permanently breaking the scraper. All collected data (September 26, 2025 – March 16, 2026) has been archived to GCS.
>
> The scraper and archive services have been decommissioned. The dashboard and alerts API remain live and serve the historical dataset. See [docs/POST_MORTEM.md](./docs/POST_MORTEM.md) for the full write-up.

---

## Live Demo

A live version of the data analysis dashboard is deployed and accessible here:

**[https://dashboard.whyhireleong.com/](https://dashboard.whyhireleong.com/)**


![Dashboard Demo](./assets/AlertDashboardWeekendDemo.gif)

---

## Core Features

*   **Map + Timeline**: Leaflet.js map and alert lifespan timeline view.
*   **Filtering**: Filter by subtype and street.
*   **API**: Firebase Anonymous Auth, per-user rate limiting, gzip JSONL responses.
*   **IaC**: Terraform for all GCP infrastructure.
*   **CI/CD**: GitHub Actions — lint, test, Docker build, push, deploy.
*   **Serverless**: Cloud Run + Firestore.

---

## Tech Stack

| Category      | Technology                                                                                                                            |
|---------------|---------------------------------------------------------------------------------------------------------------------------------------|
| **Backend**   | <img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white" alt="Go">                                                    |
| **Frontend**  | <img src="https://img.shields.io/badge/JavaScript-F7DF1E?logo=javascript&logoColor=black" alt="JavaScript"> <img src="https://img.shields.io/badge/HTML5-E34F26?logo=html5&logoColor=white" alt="HTML5"> <img src="https://img.shields.io/badge/CSS3-1572B6?logo=css3&logoColor=white" alt="CSS3"> |
| **Cloud**     | <img src="https://img.shields.io/badge/Google_Cloud-4285F4?logo=google-cloud&logoColor=white" alt="Google Cloud"> <img src="https://img.shields.io/badge/Cloud_Run-4285F4" alt="Cloud Run"> <img src="https://img.shields.io/badge/Firestore-FFCA28?logo=firebase&logoColor=black" alt="Firestore"> <img src="https://img.shields.io/badge/Firebase-FFCA28?logo=firebase&logoColor=black" alt="Firebase Hosting"> |
| **IaC**       | <img src="https://img.shields.io/badge/Terraform-7B42BC?logo=terraform&logoColor=white" alt="Terraform">                                |
| **CI/CD**     | <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?logo=github-actions&logoColor=white" alt="GitHub Actions"> <img src="https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white" alt="Docker"> |
| **Mapping**   | <img src="https://img.shields.io/badge/Leaflet-199900" alt="Leaflet.js">                                                              |

---

## Architecture Overview

*   **`scraper-service`**: Cloud Run job triggered by Cloud Scheduler. Fetches data from Waze, deduplicates by UUID, writes to Firestore.
*   **`alerts-service`**: HTTP API for the frontend. Checks GCS archives first, falls back to Firestore. Streams gzip JSONL.
*   **`archive-service`**: Daily Cloud Run job. Moves the previous day's Firestore data to GCS as `archives/YYYY-MM-DD.jsonl.gz`.

See **[Architecture Document](./docs/ARCHITECTURE.md)** for details.

---

## Why I Built This

Curiosity during drives between Sydney and Canberra. Technical decisions are in the [ADR](./docs/ADR.md).

---

## Project Documentation

*   **[ARCHITECTURE.md](./docs/ARCHITECTURE.md)**: A detailed explanation of the system's architecture, components, and data flow.
*   **[ADR.md](./docs/ADR.md)**: An Architectural Decision Record (ADR) that chronicles the key engineering decisions and trade-offs made during development.
*   **[TERRAFORM_MIGRATION_SUMMARY.md](./terraform/TERRAFORM_MIGRATION_SUMMARY.md)**: Documentation of the infrastructure migration to Terraform.
*   **[SECURITY.md](./SECURITY.md)**: Security considerations and documentation of public-safe configurations.
*   **[TESTING.md](./docs/TESTING.md)**: Comprehensive testing guide covering architecture, best practices, and CI/CD integration.

---

## Getting Started (Local Development)

> **Note**: The scraper and archive services have been decommissioned and are no longer deployed. The steps below are preserved for reference — only the alerts service and frontend are actively running.

### Prerequisites
*   Go (1.24+)
*   Google Cloud SDK (`gcloud`)
*   Firebase CLI (`firebase-tools`)
*   Docker (for building and deploying containerized services)
*   Terraform (1.0+) - for infrastructure deployment
*   A Google Cloud Platform account with billing enabled

### 1. Clone the Repository
```bash
git clone https://github.com/Lllllllleong/wazePoliceScraperGCP.git
cd wazePoliceScraperGCP
```

### 2. Configure Environment
Create a `.env` file in the root directory by copying the template:
```bash
cp .env.example .env
```
Edit the `.env` file and set the following variables:

| Variable             | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| `GCP_PROJECT_ID`     | Your Google Cloud project ID.                                               |
| `FIRESTORE_COLLECTION` | The name of the Firestore collection to store police alerts (default: `police_alerts`). |
| `GCS_BUCKET_NAME`    | The name of the Google Cloud Storage bucket for archiving old alerts.       |
| `RATE_LIMIT_PER_MINUTE` | Rate limit per user for the alerts service (defaults to 30).             |
| `PORT`               | The port for the backend services to run on (defaults to 8080).             |
| `FIREBASE_AUTH_EMULATOR_HOST` | (Optional) For local development with Firebase emulator (e.g., `localhost:9099`). |

**Note**: Bounding boxes for the scraper are configured in [`configs/bboxes.yaml`](configs/bboxes.yaml), not via environment variables.


### 3. Authenticate with Google Cloud
```bash
gcloud auth login
gcloud auth application-default login
gcloud config set project YOUR_GCP_PROJECT_ID
```

### 4. Run a Backend Service

**Alerts service** (still active):
```bash
# Load environment variables (on Linux/macOS)
export $(cat .env | xargs)

go run ./cmd/alerts-service/main.go
```
The service will start on `http://localhost:8080`.

**Scraper / archive services** (decommissioned — for reference only):
```bash
go run ./cmd/scraper-service/main.go
go run ./cmd/archive-service/main.go
```

### 5. Run the Frontend Dashboard
```bash
cd dataAnalysis
firebase emulators:start
```
Dashboard: `http://localhost:5000`  
Firebase Auth Emulator: `localhost:9099`  
Configuration in `dataAnalysis/public/config.js` detects localhost and uses emulator endpoints.

---

## Testing

For complete testing documentation, see **[docs/TESTING.md](docs/TESTING.md)**.

### Quick Test Commands

**Backend:**
```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run integration tests (requires Firestore emulator)
export FIRESTORE_EMULATOR_HOST=localhost:8080
go test -tags=integration -v ./internal/storage/...
```

**Frontend:**
```bash
cd dataAnalysis
npm test                # Run all tests
npm run test:coverage   # With coverage report
```

### Coverage Status

| Component | Coverage |
|-----------|----------|
| Backend (Go) | ~15% (target: 60%+) |
| Frontend (JS) | 100% |
| Integration Tests | Firestore emulator |

Tests run automatically in CI/CD with race detection. See [docs/TESTING.md](docs/TESTING.md) for details.

---

## Configuring Geographic Areas

The scraper fetches data for specific geographic regions defined in [`configs/bboxes.yaml`](configs/bboxes.yaml).

### Adding/Modifying Bounding Boxes

1. Find your desired area coordinates using [Bounding Box Tool](https://boundingbox.klokantech.com/)
2. Select **CSV** format (returns: `west,south,east,north`)
3. Add to `configs/bboxes.yaml`:

```yaml
bboxes:
  - name: "Your Region Name"
    bbox: "west,south,east,north"  # longitude_min,latitude_min,longitude_max,latitude_max
    description: "Description of the area"
```

**Example**:
```yaml
  - name: "Sydney CBD"
    bbox: "151.1,-33.9,151.3,-33.8"
    description: "Sydney Central Business District"
```

### Current Coverage
The default configuration covers the Sydney-Canberra corridor (Hume Highway) with 4 overlapping bounding boxes.

---

## Troubleshooting

### Common Issues

#### Firebase Emulator Connection Errors
```bash
Error: Failed to connect to Firebase Auth Emulator
```
**Solution**: Ensure the emulator is running and `FIREBASE_AUTH_EMULATOR_HOST=localhost:9099` is set in your environment.

#### GCP Authentication Errors
```bash
Error: could not find default credentials
```
**Solution**: Run `gcloud auth application-default login` to set up local credentials.

#### Port Already in Use
```bash
Error: bind: address already in use
```
**Solution**: Change the `PORT` environment variable or kill the process using the port:
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9
```

#### Docker Build Failures
```bash
Error: failed to solve with frontend dockerfile.v0
```
**Solution**: Ensure Docker daemon is running and you have sufficient disk space. Try cleaning up:
```bash
docker system prune -a
```

#### Firestore Permission Denied
```bash
Error: 7 PERMISSION_DENIED
```
**Solution**: Verify your service account has the `roles/datastore.user` role and the Firestore API is enabled.

---

## Monitoring and Logs

### Viewing Service Logs

#### Cloud Run Logs
```bash
# View scraper service logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=scraper-service" --limit 50 --format json

# View alerts service logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=alerts-service" --limit 50

# Stream logs in real-time
gcloud logging tail "resource.type=cloud_run_revision"
```

#### View Logs in Console
1. Navigate to [Cloud Run Console](https://console.cloud.google.com/run)
2. Click on the service name
3. Go to **Logs** tab

### Monitoring Dashboards

*   **Cloud Run Metrics**: Request count, latency, error rates
*   **Firestore Metrics**: Read/write operations, document counts
*   **Cloud Storage**: Bucket size, request counts
*   **Cloud Scheduler**: Job execution history

Access via [GCP Monitoring Console](https://console.cloud.google.com/monitoring).

---

## Cost Estimates

Based on typical usage patterns for this system:

**Post-decommission** (current state — alerts service + frontend only):

| Service                | Usage                          | Estimated Monthly Cost |
|------------------------|--------------------------------|------------------------|
| **Cloud Run**          | 1 service (alerts), low traffic | $0 - $1               |
| **Firestore**          | Read-only, low traffic         | ~$0                    |
| **Cloud Storage**      | ~10GB static archives          | ~$0.20                 |
| **Artifact Registry**  | Container image storage        | $0.10 - $0.50          |
| **Firebase Hosting**   | Static site, minimal bandwidth | Free tier              |
| **Networking**         | Egress traffic                 | $0 - $1                |

**Total (current)**: ~**$0.30 - $2/month**

<details>
<summary>Original (when fully operational)</summary>

| Service                | Usage                          | Estimated Monthly Cost |
|------------------------|--------------------------------|------------------------|
| **Cloud Run**          | 3 services, minimal traffic    | $0 - $5                |
| **Firestore**          | ~50K writes, 100K reads/day    | $1 - $10               |
| **Cloud Storage**      | 10GB storage + requests        | $0.20 - $1             |
| **Cloud Scheduler**    | 2 jobs (scraper + archive)     | $0.10                  |
| **Artifact Registry**  | Container image storage        | $0.10 - $0.50          |
| **Firebase Hosting**   | Static site, minimal bandwidth | Free tier              |
| **BigQuery**           | (Optional) Minimal usage       | $0 - $1                |
| **Networking**         | Egress traffic                 | $1 - $5                |

**Total**: $2 - $25/month. Assumes scraper every 5-10 minutes, 30-day Firestore retention, low dashboard traffic.

</details>

### Cost Tips
*   Cloud Run minimum instances = 0 (scale to zero)
*   Set budget alerts in GCP Console
*   Coldline Storage for archives ($0.004/GB/month)
*   Monitor via [GCP Billing Dashboard](https://console.cloud.google.com/billing)

---

## API Documentation

### Alerts Service Endpoint

**Base URL**: `https://alerts-service-<hash>-uc.a.run.app` (Cloud Run URL)

#### `GET /police_alerts`

Retrieve police alerts for specified dates.

**Authentication**: Required (Firebase ID Token)

**Headers**:
```http
Authorization: Bearer <FIREBASE_ID_TOKEN>
```

**Query Parameters**:
```
dates=2026-01-08,2026-01-09
```

Maximum 7 dates per request.

**Example Request**:
```
GET /police_alerts?dates=2026-01-08,2026-01-09
```

**Response**: JSONL stream (GZIP compressed)
```jsonl
{"UUID":"...","Type":"POLICE","Subtype":"POLICE_VISIBLE","PublishTime":"2026-01-08T10:30:00Z","ExpireTime":"2026-01-08T11:00:00Z",...}
{"UUID":"...","Type":"POLICE","Subtype":"POLICE_HIDING","PublishTime":"2026-01-08T11:45:00Z","ExpireTime":"2026-01-08T12:15:00Z",...}
```

**Note**: No JSON tags — field names in responses match Go struct names. See [Data Schema](#data-schema) for the full list.

**Rate Limiting**: 30 requests per minute per authenticated user

**Error Responses**:
*   `401 Unauthorized`: Missing or invalid Firebase token
*   `429 Too Many Requests`: Rate limit exceeded
*   `400 Bad Request`: Invalid date format
*   `500 Internal Server Error`: Server-side error

---

## Data Schema

### PoliceAlert Model

Stored in Firestore collection `police_alerts`:

```go
type PoliceAlert struct {
    UUID         string    `firestore:"uuid"`           // Unique identifier from Waze
    ID           string    `firestore:"id,omitempty"`   // Additional ID field
    Type         string    `firestore:"type"`           // Alert type (e.g., "POLICE")
    Subtype      string    `firestore:"subtype"`        // Alert subtype (e.g., "POLICE_VISIBLE")
    Street       string    `firestore:"street,omitempty"`      // Street name
    City         string    `firestore:"city,omitempty"`        // City name
    Country      string    `firestore:"country,omitempty"`     // Country code
    LocationGeo  *latlng.LatLng `firestore:"location_geo"` // Geographic coordinates (latitude/longitude)
    Reliability  int       `firestore:"reliability,omitempty"` // Reliability score
    Confidence   int       `firestore:"confidence,omitempty"`   // Confidence level
    ReportRating int       `firestore:"report_rating,omitempty"` // Report rating
    PublishTime  time.Time `firestore:"publish_time"`   // When alert was first published (from pubMillis)
    ScrapeTime   time.Time `firestore:"scrape_time"`    // First time we scraped this alert
    ExpireTime   time.Time `firestore:"expire_time"`    // Last time we saw this alert (assumed expired after)
    LastVerificationTime *time.Time `firestore:"last_verification_time,omitempty"` // Latest comment timestamp
    ActiveMillis           int64  `firestore:"active_millis"` // Alert duration (expireMillis - pubMillis)
    LastVerificationMillis *int64 `firestore:"last_verification_millis,omitempty"` // Latest comment reportMillis
    NThumbsUpInitial int `firestore:"n_thumbs_up_initial"` // Initial thumbs up count
    NThumbsUpLast    int `firestore:"n_thumbs_up_last"`    // Most recent thumbs up count
    RawDataInitial string `firestore:"raw_data_initial"` // First scrape JSON
    RawDataLast    string `firestore:"raw_data_last"`    // Most recent scrape JSON
}
```

**Note**: No JSON tags on `PoliceAlert`, so field names in API responses match Go struct names (`UUID`, `PublishTime`, `ExpireTime`, etc.).

### Alert Subtypes

Common police alert subtypes from Waze:
*   `POLICE_VISIBLE`: Visible police presence
*   `POLICE_HIDING`: Hidden/speed trap
*   `POLICE_WITH_MOBILE_CAMERA`: Police with mobile speed camera
*   `POLICE_GENERAL`: General police alert (empty string subtype)

### Archive Format (GCS)

Archived alerts are stored as **JSONL** (JSON Lines) files in Cloud Storage:

**File Path**: `gs://BUCKET_NAME/archives/YYYY-MM-DD.jsonl.gz`

**Format**: One JSON object per line, stored uncompressed. The `.gz` extension is a misnomer — gzip compression only applies to the HTTP response from the alerts service, not the stored file.

---

## Project Structure

```
.
├── cmd/                  # Main applications for the microservices
│   ├── alerts-service/   # Serves alert data to the frontend
│   ├── archive-service/  # Archives old data from Firestore to GCS
│   └── scraper-service/  # Scrapes police alerts from Waze
├── dataAnalysis/         # Frontend dashboard application
├── internal/             # Shared Go packages
│   ├── models/           # Data models for alerts and Waze API
│   ├── storage/          # Firestore and GCS storage logic
│   └── waze/             # Waze API client
├── terraform/            # Infrastructure as Code (Terraform)
│   ├── environments/     # Environment-specific configurations
│   └── modules/          # Reusable Terraform modules
└── .github/workflows/    # CI/CD workflows for GitHub Actions
```

---

## Contact

For questions, issues, or collaboration:

*   **GitHub**: [@Lllllllleong](https://github.com/Lllllllleong)
*   **Email**: chanleongyin8@gmail.com
*   **Issues**: [GitHub Issues](https://github.com/Lllllllleong/wazePoliceScraperGCP/issues)

---

## Acknowledgments

*   **Waze**: Live traffic data API
*   **Google Cloud Platform**: Serverless infrastructure
*   **Leaflet.js**: Open-source mapping library
*   **Firebase**: Authentication and hosting
*   Open-source community for libraries and tools

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Deployment

### Automated Deployment

GitHub Actions workflows in `.github/workflows/` handle:
1.  Lint and test
2.  Docker build and push to Artifact Registry
3.  Deploy to Cloud Run

### Manual Deployment with Terraform

#### Prerequisites
1. **Enable Required GCP APIs**:
   ```bash
   gcloud services enable run.googleapis.com \
     firestore.googleapis.com \
     storage-api.googleapis.com \
     cloudscheduler.googleapis.com \
     artifactregistry.googleapis.com \
     bigquery.googleapis.com
   ```

2. **Create a GCS bucket for Terraform state**:
   ```bash
   gcloud storage buckets create gs://YOUR-PROJECT-terraform-state --location=us-central1
   ```

#### Deploy Infrastructure

1. **Navigate to Terraform directory**:
   ```bash
   cd terraform/environments/prod
   ```

2. **Update `terraform.tfvars`** with your project details:
   ```hcl
   project_id = "your-project-id"
   region     = "us-central1"
   ```

3. **Update backend configuration** in `backend.tf`:
   ```hcl
   bucket = "your-project-terraform-state"
   ```

4. **Initialize and deploy**:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

#### Deploy Frontend Dashboard

1. **Update Firebase configuration** in `dataAnalysis/public/config.js` with your project details.

2. **Deploy to Firebase Hosting**:
   ```bash
   cd dataAnalysis
   firebase login
   firebase use --add  # Select your Firebase project
   firebase deploy --only hosting
   ```

#### Build and Deploy Services Manually

If you need to deploy services without CI/CD:

```bash
# Build and push Docker image
docker build -f Dockerfile.scraper -t us-central1-docker.pkg.dev/PROJECT_ID/waze-scraper/scraper-service:latest .
docker push us-central1-docker.pkg.dev/PROJECT_ID/waze-scraper/scraper-service:latest

# Deploy to Cloud Run
gcloud run deploy scraper-service \
  --image us-central1-docker.pkg.dev/PROJECT_ID/waze-scraper/scraper-service:latest \
  --region us-central1 \
  --platform managed
```

For detailed infrastructure documentation, see [`terraform/TERRAFORM_MIGRATION_SUMMARY.md`](terraform/TERRAFORM_MIGRATION_SUMMARY.md).