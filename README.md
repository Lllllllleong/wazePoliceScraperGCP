# Waze Police Alert Analysis System

[![CI/CD Status](https://img.shields.io/badge/CI%2FCD-Passing-brightgreen)](https://github.com/Lllllllleong/wazePoliceScraperGCP/actions)
[![codecov](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP/graph/badge.svg)](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP)
[![Go Report Card](https://goreportcard.com/badge/github.com/Lllllllleong/wazePoliceScraperGCP)](https://goreportcard.com/report/github.com/Lllllllleong/wazePoliceScraperGCP)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Branch Protection](https://img.shields.io/badge/Branch%20Protection-Enabled-success)](https://github.com/Lllllllleong/wazePoliceScraperGCP/settings/rules)

A professional, cloud-native system for scraping, storing, and analyzing police alert data from Waze's live traffic feed. This project demonstrates a complete, production-ready application built with a microservices architecture on Google Cloud Platform.

---

### ✨ Live Demo

A live version of the data analysis dashboard is deployed and accessible here:

**[https://dashboard.whyhireleong.com/](https://dashboard.whyhireleong.com/)**


![Dashboard Demo](./assets/AlertDashboardWeekendDemo.gif)

---

## 🚀 Core Features

*   **Automated Data Scraping**: A serverless Go service runs on a schedule to automatically fetch and store police alert data.
*   **Interactive Map Visualization**: A rich frontend dashboard built with vanilla JavaScript and Leaflet.js to display alerts on an interactive map.
*   **High-Fidelity Timeline**: Accurately visualizes the true lifespan of each alert, allowing for powerful temporal analysis.
*   **Advanced Filtering**: A dynamic, tag-based UI to filter data by multiple subtypes and streets.
*   **Microservices Architecture**: A robust backend composed of distinct services for scraping, serving data, and archiving.
*   **Secure API**: Protected by Firebase Anonymous Authentication with per-user rate limiting.
*   **Infrastructure as Code**: Full Terraform implementation for reproducible infrastructure deployment.
*   **CI/CD Automation**: Fully automated build, test, and deployment pipelines using GitHub Actions.
*   **Serverless & Scalable**: Built entirely on serverless technologies (Cloud Run, Firestore) for cost-efficiency and scalability.

---

## 🛠️ Tech Stack

| Category      | Technology                                                                                                                            |
|---------------|---------------------------------------------------------------------------------------------------------------------------------------|
| **Backend**   | <img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white" alt="Go">                                                    |
| **Frontend**  | <img src="https://img.shields.io/badge/JavaScript-F7DF1E?logo=javascript&logoColor=black" alt="JavaScript"> <img src="https://img.shields.io/badge/HTML5-E34F26?logo=html5&logoColor=white" alt="HTML5"> <img src="https://img.shields.io/badge/CSS3-1572B6?logo=css3&logoColor=white" alt="CSS3"> |
| **Cloud**     | <img src="https://img.shields.io/badge/Google_Cloud-4285F4?logo=google-cloud&logoColor=white" alt="Google Cloud"> <img src="https://img.shields.io/badge/Cloud_Run-4285F4" alt="Cloud Run"> <img src="https://img.shields.io/badge/Firestore-FFCA28?logo=firebase&logoColor=black" alt="Firestore"> <img src="https://img.shields.io/badge/Firebase-FFCA28?logo=firebase&logoColor=black" alt="Firebase Hosting"> |
| **CI/CD**     | <img src="https://img.shields.io/badge/GitHub_Actions-2088FF?logo=github-actions&logoColor=white" alt="GitHub Actions"> <img src="https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white" alt="Docker"> |
| **Mapping**   | <img src="https://img.shields.io/badge/Leaflet-199900" alt="Leaflet.js">                                                              |

---

## 🏗️ Architecture Overview

The system is designed as a set of cooperating microservices deployed on Google Cloud Run. This serverless architecture ensures that resources are only consumed when a service is active, making it highly cost-effective.

*   **`scraper-service`**: A Go application on Cloud Run, triggered by Cloud Scheduler, that fetches data from Waze and saves it to Firestore.
*   **`alerts-service`**: A Go API on Cloud Run that serves alert data to the frontend with Firebase Authentication and rate limiting, intelligently fetching from GCS archives or live from Firestore with GZIP-compressed JSONL streaming.
*   **`archive-service`**: A Go application on Cloud Run, triggered daily by Cloud Scheduler, that moves older data from Firestore to Google Cloud Storage for long-term archival.

For a detailed breakdown of the system design, data flow, and technology rationale, please see the **[Architecture Document](./ARCHITECTURE.md)**.

---

## 💡 Why I Built This

This project was spawned out of curiosity developed from my numerous drives between Sydney and Canberra. I also took this project as a chance to demonstrate production-grade software engineering practices, including:

- **Microservices Architecture**: Designing loosely-coupled, independently deployable services
- **Cloud-Native Development**: Leveraging serverless technologies for cost efficiency and scalability
- **Infrastructure as Code**: Managing infrastructure declaratively with Terraform
- **API Security**: Implementing authentication, rate limiting, and CORS protection
- **Data Streaming**: Efficient handling of large datasets with JSONL streaming and GZIP compression
- **CI/CD Automation**: Full automation from code commit to production deployment

The technical decisions made throughout development are documented in the [ADR (Architectural Decision Record)](./docs/ADR.md).

---

## 📖 Project Documentation

This project adheres to a high standard of documentation to demonstrate professional development practices.

*   **[ARCHITECTURE.md](./ARCHITECTURE.md)**: A detailed explanation of the system's architecture, components, and data flow.
*   **[ADR.md](./docs/ADR.md)**: An Architectural Decision Record (ADR) that chronicles the key engineering decisions and trade-offs made during development.
*   **[TERRAFORM_MIGRATION_SUMMARY.md](./terraform/TERRAFORM_MIGRATION_SUMMARY.md)**: Documentation of the infrastructure migration to Terraform.
*   **[SECURITY.md](./SECURITY.md)**: Security considerations and documentation of public-safe configurations.

---

## 🚀 Getting Started (Local Development)

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

### 4. Run a Backend Service (e.g., Scraper)
```bash
# Load environment variables (on Linux/macOS)
export $(cat .env | xargs)

# Run the scraper service
go run ./cmd/scraper-service/main.go
```
The service will start on `http://localhost:8080`.

### 5. Run the Frontend Dashboard
The frontend is a simple static site served via Firebase Hosting emulator.
```bash
cd dataAnalysis
firebase emulators:start
```
The dashboard will be available at `http://localhost:5000` with Firebase Auth Emulator at `localhost:9099`. The configuration in `dataAnalysis/public/config.js` automatically detects localhost and uses the appropriate endpoints and emulators.

---

## 🧪 Testing

This project maintains comprehensive test coverage across both backend and frontend components, demonstrating professional testing practices.

### Backend Tests (Go)

```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run tests for specific packages
go test -v ./internal/models/
go test -v ./internal/waze/
go test -v ./internal/storage/
go test -v ./cmd/scraper-service/
go test -v ./cmd/alerts-service/
go test -v ./cmd/archive-service/
```

### Frontend Tests (JavaScript)

```bash
cd dataAnalysis

# Install dependencies
npm install

# Run tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage report
npm run test:coverage
```

### Test Structure

**Backend (`internal/` and `cmd/`):**
*   **`internal/models/`**: Data model validation and API request/response structures
*   **`internal/waze/`**: Waze API client, HTTP mocking, BBox parsing, deduplication logic
*   **`internal/storage/`**: Firestore operations, filtering logic, alert lifecycle management
*   **`cmd/scraper-service/`**: HTTP handler tests, request validation
*   **`cmd/alerts-service/`**: Middleware tests (CORS, Auth, Rate Limiting, GZIP), streaming
*   **`cmd/archive-service/`**: JSONL creation, idempotency logic, date handling

**Frontend (`dataAnalysis/tests/`):**
*   **`utils.test.js`**: Date formatting, timestamp parsing utilities
*   **`filters.test.js`**: Client-side filtering, deduplication, sorting logic
*   **`geojson.test.js`**: GeoJSON transformation for map visualization

### CI/CD Testing

All pull requests and commits trigger automated testing via GitHub Actions:
*   **Go linting** with `golangci-lint`
*   **Unit test execution** with race detection (`-race` flag)
*   **Coverage threshold enforcement** (minimum 60% coverage)
*   **Coverage reporting** to Codecov with badge generation
*   **Frontend syntax validation** and test execution
*   **Build validation** for all services

### Coverage Requirements

| Component | Current Coverage | Target |
|-----------|-----------------|--------|
| Backend (Go) | ~25% | 60%+ |
| Frontend (JS) | Setup complete | 60%+ |

**Note**: The current backend coverage is limited because many functions interact with external services (Firestore, GCS, Firebase Auth). The path to higher coverage includes:
- Adding integration tests with Firestore/Firebase emulators
- Implementing more dependency injection patterns
- Expanding HTTP mock testing

Coverage reports are automatically uploaded to [Codecov](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP) on every push.

---

## 📍 Configuring Geographic Areas

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

## � Troubleshooting

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

## 📊 Monitoring and Logs

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

## 💰 Cost Estimates

Based on typical usage patterns for this system:

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

**Total Estimated Cost**: **$2 - $25/month**

### Cost Optimization Tips
*   Use Cloud Run's **minimum instances = 0** for auto-scaling to zero
*   Set up **budget alerts** in GCP Console
*   Archive old data to **Coldline Storage** ($0.004/GB/month)
*   Enable **Firestore deletion protection** but regularly clean up old documents
*   Monitor costs via [GCP Billing Dashboard](https://console.cloud.google.com/billing)

**Note**: Costs depend heavily on scraping frequency, data retention, and API traffic. The above estimates assume:
- Scraper running every 5-10 minutes
- 30-day data retention in Firestore
- Low to moderate dashboard usage

---

## 📡 API Documentation

### Alerts Service Endpoint

**Base URL**: `https://alerts-service-<hash>-uc.a.run.app` (Cloud Run URL)

#### `POST /alerts`

Retrieve police alerts for specified dates.

**Authentication**: Required (Firebase ID Token)

**Headers**:
```http
Content-Type: application/json
Authorization: Bearer <FIREBASE_ID_TOKEN>
```

**Request Body**:
```json
{
  "dates": [
    "2026-01-08T00:00:00.000Z",
    "2026-01-09T00:00:00.000Z"
  ]
}
```

**Response**: JSONL stream (GZIP compressed)
```jsonl
{"uuid":"...","lat":-35.123,"lon":149.456,"reportTime":"2026-01-08T10:30:00Z",...}
{"uuid":"...","lat":-33.987,"lon":151.234,"reportTime":"2026-01-08T11:45:00Z",...}
```

**Response Fields**: See [Data Schema](#-data-schema) section below.

**Rate Limiting**: 30 requests per minute per authenticated user

**Error Responses**:
*   `401 Unauthorized`: Missing or invalid Firebase token
*   `429 Too Many Requests`: Rate limit exceeded
*   `400 Bad Request`: Invalid date format
*   `500 Internal Server Error`: Server-side error

---

## 🗄️ Data Schema

### PoliceAlert Model

Stored in Firestore collection `police_alerts`:

```go
type PoliceAlert struct {
    UUID         string    `json:"uuid" firestore:"uuid"`           // Unique identifier from Waze
    Latitude     float64   `json:"lat" firestore:"lat"`             // Latitude coordinate
    Longitude    float64   `json:"lon" firestore:"lon"`             // Longitude coordinate
    ReportTime   time.Time `json:"reportTime" firestore:"reportTime"` // When alert was first reported
    Street       string    `json:"street" firestore:"street"`       // Street name
    City         string    `json:"city" firestore:"city"`           // City name
    Country      string    `json:"country" firestore:"country"`     // Country code
    Subtype      string    `json:"subtype" firestore:"subtype"`     // Alert subtype (e.g., "POLICE_VISIBLE")
    Reliability  int       `json:"reliability" firestore:"reliability"` // Reliability score
    Confidence   int       `json:"confidence" firestore:"confidence"`   // Confidence level
    NumThumbsUp  int       `json:"nThumbsUp" firestore:"nThumbsUp"`    // User confirmations
    ScrapedAt    time.Time `json:"scrapedAt" firestore:"scrapedAt"`    // When we scraped this alert
}
```

### Alert Subtypes

Common police alert subtypes from Waze:
*   `POLICE_VISIBLE`: Visible police presence
*   `POLICE_HIDING`: Hidden/speed trap
*   `POLICE_GENERAL`: General police alert
*   `POLICE_CARS`: Multiple police vehicles

### Archive Format (GCS)

Archived alerts are stored as **JSONL** (JSON Lines) files in Cloud Storage:

**File Path**: `gs://BUCKET_NAME/archives/YYYY-MM-DD.jsonl.gz`

**Format**: One JSON object per line, GZIP compressed

---

## �📁 Project Structure
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

## 📬 Contact

For questions, issues, or collaboration:

*   **GitHub**: [@Lllllllleong](https://github.com/Lllllllleong)
*   **Email**: chanleongyin8@gmail.com
*   **Issues**: [GitHub Issues](https://github.com/Lllllllleong/wazePoliceScraperGCP/issues)

---

## 🙏 Acknowledgments

*   **Waze**: For providing the live traffic data API that makes this project possible
*   **Google Cloud Platform**: For the robust serverless infrastructure
*   **Leaflet.js**: For the excellent open-source mapping library
*   **Firebase**: For authentication and hosting services
*   The open-source community for countless libraries and tools used in this project

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🚢 Deployment

### Automated Deployment (Recommended)

This project is configured for fully automated deployment using **GitHub Actions**.

When code is pushed to the `main` branch, the CI/CD workflows located in `.github/workflows/` will automatically:
1.  Lint and test the Go source code.
2.  Build a Docker container for the relevant service.
3.  Push the container to Google Artifact Registry.
4.  Deploy the new container version to Google Cloud Run.

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