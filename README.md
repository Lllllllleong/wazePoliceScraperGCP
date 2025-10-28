# Waze Police Alert Analysis System

[![CI/CD Status](https://img.shields.io/badge/CI%2FCD-Passing-brightgreen)](https://github.com/Lllllllleong/wazePoliceScraperGCP/actions)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A professional, cloud-native system for scraping, storing, and analyzing police alert data from Waze's live traffic feed. This portfolio project demonstrates a complete, production-ready application built with a microservices architecture on Google Cloud Platform.

---

### ✨ Live Demo

A live version of the data analysis dashboard is deployed and accessible here:

**[https://policealert.whyhireleong.com/](https://policealert.whyhireleong.com/)**

*(Note: Data collection was terminated on Oct 31, 2025, as the methodology was found to violate Waze's Terms of Service. The existing data is retained for demonstration and analysis.)*

![Dashboard Screenshot](https://i.imgur.com/your-screenshot-url.png) 
*(A placeholder for a screenshot of the dashboard in action)*

---

## 🚀 Core Features

*   **Automated Data Scraping**: A serverless Go service runs on a schedule to automatically fetch and store police alert data.
*   **Interactive Map Visualization**: A rich frontend dashboard built with vanilla JavaScript and Leaflet.js to display alerts on an interactive map.
*   **High-Fidelity Timeline**: Accurately visualizes the true lifespan of each alert, allowing for powerful temporal analysis.
*   **Advanced Filtering**: A dynamic, tag-based UI to filter data by multiple subtypes and streets.
*   **Microservices Architecture**: A robust backend composed of distinct services for scraping, serving data, and archiving.
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

*   **`scraper-service`**: A Go application triggered by Cloud Scheduler to fetch data from Waze and save it to Firestore.
*   **`alerts-service`**: A Go API that serves alert data to the frontend, intelligently fetching from GCS archives or live from Firestore.
*   **`archive-service`**: A Go application triggered daily to move older data from Firestore to Google Cloud Storage for long-term archival.

For a detailed breakdown of the system design, data flow, and technology rationale, please see the **[Architecture Document](./ARCHITECTURE.md)**.

---

## 📖 Project Documentation

This project adheres to a high standard of documentation to demonstrate professional development practices.

*   **[ARCHITECTURE.md](./ARCHITECTURE.md)**: A detailed explanation of the system's architecture, components, and data flow.
*   **[DECISIONS.md](./DECISIONS.md)**: An Architectural Decision Record (ADR) that chronicles the key engineering decisions and trade-offs made during development.

---

## 🚀 Getting Started (Local Development)

### Prerequisites
*   Go (1.21+)
*   Google Cloud SDK (`gcloud`)
*   Firebase CLI (`firebase-tools`)
*   Docker

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
Edit the `.env` file and set your `GCP_PROJECT_ID`.

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
The frontend is a simple static site.
```bash
cd dataAnalysis
firebase serve
```
The dashboard will be available at `http://localhost:5000`. You will need to update `dataAnalysis/public/config.js` to point to your local backend service for it to work.

---

## 🚢 Deployment

This project is configured for fully automated deployment using **GitHub Actions**.

When code is pushed to the `main` branch, the CI/CD workflows located in `.github/workflows/` will automatically:
1.  Lint and test the Go source code.
2.  Build a Docker container for the relevant service.
3.  Push the container to Google Artifact Registry.
4.  Deploy the new container version to Google Cloud Run.

Manual deployment is also possible using the scripts in the `/scripts` directory, which mirror the CI/CD process.