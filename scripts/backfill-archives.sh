#!/usr/bin/env bash
# backfill-archives.sh
#
# Triggers the archive service for each date from 2026-01-10 to 2026-04-05,
# backfilling missing GCS archives before Firestore is torn down.
#
# Dates with data (Jan 10 – Mar 16):  expect "Successfully archived N alerts"
# Dates without data (Mar 17 – Apr 5): expect "No alerts to archive"
# Already-archived dates:              skipped automatically (idempotent)
#
# Usage:
#   ./scripts/backfill-archives.sh
#
# Requirements:
#   - gcloud CLI authenticated with an account that has the Cloud Run Invoker role
#   - curl

set -euo pipefail

ARCHIVE_SERVICE_URL="https://archive-service-u6cjbro2iq-uc.a.run.app/"
START_DATE="2026-01-10"
END_DATE="2026-04-05"
TOKEN_REFRESH_INTERVAL=30

# Generate list of dates from START_DATE to END_DATE (inclusive)
generate_dates() {
    local current="$1"
    local end="$2"
    while [[ "$current" < "$end" || "$current" == "$end" ]]; do
        echo "$current"
        current=$(date -d "$current + 1 day" +%Y-%m-%d 2>/dev/null || date -v+1d -j -f "%Y-%m-%d" "$current" +%Y-%m-%d)
    done
}

echo "Backfilling archives from $START_DATE to $END_DATE"
echo "Service: $ARCHIVE_SERVICE_URL"
echo "---"

TOKEN=$(gcloud auth print-identity-token)
count=0
errors=0

while IFS= read -r date; do
    # Refresh token every TOKEN_REFRESH_INTERVAL requests
    if (( count > 0 && count % TOKEN_REFRESH_INTERVAL == 0 )); then
        echo "[token] Refreshing identity token..."
        TOKEN=$(gcloud auth print-identity-token)
    fi

    response=$(curl -s -w "\n%{http_code}" -X POST "$ARCHIVE_SERVICE_URL" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"date\": \"$date\"}")

    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)

    if [[ "$status" == "200" ]]; then
        echo "[ok]  $date — $body"
    else
        echo "[err] $date — HTTP $status — $body"
        (( errors++ )) || true
    fi

    (( count++ )) || true
done < <(generate_dates "$START_DATE" "$END_DATE")

echo "---"
echo "Done. $count dates processed, $errors errors."

if (( errors > 0 )); then
    echo "Re-run this script to retry failed dates — already-archived dates will be skipped."
    exit 1
fi
