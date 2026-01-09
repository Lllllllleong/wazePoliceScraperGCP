# Architectural Decision Record (ADR)

This document records the key architectural decisions made during the development of the Waze Police Scraper project. Each record describes a significant decision, its context, and its consequences, showcasing the engineering thought process behind the application's evolution.

---

## ADR-001: Implement a Backend-for-Frontend (BFF) API



**Status**: Implemented

### Context

The initial prototype of the dashboard connected directly from the browser to Google Cloud Firestore. This architecture was simple but had significant drawbacks, including exposing GCP credentials on the client-side, requiring complex query logic in JavaScript, and fetching more data than necessary.

### Decision

We implemented a dedicated backend-for-frontend (BFF) service, the **`alerts-service`**, written in Go. The frontend dashboard now communicates exclusively with this secure API endpoint, which acts as a proxy and business logic layer for all database interactions.

### Consequences

*   **Positive**:
    *   **Enhanced Security**: Firestore credentials are no longer exposed to the browser. The API enforces a clear security boundary.
    *   **Improved Performance**: The API handles complex, multi-date queries and server-side deduplication, sending a minimal, clean dataset to the client. This drastically reduces data transfer and improves frontend rendering speed.
    *   **Centralized Logic**: All data access rules, filtering logic, and data shaping are centralized in the Go service, making the system easier to maintain, test, and evolve.
    *   **Scalability**: The API layer provides a natural point for future enhancements like response caching, rate limiting, and advanced authentication.

*   **Negative**:
    *   Introduced an additional microservice to build, deploy, and maintain, increasing the overall architectural complexity.

---

## ADR-002: Standardize on Explicit UTC for All Time Operations



**Status**: Implemented

### Context

The application suffered from critical timezone-related bugs. Users in different timezones (e.g., Canberra, UTC+11) would select a date like "October 20th" but see data from the wrong 24-hour period. This was caused by the JavaScript frontend defaulting to the user's local timezone while the Go backend operated in UTC, creating a fundamental mismatch.

### Decision

We enforced a strict, explicit UTC-first policy across the entire application stack:
1.  **Data Storage**: All timestamps in Firestore are stored in UTC.
2.  **Backend Logic**: The Go backend exclusively uses `time.UTC` for all date parsing, queries, and calculations.
3.  **Frontend Logic**: The JavaScript frontend is responsible for converting all user-selected dates into explicit UTC timestamps (e.g., `2025-10-20T00:00:00.000Z`) before sending them to the API.

### Consequences

*   **Positive**:
    *   **Eliminated Timezone Bugs**: All date-related inconsistencies were resolved. A user selecting "October 20th" now sees data for that entire calendar day in UTC, regardless of their location.
    *   **Global Consistency**: The system now behaves predictably and reliably for users in any timezone.
    *   **Simplified Backend Logic**: The backend no longer needs to be aware of client timezones, simplifying its design.

*   **Negative**:
    *   Places the responsibility of correct timezone handling on the client-side, requiring careful implementation in the JavaScript code.

---

## ADR-003: Implement an On-Demand, Multi-Date Loading Strategy



**Status**: Implemented

### Context

The initial version of the dashboard loaded the entire Firestore collection on page load. This was slow, inefficient, and did not scale. The UI then evolved to a simple start/end date filter, which was still inefficient for querying non-contiguous days.

### Decision

