# Canberra Timezone Fix

## Problem
When users in Canberra (AEDT/AEST, UTC+10/+11) selected a date like "2025-10-20", they were getting alerts from the wrong 24-hour period because:
- Frontend was treating dates as UTC
- Backend stores all data in UTC (correct)
- User expects "2025-10-20" to mean "2025-10-20 in Canberra time"

## Solution
Updated the **frontend only** to properly handle Canberra timezone without modifying the scraper or backend storage:

### 1. Firestore Direct Access (Legacy)
**File**: `dataAnalysis/public/app.js` - `loadAlertsFromFirestore()`

```javascript
// OLD: Parsed dates as UTC
const dayStart = new Date(dateStr + 'T00:00:00.000Z'); // UTC
const dayEnd = new Date(dateStr + 'T23:59:59.999Z');   // UTC

// NEW: Parse dates in local Canberra timezone
const localDate = new Date(dateStr + 'T00:00:00'); // Local timezone
const dayStart = new Date(localDate.getTime());
const dayEnd = new Date(localDate.getTime() + 24 * 60 * 60 * 1000 - 1);
```

When querying Firestore, these local times are automatically converted to UTC timestamps, so the backend query works correctly.

### 2. API Access
**File**: `dataAnalysis/public/app.js` - `loadAlertsFromAPI()`

Since the API backend interprets dates as UTC, we:
1. Convert Canberra dates to multiple UTC dates that cover the same 24-hour period
2. Send expanded UTC date list to API
3. Filter returned alerts client-side to only include those active during Canberra day

```javascript
// Example: User selects 2025-10-20 in Canberra (UTC+11)
// This is 2025-10-19 13:00 UTC to 2025-10-20 12:59 UTC
// So we query API for UTC dates: ["2025-10-19", "2025-10-20"]
// Then filter client-side to only keep alerts active during Canberra Oct 20
```

## Testing
1. Open browser console (enabled console.log for debugging)
2. Select today's date (2025-10-20)
3. Click "Load Data"
4. Check console output:
   ```
   Loading alerts via API for dates (in Canberra timezone): ["2025-10-20"]
   Converted to UTC dates for API query: ["2025-10-19", "2025-10-20"]
   ✅ API returned X alerts for UTC dates 2025-10-19, 2025-10-20
   📦 Processed Y unique alerts from API
   ```
5. Verify alerts shown are from Oct 20 00:00-23:59 Canberra time

## Key Changes
- **No scraper changes** - continues to store UTC timestamps correctly
- **No API backend changes** - continues to interpret dates as UTC
- **Frontend changes only** - handles timezone conversion client-side
- **Console logging enabled** - for debugging and verification

## Files Modified
1. `dataAnalysis/public/app.js`:
   - Re-enabled console.log (commented out disabling code)
   - Updated `loadAlertsFromFirestore()` to use local timezone
   - Updated `loadAlertsFromAPI()` to convert Canberra dates to UTC range and filter results

## Timezone Behavior
- **Canberra AEDT** (Oct-Apr): UTC+11
- **Canberra AEST** (Apr-Oct): UTC+10
- JavaScript automatically handles DST transitions via local timezone
- All dates selected by user are interpreted as Canberra local dates
- Backend storage remains UTC (ISO 8601 / RFC3339 format)
