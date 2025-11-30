# GZIP Testing Quick Start Guide

## Prerequisites
- Service compiled successfully ✓
- GCP credentials configured (gcloud auth application-default login)
- Required environment variables set

## Step 1: Set Environment Variables

```bash
# Required variables
export GCP_PROJECT_ID="your-gcp-project-id"
export GCS_BUCKET_NAME="your-bucket-name"

# Optional variables (have defaults)
export PORT="8080"
export FIRESTORE_COLLECTION="police_alerts"
export RATE_LIMIT_PER_MINUTE="30"
```

**To find your values:**
```bash
# Get your project ID
gcloud config get-value project

# List your GCS buckets
gsutil ls
```

## Step 2: Start the Service

**Option A: Using the helper script**
```bash
./start-service.sh
```

**Option B: Run directly**
```bash
./alerts-service.exe
```

The service will start on http://localhost:8080

## Step 3: Run Tests

**In a new terminal**, run the test script:

```bash
# Test with default date (2025-10-15)
./test-gzip.sh

# Test with a specific date
TEST_DATE=2025-10-20 ./test-gzip.sh
```

## Manual Testing Commands

If you prefer to test manually:

### Test 1: Check GZIP Headers
```bash
curl -H "Accept-Encoding: gzip" \
     "http://localhost:8080/police_alerts?dates=2025-10-15" \
     --compressed -v 2>&1 | grep -i "content-encoding"
```
**Expected**: `< Content-Encoding: gzip`

### Test 2: Measure Compression Ratio
```bash
# Compressed
curl -H "Accept-Encoding: gzip" \
     "http://localhost:8080/police_alerts?dates=2025-10-15" \
     --compressed -w '\nCompressed: %{size_download} bytes\n' -o /dev/null

# Uncompressed
curl "http://localhost:8080/police_alerts?dates=2025-10-15" \
     -w '\nUncompressed: %{size_download} bytes\n' -o /dev/null
```
**Expected**: 70-90% size reduction

### Test 3: Verify Data Integrity
```bash
curl -H "Accept-Encoding: gzip" \
     "http://localhost:8080/police_alerts?dates=2025-10-15" \
     --compressed -o test.jsonl

# Check if valid JSON
head -n 1 test.jsonl | jq .
```

### Test 4: Health Check
```bash
curl http://localhost:8080/health
```
**Expected**: `OK`

## Expected Results

### ✅ Success Indicators:
- Service starts without errors
- `Content-Encoding: gzip` header present when requested
- `Vary: Accept-Encoding` header present (cache compatibility)
- Compression ratio: 70-90% reduction
- Data decompresses to valid JSONL
- All required fields present (UUID, Type, LocationGeo, PublishTime, ExpireTime)

### ⚠️ Common Issues:

**Issue**: "Service is not running"
- **Solution**: Check if ports are available, verify environment variables

**Issue**: "No data returned"
- **Solution**: Test with dates that have archived data (check GCS bucket)

**Issue**: Compression ratio below 70%
- **Solution**: This may be normal for streaming responses with frequent flushes
- Still acceptable if >50%

## Next Steps After Testing

1. ✅ Verify all tests pass
2. 📊 Review compression metrics
3. 🚀 Deploy to Cloud Run (use existing CI/CD pipeline)
4. 📈 Monitor Cloud Run metrics for 24-48 hours:
   - Response size (should drop 70-90%)
   - Bandwidth costs (should drop similarly)
   - Error rate (should remain <1%)
5. 🔄 If successful, consider Phase 2 (Firestore streaming)

## Troubleshooting

### View service logs:
The service logs to stdout, so you'll see logs in the terminal where you started it.

### Stop the service:
Press `Ctrl+C` in the terminal running the service.

### Check if port is in use:
```bash
netstat -ano | grep :8080
```

### Re-compile after changes:
```bash
go build -o alerts-service.exe cmd/alerts-service/main.go
```

## Files Created
- `test-gzip.sh` - Automated test suite
- `start-service.sh` - Helper script to start service
- `alerts-service.exe` - Compiled binary
- `TESTING_GUIDE.md` - This file

## Reference
See `PERFORMANCE_OPTIMIZATION_PLAN.md` for detailed implementation plan and rationale.
