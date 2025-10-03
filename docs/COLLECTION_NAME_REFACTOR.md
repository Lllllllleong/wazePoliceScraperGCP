# Firestore Collection Name Refactor

## Overview
The Firestore collection name is now configurable via the `FIRESTORE_COLLECTION` environment variable instead of being hardcoded.

## Changes Made

### 1. Environment Configuration
**Files Modified:**
- `.env.example`
- `.env`

**New Variable:**
```bash
FIRESTORE_COLLECTION=police_alerts
```

**Default Value:** `police_alerts` (if not set)

### 2. Code Changes

#### `internal/storage/firestore.go`
- Added `collectionName` field to `FirestoreClient` struct
- Updated `NewFirestoreClient()` to accept `collectionName` parameter
- Falls back to `"police_alerts"` if empty string is passed

```go
func NewFirestoreClient(ctx context.Context, projectID, collectionName string) (*FirestoreClient, error)
```

#### `internal/storage/police_alerts.go`
- Removed hardcoded `policeAlertsCollection` constant
- Updated `processPoliceAlert()` to use `fc.collectionName`
- Updated `GetPoliceAlertsByDateRange()` to use `fc.collectionName`

#### `cmd/scraper/main.go`
- Added `collectionName` global variable
- Reads `FIRESTORE_COLLECTION` from environment at startup
- Passes collection name to `NewFirestoreClient()`
- Logs collection name at startup

#### `cmd/exporter/main.go`
- Added `--collection` command-line flag
- Reads `FIRESTORE_COLLECTION` from environment as default
- Passes collection name to `NewFirestoreClient()`
- Logs collection name when connecting

### 3. Deployment Scripts

#### `scripts/deploy.sh`
- Reads `FIRESTORE_COLLECTION` from environment (default: `police_alerts`)
- Passes to Cloud Run via `--set-env-vars`
- Logs collection name during deployment

#### `scripts/deploy.bat`
- Reads `FIRESTORE_COLLECTION` from environment (default: `police_alerts`)
- Passes to Cloud Run via `--set-env-vars`
- Logs collection name during deployment

## Usage

### Local Development
```bash
# Set in .env file
FIRESTORE_COLLECTION=police_alerts

# Or export directly
export FIRESTORE_COLLECTION=my_custom_collection

# Run scraper
go run cmd/scraper/main.go
```

### Exporter Tool
```bash
# Use environment variable
export FIRESTORE_COLLECTION=police_alerts
go run cmd/exporter/main.go --project=my-project --start=2025-10-03 --end=2025-10-05

# Or use command-line flag
go run cmd/exporter/main.go \
  --project=my-project \
  --collection=my_custom_collection \
  --start=2025-10-03 \
  --end=2025-10-05
```

### Cloud Run Deployment
```bash
# Using default from environment
export FIRESTORE_COLLECTION=police_alerts
./scripts/deploy.sh

# Or set custom collection
export FIRESTORE_COLLECTION=prod_police_alerts
./scripts/deploy.bat
```

### Update Existing Cloud Run Service
```bash
gcloud run services update waze-scraper \
  --region=us-central1 \
  --set-env-vars FIRESTORE_COLLECTION=new_collection_name
```

## Benefits

1. **Flexibility**: Easy to switch between collections (dev, staging, prod)
2. **Testing**: Can use different collections for testing without code changes
3. **Multi-tenancy**: Can deploy multiple instances with different collections
4. **Configuration**: All configuration in one place (.env file)
5. **Safety**: Default value prevents accidental empty collection names

## Migration Notes

- **No breaking changes**: Default value matches the previous hardcoded value (`police_alerts`)
- **Backwards compatible**: Existing deployments continue to work
- **Optional**: Environment variable is optional; defaults to `police_alerts`

## Examples

### Development Environment
```bash
FIRESTORE_COLLECTION=dev_police_alerts
```

### Staging Environment
```bash
FIRESTORE_COLLECTION=staging_police_alerts
```

### Production Environment
```bash
FIRESTORE_COLLECTION=police_alerts
```

### Regional Collections
```bash
# Sydney region
FIRESTORE_COLLECTION=police_alerts_sydney

# Canberra region
FIRESTORE_COLLECTION=police_alerts_canberra
```
