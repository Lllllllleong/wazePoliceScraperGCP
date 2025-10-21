# Waze Police Alerts API

This is a Go-based HTTP API service that provides access to police alert data stored in Firestore. It's designed to be deployed as a Google Cloud Run service and serves as a backend for the frontend dashboard.

## Architecture

- **Runtime**: Go 1.21
- **Platform**: Google Cloud Run (serverless)
- **Database**: Google Cloud Firestore
- **Deployment**: Docker container

## API Endpoints

### `POST /api/alerts`

Retrieves police alerts for specified dates with optional filters.

**Request Body:**
```json
{
  "dates": ["2025-10-20", "2025-10-21"],
  "subtypes": ["POLICE_VISIBLE", "POLICE_HIDING"],  // Optional
  "streets": ["Main Street", "Highway 1"]           // Optional
}
```

**Response:**
```json
{
  "success": true,
  "message": "Successfully retrieved 150 alerts",
  "alerts": [
    {
      "uuid": "abc123...",
      "type": "POLICE",
      "subtype": "POLICE_VISIBLE",
      "street": "Main Street",
      "city": "Sydney",
      "country": "AU",
      "location_geo": {
        "latitude": -33.8688,
        "longitude": 151.2093
      },
      "publish_time": "2025-10-20T10:30:00Z",
      "expire_time": "2025-10-20T12:45:00Z",
      "active_millis": 8100000,
      // ... other fields
    }
  ],
  "stats": {
    "total_alerts": 150,
    "dates_queried": ["2025-10-20", "2025-10-21"],
    "subtypes_filtered": ["POLICE_VISIBLE"],
    "streets_filtered": []
  }
}
```

### `GET /health`

Health check endpoint for monitoring.

**Response:** `OK` (200 status)

## Environment Variables

- `GCP_PROJECT_ID` - Google Cloud Project ID (required)
- `FIRESTORE_COLLECTION` - Firestore collection name (default: "police_alerts")
- `PORT` - Server port (default: 8080, Cloud Run sets this automatically)

## Local Development

1. **Set environment variables:**
   ```bash
   export GCP_PROJECT_ID="your-project-id"
   export FIRESTORE_COLLECTION="police_alerts"
   ```

2. **Run locally:**
   ```bash
   go run cmd/api/main.go
   ```

3. **Test the API:**
   ```bash
   curl -X POST http://localhost:8080/api/alerts \
     -H 'Content-Type: application/json' \
     -d '{"dates": ["2025-10-20"]}'
   ```

## Deployment

### Using the deployment script (recommended):

**Linux/Mac:**
```bash
chmod +x scripts/deploy-api.sh
./scripts/deploy-api.sh
```

**Windows:**
```bash
scripts\deploy-api.bat
```

### Manual deployment:

```bash
gcloud run deploy waze-alerts-api \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars GCP_PROJECT_ID=your-project-id,FIRESTORE_COLLECTION=police_alerts \
  --dockerfile Dockerfile.api
```

## Frontend Integration

After deployment, update `dataAnalysis/public/config.js`:

```javascript
window.API_CONFIG = {
    alertsEndpoint: "https://your-api-url.run.app/api/alerts",
    useAPI: true,
    timeout: 30000
};
```

## CORS Support

The API includes CORS middleware that allows cross-origin requests from any domain. This is necessary for the frontend dashboard hosted on Firebase to access the API.

## Code Structure

```
cmd/api/main.go              # Main API server
internal/models/api.go       # Request/response models
internal/storage/
  ├── firestore.go           # Firestore client
  └── police_alerts.go       # Alert query methods
```

## Features

- ✅ Multi-date querying with deduplication
- ✅ Optional subtype and street filtering
- ✅ CORS support for frontend access
- ✅ Comprehensive error handling
- ✅ Health check endpoint
- ✅ Logging and statistics
- ✅ Serverless scaling with Cloud Run

## Performance Considerations

- Queries are optimized to fetch all requested dates in parallel (server-side deduplication)
- Results are deduplicated by UUID across multiple date queries
- Firestore indexes may be required for optimal performance:
  - Composite index on: `expire_time ASC, publish_time ASC`

## Security

- Currently allows unauthenticated access (`--allow-unauthenticated`)
- For production, consider:
  - Adding API key authentication
  - Rate limiting
  - IP allowlisting
  - Firebase Authentication integration

## Monitoring

Monitor the service in Google Cloud Console:
- **Logs**: Cloud Run → waze-alerts-api → Logs
- **Metrics**: Request count, latency, error rate
- **Health**: Use `/health` endpoint for uptime monitoring
