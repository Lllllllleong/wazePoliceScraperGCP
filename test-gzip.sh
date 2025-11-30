#!/bin/bash
# GZIP Compression Testing Script for Alerts Service
# Based on PERFORMANCE_OPTIMIZATION_PLAN.md Phase 3.1

set -e

echo "======================================"
echo "GZIP Compression Testing Script"
echo "======================================"
echo ""

# Configuration
PORT="${PORT:-8080}"
BASE_URL="http://localhost:${PORT}"
TEST_DATE="${TEST_DATE:-2025-10-15}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper function for test status
test_passed() {
    echo -e "${GREEN}✓ PASSED${NC}: $1"
}

test_failed() {
    echo -e "${RED}✗ FAILED${NC}: $1"
}

test_info() {
    echo -e "${YELLOW}ℹ INFO${NC}: $1"
}

echo "Checking if service is running on port ${PORT}..."
if ! curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Service is not running on port ${PORT}${NC}"
    echo "Please start the service first with:"
    echo "  export GCP_PROJECT_ID=your-project-id"
    echo "  export GCS_BUCKET_NAME=your-bucket-name"
    echo "  ./alerts-service.exe"
    exit 1
fi
test_passed "Service is running"
echo ""

# Test 1: Verify GZIP Headers
echo "======================================"
echo "Test 1: GZIP Headers Verification"
echo "======================================"

echo "Testing with Accept-Encoding: gzip..."
RESPONSE=$(curl -s -H "Accept-Encoding: gzip" "${BASE_URL}/police_alerts?dates=${TEST_DATE}" -v 2>&1)

if echo "$RESPONSE" | grep -q "Content-Encoding: gzip"; then
    test_passed "Content-Encoding: gzip header present"
else
    test_failed "Content-Encoding: gzip header missing"
    echo "$RESPONSE" | grep -i "content-encoding"
fi

if echo "$RESPONSE" | grep -q "Vary: Accept-Encoding"; then
    test_passed "Vary: Accept-Encoding header present (cache compatibility)"
else
    test_failed "Vary: Accept-Encoding header missing"
    echo "$RESPONSE" | grep -i "vary"
fi
echo ""

# Test 2: Verify Uncompressed Fallback
echo "======================================"
echo "Test 2: Uncompressed Fallback"
echo "======================================"

echo "Testing without Accept-Encoding header..."
RESPONSE_UNCOMPRESSED=$(curl -s "${BASE_URL}/police_alerts?dates=${TEST_DATE}" -v 2>&1)

if echo "$RESPONSE_UNCOMPRESSED" | grep -q "Content-Encoding: gzip"; then
    test_failed "Service should NOT compress when client doesn't request it"
else
    test_passed "Uncompressed response for non-gzip clients"
fi
echo ""

# Test 3: Compression Ratio Measurement
echo "======================================"
echo "Test 3: Compression Ratio"
echo "======================================"

echo "Measuring compressed response size..."
COMPRESSED_SIZE=$(curl -s -H "Accept-Encoding: gzip" "${BASE_URL}/police_alerts?dates=${TEST_DATE}" --compressed -w '%{size_download}' -o /tmp/compressed.jsonl 2>/dev/null)
echo "Compressed size: ${COMPRESSED_SIZE} bytes"

echo "Measuring uncompressed response size..."
UNCOMPRESSED_SIZE=$(curl -s "${BASE_URL}/police_alerts?dates=${TEST_DATE}" -w '%{size_download}' -o /tmp/uncompressed.jsonl 2>/dev/null)
echo "Uncompressed size: ${UNCOMPRESSED_SIZE} bytes"