We refactored the data loading mechanism to be on-demand and user-driven:
1.  **Multi-Date Picker**: Implemented a calendar UI (Flatpickr) allowing users to select multiple, non-contiguous dates.
2.  **On-Demand Fetching**: Data is only fetched when the user explicitly clicks a "Load Data" button.
3.  **Efficient Batching**: The frontend sends a single API request containing all selected dates. The backend then queries each day in parallel.
4.  **Deduplication**: A `Map` data structure is used (with the alert's UUID as the key) to ensure that alerts active across multiple selected days appear only once.

### Consequences

*   **Positive**:
    *   **Massive Performance Improvement**: The initial page load is now instant. Users only pay the data-loading cost for the specific days they are interested in.
    *   **Network Efficiency**: Queries are targeted, and the total data transferred is significantly reduced.
    *   **Enhanced User Experience**: The staged UI (select dates, then load) provides a clear, guided workflow. The multi-date picker is more flexible than a simple date range.

*   **Negative**:
    *   Users must now take an explicit action ("Load Data") to see information, slightly increasing the number of clicks required.

---

## ADR-004: Adopt a High-Fidelity Timeline Visualization



**Status**: Implemented

### Context

A key goal was to visualize the temporal nature of alerts. An initial attempt used a library (`TimestampedGeoJson`) that assigned a *fixed, global duration* (e.g., 10 minutes) to every alert on the timeline, regardless of its actual lifespan. This was misleading and failed to represent the data accurately.

### Decision

We replaced the initial implementation with Leaflet.js combined with the `Leaflet.Timeline` plugin and `TimelineSlider` control. The data processing was updated so that each alert is represented on the timeline with its **actual, individual lifespan**, derived from its `publish_time` and `expire_time`.

### Consequences

*   **Positive**:
    *   **Drastic Improvement in Temporal Accuracy**: The visualization is now a true and accurate representation of reality. An alert that was active for 4 hours now appears on the timeline for 4 hours, not 10 minutes.
    *   **Richer Analysis**: Enables genuine insights into alert durations, overlaps, and patterns that were previously impossible to see.
    *   **Superior User Experience**: The `TimelineSlider` provides smoother scrubbing, better controls, and more detailed popups, offering a more professional and interactive experience.

*   **Negative**:
    *   Required more complex data preparation to create the GeoJSON features with correct start and end times for each alert.

---

## ADR-005: Evolve to a Unified, Tag-Based Filtering System



**Status**: Implemented

### Context

The dashboard's filtering capabilities evolved significantly. They began as simple, hardcoded dropdowns, became fragmented across different parts of the UI, and lacked consistency.

### Decision

We executed a series of refactors to create a single, powerful, and reusable filtering system:
1.  **From Dropdowns to Tags**: Replaced single-choice dropdowns with a multi-select, tag-based UI. Users can now select multiple filter criteria (e.g., multiple subtypes or streets) and see them as removable tags.
2.  **Dynamic Population**: Filter options are now dynamically populated based on the data loaded, ensuring that only relevant choices are presented.
3.  **Default to Quality**: A "Verified Only" filter (alerts with >0 thumbs up) was added and enabled by default to immediately show users the most reliable data.
4.  **Componentization**: The entire filtering UI was encapsulated into a single, reusable, and expandable component. This component was then integrated into all relevant analysis pages, providing a consistent experience everywhere.

### Consequences

*   **Positive**:
    *   **Consistent and Intuitive UX**: Users learn the powerful tag-based interface once and can use it across the entire application.
    *   **Increased Filtering Power**: The multi-select capability allows for more complex and useful data exploration.
    *   **Improved Code Quality**: The component-based architecture is more maintainable, reusable, and follows modern frontend design principles.
    *   **Better User Guidance**: Defaulting to "Verified" alerts and providing dynamic options improves the out-of-the-box experience.

*   **Negative**:
    *   The initial implementation of the tag-based UI was more complex than simple dropdowns.

---
## ADR-006: Streaming JSONL Response Strategy

**Status**: Implemented

### Context

As the dataset grew, the `alerts-service` began to struggle with memory usage and latency. Loading a week's worth of data meant fetching thousands of documents from Firestore, unmarshaling them into Go structs, marshaling them back into a giant JSON array, and sending it to the client. This caused:
*   **High Memory Spikes**: The service had to hold the entire response in memory.
*   **Slow TTFB**: The client received nothing until the entire payload was ready.
*   **Browser Freeze**: Parsing a massive JSON blob blocked the browser's main thread.

### Decision

We re-architected the API to use **Newline Delimited JSON (JSONL) Streaming**:
1.  **Streaming**: The backend now writes each alert to the response writer as soon as it is processed, separated by a newline. It does not buffer the whole response.
2.  **Client-Side Parsing**: The frontend uses the `NDJSON` format to parse the incoming stream line-by-line.

### Consequences

*   **Positive**:
    *   **Low Memory Footprint**: The backend memory usage is now constant, regardless of the response size.
    *   **Immediate Feedback**: The client starts receiving and rendering data immediately, improving perceived performance.
    *   **Scalability**: The system can now handle arbitrarily large time ranges without crashing.

*   **Negative**:
    *   **Frontend Complexity**: The frontend code had to be updated to handle stream reading and NDJSON parsing, which is more complex than a simple `fetch().json()`.

---

## ADR-007: Enable GZIP Compression

**Status**: Implemented

### Context

While the streaming architecture solved memory and latency issues, the raw JSONL data was still voluminous. Transferring large datasets (e.g., multiple weeks of alerts) consumed significant bandwidth, which could be costly and slow for users on mobile networks.

### Decision

We implemented **GZIP compression middleware** in the `alerts-service`. The server transparently compresses the streaming response if the client's `Accept-Encoding` header includes `gzip`.

### Consequences

*   **Positive**:
    *   **Massive Bandwidth Reduction**: Compression reduced the payload size by approximately **88%**.
    *   **Faster Downloads**: Smaller payloads mean faster data transfer times, especially on slower networks.
    *   **Cost Savings**: Reduced egress traffic from Cloud Run lowers infrastructure costs.

*   **Negative**:
    *   **CPU Overhead**: Compression adds a small amount of CPU overhead to the backend, but this is negligible compared to the bandwidth savings.

---

## ADR-008: Infrastructure as Code with Terraform

**Status**: Implemented

### Context

The project's infrastructure (Cloud Run services, BigQuery datasets, Scheduler jobs, IAM roles) was initially created and managed manually via the Google Cloud Console. This "ClickOps" approach led to several issues:
*   **Configuration Drift**: It was difficult to track changes or ensure that the production environment matched the intended configuration.
*   **Lack of Reproducibility**: Recreating the environment (e.g., for testing or disaster recovery) was a manual, error-prone process.
*   **Security Risks**: IAM permissions were often granted ad-hoc, leading to overprivileged service accounts that were hard to audit.

### Decision

We migrated the entire production infrastructure to **Terraform**. All GCP resources are now defined as code in the `terraform/` directory, organized into reusable modules (e.g., `cloud-run`, `firestore`, `iam`). State is managed remotely in a GCS bucket with locking and versioning enabled.

### Consequences

*   **Positive**:
    *   **Reproducibility**: The entire environment can be provisioned or destroyed with a single command (`terraform apply` / `terraform destroy`).
    *   **Version Control**: Infrastructure changes are now tracked in Git, enabling code reviews, history, and rollbacks for infrastructure just like application code.
    *   **Automated Deployments**: Terraform is integrated into the CI/CD pipeline, ensuring that infrastructure changes are automatically validated and applied.
    *   **Drift Detection**: We can easily detect if the actual cloud resources have deviated from the defined configuration.

*   **Negative**:
    *   **Learning Curve**: The team needed to learn HCL (HashiCorp Configuration Language) and Terraform concepts.
    *   **State Management**: Managing the Terraform state file requires care to avoid corruption or secrets leakage (though remote state helps mitigate this).

---

## ADR-010: Implement Firebase Anonymous Authentication with Per-User Rate Limiting

**Status**: Implemented

### Context

The `alerts-service` API was initially a completely public endpoint with no authentication. This posed several risks:
*   **Cost Risk**: Malicious actors or automated bots could overwhelm the service with requests, leading to unexpected Cloud Run and Firestore costs.
*   **Resource Exhaustion**: Excessive traffic could degrade service performance for legitimate users.
*   **No User Tracking**: It was impossible to identify individual users for rate limiting or analytics purposes.

Traditional authentication (username/password, OAuth) was too complex and would create friction for what is intended to be a publicly accessible educational demo.

### Decision

We implemented **Firebase Anonymous Authentication** combined with **per-user rate limiting**:

1.  **Frontend**: The dashboard automatically signs in users anonymously using Firebase Auth SDK, obtaining a short-lived ID token.
2.  **Token Management**: The frontend manages automatic token refresh every 50 minutes to maintain valid credentials.
3.  **Backend Authentication**: The `alerts-service` validates the Firebase ID token on every request using the Firebase Admin SDK.
4.  **Per-User Rate Limiting**: Each authenticated user (identified by UID) is assigned their own rate limiter. The default limit is 30 requests per minute, configurable via environment variable (`RATE_LIMIT_PER_MINUTE`).
5.  **Automatic Cleanup**: Old rate limiters are periodically cleaned up to prevent memory leaks.

### Consequences

*   **Positive**:
    *   **Zero User Friction**: Users are authenticated transparently in the background with no registration, login forms, or credentials required.
    *   **Cost Protection**: Rate limiting prevents abuse and controls Cloud Run scaling costs.
    *   **Fair Usage**: Each user gets their own rate limit, ensuring fair access to the service.
    *   **User Identification**: Firebase provides stable UIDs for analytics and abuse monitoring without collecting personal information.
    *   **Production-Ready Security**: Demonstrates real-world API protection patterns suitable for production applications.

*   **Negative**:
    *   **Dependency**: Adds Firebase Auth as a critical dependency. Service outages would affect API access.
    *   **Token Management Complexity**: Frontend must handle token refresh, expiration, and retry logic.
    *   **Development Overhead**: Requires Firebase Auth Emulator setup for local development and testing.

---

## ADR-009: Least Privilege IAM Strategy

**Status**: Implemented

### Context

An internal security audit revealed that the application was running with the default Compute Engine service account, which possesses the `roles/editor` permission. This granted the application full administrative access to the entire GCP project, a significant violation of the Principle of Least Privilege. A compromise of any service could have led to a total project takeover.

### Decision

We implemented a strict **Service Identity Segmentation** strategy. We replaced the single default identity with three dedicated, task-specific service accounts:

1.  **`scraper-sa`**: Granted **Write-Only** access to Firestore. It cannot read data or access GCS.
2.  **`alerts-sa`**: Granted **Read-Only** access to Firestore and GCS (for reading archives). It cannot write data.
3.  **`archive-sa`**: Granted specific permissions to read from Firestore and read/write to the GCS archive bucket.

### Consequences

*   **Positive**:
    *   **Reduced Blast Radius**: If one service is compromised (e.g., the Scraper), the attacker cannot access data from other services or modify infrastructure.
    *   **Enhanced Security Posture**: The application now operates with the absolute minimum permissions required to function.
    *   **Auditability**: It is now clear which service is performing which actions in the audit logs.

*   **Negative**:
    *   **Complexity**: Managing multiple service accounts and their specific IAM bindings is more complex than using a single default account.
