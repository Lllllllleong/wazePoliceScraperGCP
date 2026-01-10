# System Architecture

This document provides a comprehensive overview of the Waze Police Scraper project's architecture, detailing its components, data flow, and technology stack.

---

## 1. High-Level Overview

This project is a cloud-native, event-driven system designed to scrape, store, and visualize police alert data from Waze. It follows a microservices architecture, with distinct services for data collection, data serving, and data archiving. The entire system is deployed on Google Cloud Platform and leverages serverless technologies for scalability and cost-efficiency.

The primary components are:
*   A **Scraper Service** that periodically fetches data from Waze.
*   A **Firestore Database** that serves as the central data store for fresh alerts.
*   An **Alerts Service** (API) that provides data to the frontend dashboard with Firebase Authentication and rate limiting.
*   An **Archive Service** that moves older data to long-term storage.
*   A **Data Analysis Dashboard**, which is a vanilla JavaScript single-page application for visualization.

---

## 2. Architecture Diagram

The following diagram illustrates the flow of data and the interaction between the system's components.

> **Note**: This Mermaid diagram requires the "Markdown Preview Mermaid Support" VS Code extension to render in preview. Alternatively, view on GitHub where Mermaid is natively supported.

```mermaid
flowchart TB
    %% Class Definitions for Styling
    classDef external fill:#f9f9f9,stroke:#666,stroke-width:2px,stroke-dasharray: 5 5;
    classDef compute fill:#e1f5fe,stroke:#01579b,stroke-width:2px,color:#01579b;
    classDef storage fill:#fff8e1,stroke:#ff8f00,stroke-width:2px,color:#8d6e63;
    classDef ui fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20;
    classDef trigger fill:#ede7f6,stroke:#512da8,stroke-width:2px,color:#311b92;

    %% External Triggers
    Waze([Waze Live Data API]):::external
    SchedulerScraper[[Cloud Scheduler<br/>Every Minute]]:::trigger
    SchedulerArchive[[Cloud Scheduler<br/>Daily 00:05 UTC]]:::trigger

    %% Cloud Run Services
    Scraper(Scraper Service<br/>Go on Cloud Run):::compute
    Archive(Archive Service<br/>Go on Cloud Run):::compute
    AlertsAPI(Alerts API Service<br/>Go on Cloud Run):::compute

    %% Storage Layer
    Firestore[(Firestore<br/>NoSQL Database)]:::storage
    GCS[(Cloud Storage<br/>JSONL Archives)]:::storage

    %% Frontend
    Dashboard[Data Analysis Dashboard<br/>JavaScript + Leaflet]:::ui

    %% Data Flow: Scraper Pipeline
    SchedulerScraper -->|Trigger| Scraper
    Waze -->|Police Alerts JSON| Scraper
    Scraper -->|Write PoliceAlert| Firestore

    %% Data Flow: Archive Pipeline
    SchedulerArchive -->|Trigger| Archive
    Firestore -->|Read Yesterday's Data| Archive
    Archive -->|Write JSONL.gz| GCS

    %% Data Flow: Frontend Access
    Dashboard -->|GET /police_alerts<br/>+ Firebase Auth Token| AlertsAPI
    Firestore -.->|Recent Data| AlertsAPI
    GCS -.->|Archived Data| AlertsAPI
    AlertsAPI -.->|Stream JSONL Response| Dashboard
```


---

## 3. Component Breakdown

### 3.1. Scraper Service (`scraper-service`)
*   **Technology**: Go, deployed on Cloud Run.
*   **Trigger**: Invoked by Google Cloud Scheduler every minute (`* * * * *`).
*   **Configuration**: Geographic bounding boxes defined in `configs/bboxes.yaml` (can be overridden via `WAZE_BBOXES` environment variable).
*   **Responsibilities**:
    1.  Receives a trigger to begin a scrape cycle.
    2.  Makes HTTP requests to the Waze live data endpoints for predefined geographic bounding boxes.
    3.  Parses the JSON response to extract police-related alerts.
    4.  Performs data cleaning and transformation into a standardized `PoliceAlert` model.
    5.  Writes the processed alerts to the `police_alerts` collection in Firestore.

