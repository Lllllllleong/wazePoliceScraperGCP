# Migration to API-Based Architecture

This document outlines the changes made to migrate from direct Firestore access to an API-based architecture.

## Overview

The frontend now communicates with a Go-based API service instead of directly querying Firestore. This provides:

- **Better security**: Firestore credentials not exposed to frontend
- **Server-side filtering**: More efficient queries
- **Centralized logic**: Data access logic in one place
- **Scalability**: API can handle caching, rate limiting, etc.

## Files Created

### Backend (Go)

1. **`cmd/api/main.go`**
   - New HTTP API service
   - Handles POST requests to `/api/alerts`
   - Includes CORS middleware
   - Queries Firestore and returns JSON responses

2. **`internal/models/api.go`**
   - `AlertsRequest` - Request body structure
   - `AlertsResponse` - Response structure with alerts and stats
   - `ResponseStats` - Metadata about the query

3. **`internal/storage/police_alerts.go`** (updated)
   - Added `GetPoliceAlertsByDatesWithFilters()` method
   - Queries alerts for multiple dates with optional filters
   - Server-side deduplication by UUID

4. **`Dockerfile.api`**
   - Multi-stage Docker build for API service
   - Builds from `cmd/api/main.go`
   - Alpine-based final image

5. **`scripts/deploy-api.sh`** & **`scripts/deploy-api.bat`**
   - Deployment scripts for Cloud Run
   - Sets environment variables
   - Outputs service URL for frontend configuration

6. **`docs/API.md`**
   - Complete API documentation
   - Request/response examples
   - Deployment and testing instructions

## Files Modified

### Frontend (JavaScript)

1. **`dataAnalysis/public/config.js`**
   - Added `API_CONFIG` object with endpoint URL and settings
   - `useAPI` flag to toggle between API and legacy Firestore

2. **`dataAnalysis/public/app.js`**
   - Refactored `loadAlertsForSelectedDates()` to support both modes
   - New `loadAlertsFromAPI()` - Fetches from API endpoint
   - New `parseTimestamp()` - Handles various timestamp formats
   - Legacy `loadAlertsFromFirestore()` - Kept for backward compatibility
   - API mode is now the default when configured

## API Endpoint

### Request
```http
POST /api/alerts
Content-Type: application/json

{
  "dates": ["2025-10-20", "2025-10-21"],
  "subtypes": [],  // Optional
  "streets": []    // Optional
}
```

### Response
```json
{
  "success": true,
  "message": "Successfully retrieved 150 alerts",
  "alerts": [...],
  "stats": {
    "total_alerts": 150,
    "dates_queried": ["2025-10-20", "2025-10-21"]
  }
}
```

## Deployment Steps

### 1. Deploy the API Service

**Using script:**
```bash
# Linux/Mac
chmod +x scripts/deploy-api.sh
./scripts/deploy-api.sh

# Windows
scripts\deploy-api.bat
```

**Manual:**
```bash
gcloud run deploy waze-alerts-api \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars GCP_PROJECT_ID=your-project-id,FIRESTORE_COLLECTION=police_alerts \
  --dockerfile Dockerfile.api
```

### 2. Update Frontend Configuration

After deployment, get the service URL and update `dataAnalysis/public/config.js`:

```javascript
window.API_CONFIG = {
    alertsEndpoint: "https://waze-alerts-api-xxxxx.run.app/api/alerts",
    useAPI: true,
    timeout: 30000
};
```

### 3. Deploy Frontend (if needed)

```bash
cd dataAnalysis
firebase deploy --only hosting
```

## Testing

### Test API locally:
```bash
# Start the API
export GCP_PROJECT_ID="your-project-id"
go run cmd/api/main.go

# Test endpoint
curl -X POST http://localhost:8080/api/alerts \
  -H 'Content-Type: application/json' \
  -d '{"dates": ["2025-10-20"]}'
```

### Test frontend locally:
1. Update `config.js` with local API URL: `http://localhost:8080/api/alerts`
2. Serve frontend: `cd dataAnalysis && firebase serve`
3. Open browser and test date selection/loading

## Migration Path

The system supports both modes simultaneously:

1. **API Mode** (recommended): Set `useAPI: true` in config.js
   - Frontend → API → Firestore
   - More secure, better performance

2. **Legacy Mode**: Set `useAPI: false` in config.js
   - Frontend → Firestore (direct)
   - For backward compatibility or testing

## Security Considerations

Current setup allows unauthenticated access. For production:

1. **Add authentication**:
   - API keys
   - Firebase Authentication tokens
   - OAuth 2.0

2. **Rate limiting**:
   - Cloud Armor
   - API Gateway

3. **CORS restrictions**:
   - Limit to specific frontend domain

## Performance Improvements

With API-based architecture:

- ✅ Single request for multiple dates (vs. N requests)
- ✅ Server-side deduplication (less data transferred)
- ✅ Future caching opportunities
- ✅ Better error handling and retries
- ✅ Request/response logging

## Backward Compatibility

The `loadAlertsFromFirestore()` function is preserved:
- Existing deployments continue working
- Easy rollback if needed
- Toggle via `API_CONFIG.useAPI` flag

## Next Steps (Optional)

1. Add caching layer (Redis/Memcache)
2. Implement authentication
3. Add rate limiting
4. Create aggregation endpoints (stats, summaries)
5. Add WebSocket support for real-time updates
6. Implement pagination for large datasets

## Rollback Plan

If issues occur:

1. Set `useAPI: false` in `config.js`
2. Redeploy frontend
3. Frontend falls back to direct Firestore access
4. No backend changes needed

## Files Changed Summary

```
New Files:
  cmd/api/main.go
  internal/models/api.go
  docs/API.md
  Dockerfile.api
  scripts/deploy-api.sh
  scripts/deploy-api.bat

Modified Files:
  internal/storage/police_alerts.go
  dataAnalysis/public/config.js
  dataAnalysis/public/app.js
```
