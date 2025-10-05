# GEMINI.md

> **Note**: This file provides project context for AI assistants. For human-readable documentation, see [README.md](./README.md).

## Project Overview

This is a Go-based application designed to scrape police alert data from Waze and store it in Google Cloud Firestore. The project is architected for deployment on Google Cloud Platform (GCP) and includes components for automated scraping, data export, and lifecycle tracking of alerts.

The core components are:
- **Scraper (`cmd/scraper`)**: A Cloud Run service that, when triggered, fetches police alerts from the Waze API for pre-configured or dynamically provided geographic bounding boxes. It processes these alerts and saves them to a Firestore collection, tracking their lifecycle (first seen, last seen, verification count, active duration).
- **Exporter (`cmd/exporter`)**: A command-line utility to export the collected police alert data from Firestore into JSON or JSONL files based on a specified date range.
- **Cloud Scheduler**: The intended mechanism for triggering the scraper service at regular intervals (e.g., every 2 minutes).

The application is built with a modular structure, separating concerns for the Waze API client (`internal/waze`), Firestore storage (`internal/storage`), and data models (`internal/models`).

## Building and Running

### Local Development

**Prerequisites:**
- Go 1.21+
- Google Cloud SDK (`gcloud`)

**Running the Scraper Locally:**
1.  **Set Environment Variables:**
    ```bash
    export GCP_PROJECT_ID="your-gcp-project-id"
    export FIRESTORE_COLLECTION="police_alerts" # Optional, defaults to police_alerts
    ```
2.  **Authenticate with GCP:**
    ```bash
    gcloud auth application-default login
    ```
3.  **Run the scraper:**
    ```bash
    go run cmd/scraper/main.go
    ```
    The service will start on `http://localhost:8080`.

**Running the Exporter Locally:**
1.  **Set Environment Variables:**
    ```bash
    export GCP_PROJECT_ID="your-gcp-project-id"
    ```
2.  **Run the exporter with required flags:**
    ```bash
    go run cmd/exporter/main.go --start YYYY-MM-DD --end YYYY-MM-DD --output alerts.jsonl --format jsonl
    ```

### Deployment

The project is designed to be deployed as a Docker container to Google Cloud Run.

**Deployment Command:**
-   **Linux/macOS:**
    ```bash
    ./scripts/deploy.sh
    ```
-   **Windows:**
    ```bash
    .\scripts\deploy.bat
    ```
The deployment script handles building the Docker image, pushing it to Google Artifact Registry, and deploying it as a Cloud Run service.

## Development Conventions

### Configuration
- Application configuration is primarily managed through environment variables (`GCP_PROJECT_ID`, `FIRESTORE_COLLECTION`, `WAZE_BBOXES`, `PORT`).
- Default values are provided for some configurations (e.g., `FIRESTORE_COLLECTION`, default bounding boxes).

### Code Structure
- The project follows a standard Go project layout.
- `cmd/`: Contains the main applications (scraper and exporter).
- `internal/`: Contains the core business logic, separated into:
    - `models/`: Defines the data structures for the application.
    - `storage/`: Handles interactions with Firestore.
    - `waze/`: Contains the client for the Waze API.
- `scripts/`: Contains deployment scripts.
- `configs/`: Contains configuration documentation.

### Data Models
- `internal/models/alert.go` defines two main structs:
    - `WazeAlert`: Represents a raw alert from the Waze API.
    - `PoliceAlert`: Represents the data structure stored in Firestore, which includes lifecycle tracking fields.

### Dependencies
- Go modules are used for dependency management (`go.mod`, `go.sum`).
- Key dependencies include the Google Cloud Firestore SDK.
