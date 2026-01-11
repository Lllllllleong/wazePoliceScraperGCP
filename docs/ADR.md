# Architectural Decision Record (ADR)

This document records the key architectural decisions made during the development of the Waze Police Scraper project. Each record describes a significant decision, its context, and its consequences.

---

## ADR-001: Implement a Backend-for-Frontend (BFF) API

**Status**: Implemented

### Context

The initial prototype of the dashboard connected directly from the browser to Google Cloud Firestore. This architecture was simple but had significant drawbacks, including exposing GCP credentials on the client-side, requiring complex query logic in JavaScript, and fetching more data than necessary.

### Decision

We implemented a dedicated backend-for-frontend (BFF) service, the **`alerts-service`**, written in Go. The frontend dashboard now communicates exclusively with this secure API endpoint, which acts as a proxy and business logic layer for all database interactions.

### Consequences

*   **Positive**:
    *   **Security**: Firestore credentials are no longer exposed to the browser. The API enforces a clear security boundary.
    *   **Performance**: The API handles complex, multi-date queries and server-side deduplication, sending a minimal dataset to the client. This reduces data transfer and improves frontend rendering speed.
    *   **Centralized Logic**: All data access rules, filtering logic, and data shaping are centralized in the Go service, making the system easier to maintain, test, and evolve.
    *   **Scalability**: The API layer provides a natural point for future enhancements like response caching, rate limiting, and advanced authentication.

*   **Negative**:
    *   Introduced an additional microservice to build, deploy, and maintain, increasing the overall architectural complexity.

---

## ADR-002: Standardize on Explicit UTC for All Time Operations

**Status**: Implemented

### Context

The application had timezone-related bugs. Users in different timezones (e.g., Canberra, UTC+11) would select a date like "October 20th" but see data from the wrong 24-hour period. This was caused by the JavaScript frontend defaulting to the user's local timezone while the Go backend operated in UTC, creating a mismatch.

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
    *   **Performance Improvement**: The initial page load is now instant. Users only pay the data-loading cost for the specific days they are interested in.
    *   **Network Efficiency**: Queries are targeted, and the total data transferred is reduced.
    *   **User Experience**: The staged UI (select dates, then load) provides a clear, guided workflow. The multi-date picker is more flexible than a simple date range.

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
    *   **Temporal Accuracy**: The visualization is now an accurate representation of reality. An alert that was active for 4 hours now appears on the timeline for 4 hours, not 10 minutes.
    *   **Analysis**: Enables insights into alert durations, overlaps, and patterns that were previously impossible to see.
    *   **User Experience**: The `TimelineSlider` provides smoother scrubbing, better controls, and more detailed popups.

*   **Negative**:
    *   Required more complex data preparation to create the GeoJSON features with correct start and end times for each alert.

---

## ADR-005: Evolve to a Unified, Tag-Based Filtering System

**Status**: Implemented

### Context

The dashboard's filtering capabilities evolved over time. They began as simple, hardcoded dropdowns, became fragmented across different parts of the UI, and lacked consistency.

### Decision

We executed a series of refactors to create a single, reusable filtering system:
1.  **From Dropdowns to Tags**: Replaced single-choice dropdowns with a multi-select, tag-based UI. Users can now select multiple filter criteria (e.g., multiple subtypes or streets) and see them as removable tags.
2.  **Dynamic Population**: Filter options are now dynamically populated based on the data loaded, ensuring that only relevant choices are presented.
3.  **Default to Quality**: A "Verified Only" filter (alerts with >0 thumbs up) was added and enabled by default to immediately show users the most reliable data.
4.  **Componentization**: The entire filtering UI was encapsulated into a single, reusable, and expandable component. This component was then integrated into all relevant analysis pages, providing a consistent experience everywhere.

### Consequences

*   **Positive**:
    *   **Consistent UX**: Users learn the tag-based interface once and can use it across the entire application.
    *   **Filtering Capability**: The multi-select capability allows for more complex and useful data exploration.
    *   **Code Maintainability**: The component-based architecture is more maintainable and reusable.
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
*   **Browser Freeze**: Parsing a large JSON blob blocked the browser's main thread.

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
    *   **Bandwidth Reduction**: Compression reduced the payload size by approximately **88%**.
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
    *   **Drift Detection**: Can detect if the actual cloud resources have deviated from the defined configuration.

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