### 3.2. Alerts Service (`alerts-service`)
*   **Technology**: Go, deployed on Cloud Run.
*   **Trigger**: HTTPS endpoint protected by Firebase Anonymous Authentication.
*   **Responsibilities**:
    1.  Authenticates incoming requests using Firebase ID tokens.
    2.  Enforces per-user rate limiting (configurable, default 30 requests/minute).
    3.  Receives a request from the frontend containing a list of dates.
    4.  For each date, it first checks if a pre-computed archive exists in Google Cloud Storage (GCS).
    5.  If an archive exists, it streams the data directly from the GCS file.
    6.  If no archive exists, it queries Firestore for alerts within that date's 24-hour UTC window.
    7.  Streams the results back to the frontend as a single, deduplicated JSONL stream with GZIP compression.
    8.  Acts as a secure proxy, preventing direct database access from the browser.

### 3.3. Archive Service (`archive-service`)
*   **Technology**: Go, deployed on Cloud Run.
*   **Trigger**: Invoked by Google Cloud Scheduler daily at 00:05 UTC (`5 0 * * *`).
*   **Timezone**: Uses Australia/Canberra timezone for determining day boundaries.
*   **Responsibilities**:
    1.  Receives a trigger to archive a specific day's data (e.g., yesterday).
    2.  Queries Firestore for all alerts published within that day's 24-hour UTC window.
    3.  Serializes the alerts into a JSONL file.
    4.  Uploads the resulting file to a Google Cloud Storage bucket for long-term, cost-effective storage.
    5.  (Future enhancement): Deletes the archived alerts from Firestore to reduce its size and cost.

### 3.4. Data Analysis Dashboard
*   **Technology**: Vanilla JavaScript (ES6), HTML5, CSS3, hosted on Firebase Hosting.
*   **Responsibilities**:
    1.  Authenticates users via Firebase Anonymous Authentication and manages ID token refresh.
    2.  Provides a user interface for selecting dates and applying filters.
    3.  Constructs and sends authenticated API requests to the `alerts-service` with Bearer tokens.
    4.  Parses the streamed JSONL response and loads the data into memory.
    5.  Performs client-side filtering, sorting, and statistical calculations.
    6.  Renders the filtered alerts on an interactive map (Leaflet.js with Timeline plugin for temporal visualization) and in a detailed list.

### 3.5. Supporting Infrastructure
*   **Artifact Registry**: Docker container images for all Cloud Run services are stored in Google Artifact Registry, enabling version control and efficient deployment workflows.
*   **BigQuery Dataset**: Provisioned for potential future analytics and data warehouse capabilities, allowing for advanced querying and business intelligence on historical alert data.

---

## 4. Core Technologies & Rationale

*   **Go**: Chosen for the backend services due to its high performance, low memory footprint (ideal for serverless environments), strong typing, and excellent support for concurrency, which is useful for handling multiple simultaneous requests.

*   **Google Cloud Run**: Selected as the compute platform for its serverless nature. It offers scale-to-zero capabilities, which is extremely cost-effective for services that are invoked periodically. It also provides a fully managed environment, simplifying deployment and operations.

*   **Google Cloud Firestore**: Used as the primary database for its ease of use, scalability, and real-time capabilities. Its document-based model is a good fit for the semi-structured nature of the alert data.

*   **Google Cloud Storage (GCS)**: Chosen for long-term archival storage due to its low cost and high durability. It is ideal for storing the immutable daily alert archives.

*   **Vanilla JavaScript**: Selected for the frontend to create a lightweight, fast, and dependency-free application. This approach avoids the need for a complex build pipeline and demonstrates strong foundational web development skills.

*   **GitHub Actions**: Used for CI/CD to automate the process of building, testing, and deploying the Go microservices to Cloud Run whenever code is pushed to the `main` branch.

*   **Terraform**: Adopted for Infrastructure as Code (IaC) to manage all GCP resources declaratively. Enables version control, reproducibility, and automated infrastructure deployments with proper state management.

*   **Firebase Authentication**: Implemented Anonymous Authentication to protect the API endpoints from abuse while maintaining a frictionless user experience. Combined with per-user rate limiting for cost control and service protection.

---

## 5. Data Flow Lifecycle

1.  **Collection**: A Cloud Scheduler job triggers the `scraper-service`. The service fetches raw data from Waze, filters for police alerts, and stores them in Firestore.
2.  **Archival**: A separate daily Cloud Scheduler job triggers the `archive-service`. It reads the previous day's data from Firestore and writes it as a permanent `.jsonl` file to a GCS bucket.
3.  **Retrieval**: A user visits the dashboard and selects dates. The dashboard calls the `alerts-service` API.
4.  **Serving**: The `alerts-service` intelligently serves the data, preferring the cheap and fast GCS archives when available and falling back to live Firestore queries for the most recent data.
5.  **Visualization**: The dashboard receives the data stream and renders it for the user to analyze.
