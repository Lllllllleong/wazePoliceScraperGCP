# Waze Police Alert Analysis System

[![CI/CD Status](https://img.shields.io/badge/CI%2FCD-Passing-brightgreen)](https://github.com/Lllllllleong/wazePoliceScraperGCP/actions)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A professional, cloud-native system for scraping, storing, and analyzing police alert data from Waze's live traffic feed. This portfolio project demonstrates a complete, production-ready application built with a microservices architecture on Google Cloud Platform.

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
*   Go (1.21+)
*   Google Cloud SDK (`gcloud`)
*   Firebase CLI (`firebase-tools`)
*   Docker (for building and deploying containerized services)

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
| `FIRESTORE_COLLECTION` | The name of the Firestore collection to store police alerts.                |
| `GCS_BUCKET_NAME`    | The name of the Google Cloud Storage bucket for archiving old alerts.       |
| `WAZE_BBOXES`        | A comma-separated list of bounding boxes for Waze alert scraping.           |
| `RATE_LIMIT_PER_MINUTE` | Rate limit per user for the alerts service (defaults to 30).             |
| `PORT`               | The port for the backend services to run on (defaults to 8080).             |


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

## 📁 Project Structure
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

## 🚢 Deployment

This project is configured for fully automated deployment using **GitHub Actions**.

When code is pushed to the `main` branch, the CI/CD workflows located in `.github/workflows/` will automatically:
1.  Lint and test the Go source code.
2.  Build a Docker container for the relevant service.
3.  Push the container to Google Artifact Registry.
4.  Deploy the new container version to Google Cloud Run.

Manual deployment is also possible using the Terraform configurations in the `/terraform` directory and the Dockerfiles for each service.