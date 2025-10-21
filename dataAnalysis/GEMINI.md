# Project Overview

This project is a web-based dashboard for analyzing and visualizing police alert data. It displays alerts on a map and allows users to filter and explore the data in near real-time.

## Technologies Used

*   **Frontend:** HTML, CSS, JavaScript
*   **Backend:** Firebase (Firestore)
*   **Mapping:** Leaflet.js
*   **Date Picker:** Flatpickr

## Architecture

The application is a single-page application (SPA) that fetches data from a Firestore database. The frontend is built with vanilla JavaScript and uses Leaflet.js to render alert data on an interactive map. Firebase is used for data storage and hosting.

# Building and Running

## Prerequisites

*   Node.js and npm
*   Firebase CLI

## Installation

1.  Install dependencies:
    ```bash
    npm install
    ```

## Running Locally

1.  Start the local development server:
    ```bash
    npm run serve
    ```
    This will start a local server using `firebase serve`.

## Deployment

1.  Deploy to Firebase Hosting:
    ```bash
    npm run deploy
    ```

# Development Conventions

*   **Code Style:** The project follows standard JavaScript conventions.
*   **Data:** Alert data is stored in a Firestore collection. The schema includes fields like `uuid`, `type`, `subtype`, `street`, `location_geo`, `publish_time`, and `expire_time`.
*   **Configuration:** Firebase configuration is stored in `firebase.json`. Frontend configuration, such as API keys, should be in `public/config.js`.