*   **Negative**:
    *   **Dependency**: Adds Firebase Auth as a critical dependency. Service outages would affect API access.
    *   **Token Management Complexity**: Frontend must handle token refresh, expiration, and retry logic.
    *   **Development Overhead**: Requires Firebase Auth Emulator setup for local development and testing.

---

## ADR-011: Least Privilege IAM Strategy

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
    *   **Minimum Permissions**: The application now operates with the minimum permissions required to function.
    *   **Auditability**: It is now clear which service is performing which actions in the audit logs.

*   **Negative**:
    *   **Complexity**: Managing multiple service accounts and their specific IAM bindings is more complex than using a single default account.

---

## ADR-012: Adopt Interface-Based Dependency Injection for Testability

**Status**: Implemented (January 2026)

### Context

The initial implementation of backend services created dependencies (Firestore clients, Waze API clients, GCS clients) directly inside handler functions. While this approach was simple and worked for production, it created significant testing challenges:

*   **Untestable Handlers**: Handler tests could only verify that functions were created, not their actual behavior.
*   **External Service Dependencies**: Running tests required real Firestore, GCS, and Waze API access, making tests slow, brittle, and expensive.
*   **Low Coverage**: Test coverage was limited to ~25% overall because most business logic was untestable.
*   **No Error Path Testing**: Impossible to simulate error conditions (API failures, database errors) without breaking production services.

The project had architectural foundations (separate packages for `storage`, `waze`, `models`) but lacked the final abstraction layer needed for testing.

### Decision

We implemented **interface-based dependency injection** across all backend services:

1.  **Define Clear Interfaces**: Created `AlertStore`, `AlertFetcher`, `GCSClient`, and `FirebaseAuthClient` interfaces in respective packages that define all operations needed by handlers.

2.  **Refactor for Injection**: Modified all handlers to accept interface parameters instead of creating concrete clients internally:
    ```go
    // Before: Creates clients inside
    func makeHandler(config Config) http.HandlerFunc {
        client := waze.NewClient()  // Hard dependency
        // ...
    }
    
    // After: Accepts injected interfaces
    func makeHandler(fetcher waze.AlertFetcher, store storage.AlertStore) http.HandlerFunc {
        // Uses injected dependencies
    }
    ```

3.  **Production vs Test Wiring**: In `main()`, inject real implementations. In tests, inject mocks:
    ```go
    // main.go - Production
    wazeClient := waze.NewClient()
    firestoreClient := storage.NewFirestoreClient(...)
    handler := makeHandler(wazeClient, firestoreClient)
    
    // handler_test.go - Testing
    mockFetcher := &waze.MockAlertFetcher{...}
    mockStore := &storage.MockAlertStore{...}
    handler := makeHandler(mockFetcher, mockStore)
    ```

4.  **Build Mock Implementations**: Created comprehensive mocks with configurable behavior:
    - `MockAlertStore` with call logging and customizable function responses
    - `MockAlertFetcher` that can simulate success, failures, and edge cases
    - `MockGCSClient` for archive operations testing

5.  **Server Struct Pattern**: For `alerts-service` and `archive-service`, encapsulated all dependencies in a `server` struct that implements `http.Handler`.

### Consequences

*   **Positive**:
    *   **Testability**: Coverage increased from ~25% to ~60% overall (alerts-service: 72%, archive-service: 67%, scraper-service: 47%).
    *   **Comprehensive Testing**: Can now test all code paths including error scenarios, edge cases, and concurrent operations.
    *   **Fast Tests**: Unit tests run in milliseconds without external service calls.
    *   **Maintainability**: Clear separation between interface contracts and implementations makes refactoring safer.
    *   **Documentation**: Interfaces serve as clear API contracts showing exactly what each component needs.

*   **Negative**:
    *   **Initial Complexity**: Required refactoring all handlers and creating mock implementations (~500 lines of mock code).
    *   **Boilerplate**: Each service now needs interface definitions, production implementations, and mock implementations.
    *   **Cognitive Load**: Developers must understand both the interface and implementation layers.

---

## ADR-013: Integrate Firestore Emulator Testing in CI/CD Pipeline

**Status**: Implemented (January 2026)

### Context

While unit tests with mocked dependencies provided good coverage of business logic, they couldn't verify:

*   **Database Query Correctness**: Did our Firestore queries actually return the right data?
*   **Data Transformation**: Were we correctly converting between `WazeAlert` and `PoliceAlert` models?
*   **Query Performance**: Would queries scale with realistic data volumes?
*   **Concurrent Operations**: How did Firestore handle simultaneous writes?
*   **Edge Cases**: Unicode strings, large batches, date range boundaries?

The `storage` package had only 8.4% unit test coverage because most functions directly interact with Firestore—mocking would test the mock, not the real database behavior. Integration tests existed locally but were:

*   **Not Running in CI**: Developers had to remember to run them manually with `go test -tags=integration`
*   **Easy to Skip**: No enforcement meant integration tests were often forgotten
*   **No Deployment Gate**: Broken database interactions could reach production

This created a gap: while handlers were tested with mocks, there was no automated verification that those mocks actually behaved like real Firestore.

### Decision

We implemented **automated integration testing with Firestore emulator in the CI/CD pipeline**:

1.  **Build Tag Separation**: Use Go build tags to separate integration tests from unit tests:
    ```go
    //go:build integration
    
    package storage
    
    func TestIntegration_SavePoliceAlerts(t *testing.T) {
        // Tests against real emulator
    }
    ```

2.  **Emulator Setup in CI**: Added Firestore emulator to all GitHub Actions workflows using Firebase Tools:
    ```yaml
    - name: Set up Node.js (for Firebase Tools)
      uses: actions/setup-node@v4
      with:
        node-version: '20'
    
    - name: Install Firebase Tools
      run: npm install -g firebase-tools
    
    - name: Start Firestore Emulator
      run: |
        firebase emulators:start --only firestore --project demo-test &
        # Wait for emulator to be ready
        for i in {1..30}; do
          if curl -s http://localhost:8080 > /dev/null 2>&1; then
            echo "Firestore emulator is ready!"
            exit 0
          fi
          sleep 1
        done
    
    - name: Run Integration Tests
      env:
        FIRESTORE_EMULATOR_HOST: localhost:8080
      run: go test -tags=integration -v ./internal/storage/...
    ```

3.  **Separate CI Job**: Created dedicated `integration-test` jobs that:
    - Run after unit tests pass
    - Start the Firestore emulator
    - Execute integration tests with the `-tags=integration` flag
    - Upload separate coverage reports to Codecov
    - Block deployment if integration tests fail

4.  **Comprehensive Test Suite**: Built integration tests covering:
    - CRUD operations (create, read, update, delete)
    - Complex queries (date ranges, multi-date with filters)
    - Concurrent operations (race conditions, simultaneous writes)
    - Large batch operations (500+ alerts)
    - Edge cases (Unicode street names, duplicate UUIDs)
    - Performance benchmarks (10,000 alert queries under 5s)

5.  **Local Development Support**: Updated documentation to guide developers on running integration tests locally:
    ```bash
    # Using Firebase Tools (recommended, matches CI/CD)
    npm install -g firebase-tools
    firebase emulators:start --only firestore
    
    # Alternative: Using gcloud SDK
    gcloud emulators firestore start --host-port=localhost:8080
    
    # Set environment variable and run tests
    export FIRESTORE_EMULATOR_HOST=localhost:8080
    go test -tags=integration -v ./internal/storage/...
    ```

### Consequences

*   **Positive**:
    *   **Database Verification**: Automated testing against actual Firestore behavior catches query bugs before deployment.
    *   **Coverage**: The 40% integration coverage complements unit tests, bringing total storage package coverage to ~48%.
    *   **Bug Detection**: Integration tests revealed issues with concurrent updates and Unicode handling that mocks missed.
    *   **No Production Dependencies**: Emulator tests run in isolation without GCP credentials or costs.
    *   **Deployment Safety**: Both unit and integration tests must pass before any service deploys.
    *   **Documentation**: Integration tests serve as executable documentation of Firestore usage patterns.
    *   **Fast Feedback**: Emulator starts in ~10 seconds; full integration suite runs in ~30 seconds.

*   **Negative**:
    *   **CI Time Increase**: Each service workflow now runs 20-40 seconds longer for emulator setup and integration tests.
    *   **Maintenance Overhead**: Integration tests require more maintenance than unit tests as they're sensitive to Firestore behavior changes.
    *   **Local Setup**: Developers must install and configure the Firestore emulator for local integration testing.
    *   **Test Data Management**: Integration tests require careful test data setup/teardown to avoid pollution.

---

**Last Updated**: January 11, 2026  
**Document Maintainer**: Project Team