if [ "$UNCOMPRESSED_SIZE" -gt 0 ]; then
    REDUCTION=$(echo "scale=2; (($UNCOMPRESSED_SIZE - $COMPRESSED_SIZE) * 100) / $UNCOMPRESSED_SIZE" | bc)
    echo -e "Compression ratio: ${GREEN}${REDUCTION}%${NC} reduction"
    
    if (( $(echo "$REDUCTION > 70" | bc -l) )); then
        test_passed "Compression ratio meets target (>70%)"
    elif (( $(echo "$REDUCTION > 50" | bc -l) )); then
        test_info "Compression ratio is good (>50%) but below ideal target (70-90%)"
        echo "This may be due to frequent flushing in streaming responses"
    else
        test_failed "Compression ratio below expected (got ${REDUCTION}%, expected >70%)"
    fi
else
    test_info "No data returned for test date ${TEST_DATE}"
fi
echo ""

# Test 4: Data Integrity
echo "======================================"
echo "Test 4: Data Integrity Check"
echo "======================================"

echo "Verifying decompressed data is valid JSONL..."
if [ -f /tmp/compressed.jsonl ] && [ -s /tmp/compressed.jsonl ]; then
    FIRST_LINE=$(head -n 1 /tmp/compressed.jsonl)
    if echo "$FIRST_LINE" | jq . > /dev/null 2>&1; then
        test_passed "Decompressed data is valid JSON"
        
        # Check for required fields
        REQUIRED_FIELDS=("UUID" "Type" "LocationGeo" "PublishTime" "ExpireTime")
        ALL_FIELDS_PRESENT=true
        for field in "${REQUIRED_FIELDS[@]}"; do
            if echo "$FIRST_LINE" | jq -e ".$field" > /dev/null 2>&1; then
                test_passed "Required field '$field' present"
            else
                test_failed "Required field '$field' missing"
                ALL_FIELDS_PRESENT=false
            fi
        done
    else
        test_failed "Decompressed data is not valid JSON"
        echo "First line: $FIRST_LINE"
    fi
else
    test_info "No data to validate (empty response)"
fi
echo ""

# Test 5: Response Time Comparison
echo "======================================"
echo "Test 5: Response Time"
echo "======================================"

echo "Measuring compressed response time..."
COMPRESSED_TIME=$(curl -s -H "Accept-Encoding: gzip" "${BASE_URL}/police_alerts?dates=${TEST_DATE}" --compressed -w '%{time_total}' -o /dev/null 2>/dev/null)
echo "Compressed: ${COMPRESSED_TIME}s"

echo "Measuring uncompressed response time..."
UNCOMPRESSED_TIME=$(curl -s "${BASE_URL}/police_alerts?dates=${TEST_DATE}" -w '%{time_total}' -o /dev/null 2>/dev/null)
echo "Uncompressed: ${UNCOMPRESSED_TIME}s"

TIME_DIFF=$(echo "scale=2; (($UNCOMPRESSED_TIME - $COMPRESSED_TIME) * 100) / $UNCOMPRESSED_TIME" | bc 2>/dev/null || echo "0")
if (( $(echo "$TIME_DIFF > 0" | bc -l 2>/dev/null || echo 0) )); then
    test_passed "Compressed response is faster (${TIME_DIFF}% improvement)"
elif (( $(echo "$TIME_DIFF < 10" | bc -l 2>/dev/null || echo 1) )); then
    test_info "Response times are similar (acceptable for streaming)"
else
    test_info "Compressed response slightly slower (CPU overhead is normal for small payloads)"
fi
echo ""

# Cleanup
rm -f /tmp/compressed.jsonl /tmp/uncompressed.jsonl

# Summary
echo "======================================"
echo "Test Summary"
echo "======================================"
echo "Test Date: ${TEST_DATE}"
echo "All critical tests completed!"
echo ""
echo "Next steps from PERFORMANCE_OPTIMIZATION_PLAN.md:"
echo "1. If tests pass, deploy to Cloud Run"
echo "2. Monitor Cloud Run metrics for 24-48 hours"
echo "3. Check for bandwidth cost reduction (70-90% expected)"
echo "4. Proceed with Phase 2 (Firestore streaming) if desired"
echo ""
