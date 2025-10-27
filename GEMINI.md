# Project Overview

This project is a Go-based system for scraping, storing, and analyzing police alerts from Waze's live traffic data. It consists of several microservices for scraping, serving, and archiving data. The system collects police presence alerts and provides an interactive web interface for analyzing and visualizing the data. All data is stored in Google Cloud Firestore.

## Technologies Used

*   **Backend:** Go
*   **Frontend:** HTML, CSS, JavaScript
*   **Database:** Google Cloud Firestore
*   **Deployment:** Docker, Google Cloud Build, Google Cloud Run
*   **Mapping:** Leaflet.js
*   **Date Picker:** Flatpickr

## Architecture

The project follows a microservices architecture:

*   **Scraper Service:** A Cloud Run service that fetches police alerts from the Waze API on a scheduled basis.
*   **Alerts Service:** An API that serves police alert data to the frontend dashboard.
*   **Archive Service:** A service for archiving old police alert data.
*   **Data Analysis Dashboard:** An interactive web interface for visualizing and analyzing police alert patterns.

## Building and Running

### Prerequisites

*   Go 1.21+
*   Google Cloud Platform account
*   Firebase CLI

### Installation

1.  Install Go dependencies:
    ```bash
    go mod download
    ```

### Running Locally

1.  Run the scraper locally:
    ```bash
    export GCP_PROJECT_ID=your-project-id
    export FIRESTORE_COLLECTION=police_alerts
    go run cmd/scraper/main.go
    ```

### Deployment

1.  Deploy the scraper service:
    ```bash
    ./scripts/deploy.sh
    ```
2.  Deploy the dashboard to Firebase Hosting:
    ```bash
    cd dataAnalysis
    firebase deploy --only hosting
    ```

## Development Conventions

*   **Code Style:** The project follows standard Go and JavaScript conventions.
*   **Data Model:** Police alert data is stored in a Firestore collection. The data model is defined in `internal/models/alert.go`.
*   **Configuration:** Configuration is managed through environment variables. A `.env.example` file is provided as a template.
*   **API:** The API for the alerts service is defined in `cmd/alerts-service/main.go`.
