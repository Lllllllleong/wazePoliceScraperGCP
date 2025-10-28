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

We replaced the initial implementation with a more advanced combination of `folium.plugins.Timeline` and `TimelineSlider`. The data processing was updated so that each alert is represented on the timeline with its **actual, individual lifespan**, derived from its `publish_time` and `expire_time`.

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

## ADR-006: Prototype a Modern Frontend with React/TypeScript



**Status**: Prototyped

### Context

While the primary dashboard is built with vanilla JavaScript for simplicity and zero-build deployment, it is important to demonstrate proficiency with modern, industry-standard frontend frameworks.

### Decision

In parallel with maintaining the vanilla JavaScript application, we developed a complete, high-fidelity prototype of the dashboard using **React, TypeScript, and Vite**. This was treated as a forward-looking exploration to evaluate a potential future migration and to showcase modern development capabilities.

### Consequences

*   **Positive**:
    *   **Demonstrates Modern Skillset**: Shows expertise in React's hook-based architecture, TypeScript's static typing, and modern build tooling.
    *   **Improved Architecture**: The React prototype established a clean, component-based architecture with custom hooks (`useAlertLoader`, `useAlertFilters`) and a clear separation of concerns.
    *   **Future-Ready**: Provides a clear, production-ready migration path for the project if it were to grow in complexity.
    *   **Problem Solving**: The process involved solving framework-specific challenges, such as replacing an incompatible date-picker library, further demonstrating adaptability.

*   **Negative**:
    *   Created a separate codebase to maintain in parallel with the primary vanilla JavaScript application.