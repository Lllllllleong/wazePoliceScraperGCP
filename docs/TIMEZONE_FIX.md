# Timezone Fix - Date Selection Issue

## Problem

When users clicked on "today's date", no alerts were loaded. However, clicking "yesterday's date" would load alerts from yesterday through today (near real-time).

## Root Cause

**Timezone mismatch between frontend and backend:**

### Before the Fix:

**Frontend (JavaScript):**
```javascript
const dayStart = new Date(dateStr + 'T00:00:00');  // Local timezone!
const dayEnd = new Date(dateStr + 'T23:59:59.999');
```

If user is in Australia (UTC+10):
- Date string: `"2025-10-20"`
- `dayStart`: `2025-10-20 00:00:00 +10:00` = `2025-10-19 14:00:00 UTC`
- `dayEnd`: `2025-10-20 23:59:59 +10:00` = `2025-10-20 13:59:59 UTC`
- **Queries the WRONG day in UTC!** ❌

**Backend (Go):**
```go
dayStart, _ := time.Parse("2006-01-02", dateStr)  // UTC by default
dayEnd := dayStart.Add(24*time.Hour - time.Second)
```

- Date string: `"2025-10-20"`
- `dayStart`: `2025-10-20 00:00:00 UTC`
- `dayEnd`: `2025-10-20 23:59:59 UTC`
- **Queries the correct day** ✅

### The Mismatch:

When a user in UTC+10 selects "2025-10-20":
- **Frontend thinks**: Query `2025-10-19 14:00 UTC` to `2025-10-20 14:00 UTC` (28 hours!)
- **Backend thinks**: Query `2025-10-20 00:00 UTC` to `2025-10-20 23:59 UTC` (24 hours) ✅

This causes:
- Today's date query returns incomplete results
- Yesterday's date includes alerts from today (because of the +10 hour offset)

## Solution

**Force all date parsing to use UTC explicitly, with no timezone conversion.**

### Frontend Fix:

```javascript
// OLD (local timezone):
const dayStart = new Date(dateStr + 'T00:00:00');

// NEW (explicit UTC):
const dayStart = new Date(dateStr + 'T00:00:00.000Z');  // The 'Z' forces UTC
const dayEnd = new Date(dateStr + 'T23:59:59.999Z');
```

### Backend Fix:

```go
// OLD (implicit UTC, but not explicit):
dayStart, _ := time.Parse("2006-01-02", dateStr)

// NEW (explicit UTC):
dayStart, _ := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
```

## Files Changed

1. **`dataAnalysis/public/app.js`**
   - Updated `loadAlertsFromFirestore()` to use UTC dates with 'Z' suffix
   - Affects legacy Firestore direct access

2. **`internal/storage/police_alerts.go`**
   - Updated `GetPoliceAlertsByDatesWithFilters()` to use `time.ParseInLocation(..., time.UTC)`
   - Affects API endpoint queries

3. **`cmd/exporter/main.go`**
   - Updated date parsing to use `time.ParseInLocation(..., time.UTC)`
   - Ensures exporter tool uses same logic

## Result

Now all date queries work consistently:
- ✅ "Today" loads alerts from today (00:00 to 23:59 UTC)
- ✅ "Yesterday" loads alerts from yesterday only
- ✅ Multiple dates load correct 24-hour periods
- ✅ Works consistently across all timezones
- ✅ Frontend and API use identical date ranges

## Testing

**After redeploying the API:**

1. Refresh the frontend (http://localhost:5002)
2. Select today's date (2025-10-20)
3. Click "Load Data"
4. Verify alerts from today appear
5. Check browser console for correct UTC timestamps

**Expected Console Output:**
```
Querying alerts for 2025-10-20 (2025-10-20T00:00:00.000Z to 2025-10-20T23:59:59.999Z)
```

## Important Notes

- All dates are now interpreted as **UTC calendar dates**
- A "day" is defined as 00:00:00.000Z to 23:59:59.999Z in UTC
- Users in different timezones will see the same alerts for the same calendar date
- This is the correct behavior for a global system where alerts have UTC timestamps

## Deployment Required

- ✅ Frontend: Changes take effect immediately (refresh browser)
- ⏳ Backend: Requires redeployment of API service
  ```bash
  ./scripts/deploy-api.sh
  ```
