# Gemini Code Assistant Context

This document provides context for the Gemini Code Assistant to understand the Waze Police Scraper GCP project.

## Project Overview

This project is a cloud-based system for scraping, storing, and analyzing police alerts from Waze's live traffic data. It consists of three main components:

1.  **Scraper Service:** A Go-based Cloud Run service that fetches police alerts from the Waze API on a scheduled basis.
2.  **Data Analysis Dashboard:** An interactive web interface for visualizing and analyzing police alert patterns. There are two implementations: a production-ready vanilla JavaScript version and a React prototype.

All data is stored in Google Cloud Firestore.

## Building and Running

### Backend (Go)

**Prerequisites:**

*   Go 1.21+
*   Google Cloud Platform account with a configured project.

**Environment Variables:**

Create a `.env` file in the project root:

```bash
GCP_PROJECT_ID=your-project-id
FIRESTORE_COLLECTION=police_alerts
WAZE_BBOXES="150.388,-34.255,151.009,-33.938;149.589,-34.769,150.830,-34.139"
```

**Running Locally:**

*   **Scraper:**
    ```bash
    export GCP_PROJECT_ID=your-project-id
    export FIRESTORE_COLLECTION=police_alerts
    go run cmd/scraper/main.go
    ```

**Deployment:**

The project is deployed to Google Cloud Run. Deployment scripts are available in the `scripts` directory.

*   `./scripts/deploy.sh`: Deploys the scraper service.


### Frontend (JavaScript)

The vanilla JavaScript dashboard is located in `dataAnalysis/public/`. It can be opened directly in a browser or deployed to a static hosting service like Firebase Hosting.

The React prototype is in `dataAnalysis/react-prototype/` and can be run using `npm install` and `npm run dev`.

## Development Conventions

*   **Go:** The backend is written in Go. Code is organized into `cmd` for executables and `internal` for shared packages.
*   **Frontend:** The production dashboard uses vanilla JavaScript, while a React/TypeScript prototype is also available.
*   **Configuration:** Configuration is managed through environment variables and YAML files (e.g., `configs/bboxes.yaml`).
*   **Security:** Security is a key consideration, with detailed documentation in `SECURITY_CONFIG.md`.
*   **Deployment:** Deployment is automated via shell scripts and Google Cloud Build.
