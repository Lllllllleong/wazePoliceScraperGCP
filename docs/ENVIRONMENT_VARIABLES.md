# Environment Variables Reference

## Complete List of Environment Variables

### 🔑 Required Variables

| Variable | Used By | Required? | Description |
|----------|---------|-----------|-------------|
| `GCP_PROJECT_ID` | Scraper, Exporter | ✅ Yes | Your Google Cloud Platform Project ID |

### ⚙️ Optional Variables

| Variable | Used By | Required? | Default | Description |
|----------|---------|-----------|---------|-------------|
| `PORT` | Scraper | ❌ No | `8080` | HTTP server port (Cloud Run sets automatically) |
| `WAZE_BBOXES` | Scraper | ❌ No | Sydney-Canberra | Semicolon-separated bounding boxes to scrape |
| `GOOGLE_APPLICATION_CREDENTIALS` | Both | ❌ No | - | Path to service account key JSON (local dev only) |

---

## Detailed Reference

### `GCP_PROJECT_ID`

**Purpose:** Identifies which GCP project to use for Firestore operations

**Used by:**
- `cmd/scraper/main.go` - To connect to Firestore
- `cmd/exporter/main.go` - To query Firestore data

**Format:** String (your GCP project ID)

**Examples:**
```bash
# Production
GCP_PROJECT_ID=waze-scraper-prod

# Development
GCP_PROJECT_ID=waze-scraper-dev
```

**How to find:**
1. Go to https://console.cloud.google.com/
2. Click the project dropdown at the top
3. Your project ID is shown next to the project name

---

### `PORT`

**Purpose:** Specifies which port the HTTP server listens on

**Used by:**
- `cmd/scraper/main.go` - HTTP server port

**Format:** Integer (port number)

**Default:** `8080`

**Examples:**
```bash
PORT=8080    # Default
PORT=3000    # Custom for local dev
```

**Notes:**
- Cloud Run sets this automatically (usually 8080)
- Only override for local development
- Must be > 1024 and < 65535

---

### `WAZE_BBOXES`

**Purpose:** Defines geographic areas to scrape

**Used by:**
- `cmd/scraper/main.go` - Bounding boxes for Waze API requests

**Format:** Semicolon-separated strings: `"west,south,east,north;west,south,east,north"`

**Default:** Sydney to Canberra (Hume Highway) - 4 bounding boxes
```go
"150.38822599217056,-34.254577954626086,151.00867887302994,-33.937977044844004"  // Sydney
"149.58926145838367,-34.76915040190209,150.83016722010242,-34.138639582841435"   // Middle
"149.09281124417694,-35.21080621952668,150.3337170058957,-34.583661538587855"    // Canberra
"148.80885598970738,-35.4530012424677,149.42930887056676,-35.14096097196958"     // Canberra city
```

**Format Details:**
- Each bbox: `west,south,east,north` (longitude_min, latitude_min, longitude_max, latitude_max)
- Multiple bboxes: Separate with semicolon (`;`)

**Examples:**
```bash
# Single area
WAZE_BBOXES="150.388,-34.254,151.008,-33.937"

# Multiple areas
WAZE_BBOXES="150.388,-34.254,151.008,-33.937;149.589,-34.769,150.830,-34.138"

# Singapore
WAZE_BBOXES="103.6,1.15,104.0,1.45"

# Empty = use defaults
WAZE_BBOXES=
```

**How to find coordinates:**
1. Visit: https://boundingbox.klokantech.com/
2. Draw a rectangle around your area
3. Select "CSV" format
4. Copy the coordinates

---

### `GOOGLE_APPLICATION_CREDENTIALS`

**Purpose:** Path to GCP service account key for authentication

**Used by:**
- Both scraper and exporter (when running locally)

**Format:** Absolute file path to JSON key file

**Default:** Not set (uses default authentication)

**Examples:**
```bash
# Linux/Mac
GOOGLE_APPLICATION_CREDENTIALS=/home/user/.gcp/service-account.json

# Windows
GOOGLE_APPLICATION_CREDENTIALS=C:\Users\User\.gcp\service-account.json
```

**When needed:**
- ✅ Local development without `gcloud auth application-default login`
- ✅ CI/CD pipelines
- ❌ Cloud Run (uses default service account automatically)
- ❌ When using `gcloud auth application-default login`

**How to create:**
1. Go to https://console.cloud.google.com/iam-admin/serviceaccounts
2. Select or create a service account
3. Click "Keys" → "Add Key" → "Create new key"
4. Choose JSON format
5. Download and set path in this variable

**⚠️ Security:**
- Never commit this JSON file to git!
- Add to `.gitignore`: `service-account-key.json`
- Use least-privilege permissions

---

## Environment-Specific Configurations

### Local Development

```bash
# .env file
GCP_PROJECT_ID=waze-scraper-dev
PORT=8080
WAZE_BBOXES=  # Use defaults

# Authenticate with gcloud (recommended)
# gcloud auth application-default login

# OR use service account key (not recommended)
# GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
```

### Cloud Run Deployment

```bash
# Set via gcloud deploy command
gcloud run deploy waze-scraper \
  --set-env-vars GCP_PROJECT_ID=waze-scraper-prod \
  --set-env-vars WAZE_BBOXES="150.388,-34.254,151.008,-33.937"

# PORT is set automatically by Cloud Run
# GOOGLE_APPLICATION_CREDENTIALS not needed (uses default service account)
```

### Exporter Tool

```bash
# Option 1: Use environment variable
export GCP_PROJECT_ID=waze-scraper-prod
go run cmd/exporter/main.go --start=2025-10-03 --end=2025-10-05

# Option 2: Use command-line flag
go run cmd/exporter/main.go \
  --project=waze-scraper-prod \
  --start=2025-10-03 \
  --end=2025-10-05
```

---

## Validation

### Check if variables are set:

```bash
# Linux/Mac
echo $GCP_PROJECT_ID
echo $PORT
echo $WAZE_BBOXES

# Windows (cmd)
echo %GCP_PROJECT_ID%
echo %PORT%
echo %WAZE_BBOXES%

# Windows (PowerShell)
$env:GCP_PROJECT_ID
$env:PORT
$env:WAZE_BBOXES
```

### Load .env file (local dev):

```bash
# Linux/Mac (bash)
export $(cat .env | xargs)

# Or use direnv
direnv allow

# Or source it
set -a
source .env
set +a
```

---

## Troubleshooting

### "GCP_PROJECT_ID environment variable is required"
- Set `GCP_PROJECT_ID` in your environment
- Or pass `--project` flag to exporter

### "Failed to connect to Firestore"
- Check `GCP_PROJECT_ID` is correct
- Authenticate: `gcloud auth application-default login`
- Or set `GOOGLE_APPLICATION_CREDENTIALS`

### Bounding boxes not working
- Check format: `"west,south,east,north"`
- Multiple boxes: Use semicolon separator `;`
- Verify coordinates are correct (west < east, south < north)

---

## Files

- **`.env.example`** - Template with all variables (commit this)
- **`.env`** - Your actual configuration (DO NOT commit)
- **`.gitignore`** - Should include `.env` and `*.json` keys

Make sure your `.gitignore` contains:
```
.env
service-account-key.json
*.env.local
```
