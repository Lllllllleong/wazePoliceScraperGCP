# Time Range Filter - Quick Reference

## New UI Flow

```
┌─────────────────────────────────────────────────────────┐
│  📅 Step 1: Select Dates to Load                       │
│  ┌───────────────────────────────────────────────────┐ │
│  │ [Click to select dates...] (Flatpickr)           │ │
│  └───────────────────────────────────────────────────┘ │
│  ✓ 3 days selected                                     │
│  ┌──────────────────┐  ┌──────────────────────────┐   │
│  │ 🔄 Reset Dashboard│  │ 📥 Load Data             │   │
│  └──────────────────┘  └──────────────────────────┘   │
│  ⏳ Loading day 2 of 3: 2025-10-04...                  │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  🔍 Step 2: Alert Filters (DISABLED until data loaded) │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Subtype: [All Types ▼]                           │ │
│  │ City: [All Cities ▼]                              │ │
│  │ Min Reliability: [━━━━○━━━━━] 0                   │ │
│  │ Min Thumbs Up: [0]                                │ │
│  │ [Apply Filters]  [Reset Filters]                  │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Data Loading Process

```
User selects dates: [Oct 3] [Oct 4] [Oct 6]
                     ↓
Chronologically sorted: Oct 3 → Oct 4 → Oct 6
                     ↓
┌─────────────────────────────────────────┐
│ For each date:                          │
│                                         │
│ Day Oct 3:                              │
│   Query: publish_time <= 23:59:59       │
│          expire_time >= 00:00:00        │
│   Results: Alert A, Alert B             │
│   Map: {uuid-A: Alert A, uuid-B: B}     │
│                                         │
│ Day Oct 4:                              │
│   Query: publish_time <= 23:59:59       │
│          expire_time >= 00:00:00        │
│   Results: Alert B, Alert C             │
│   Map: {uuid-A: A, uuid-B: B, uuid-C: C}│
│         (Alert B skipped - duplicate!)  │
│                                         │
│ Day Oct 6:                              │
│   Query: publish_time <= 23:59:59       │
│          expire_time >= 00:00:00        │
│   Results: Alert D                      │
│   Map: {..., uuid-D: Alert D}           │
│                                         │
└─────────────────────────────────────────┘
                     ↓
Final allAlerts = [Alert A, Alert B, Alert C, Alert D]
                     ↓
Enable Stage 2 Filters → User can now filter/view
```

## Button Behavior

### 🔄 Reset Dashboard/Data
- Clears `allAlerts` and `alertsMap`
- Clears date picker selection
- Disables Stage 2 filters
- Clears map markers
- Resets all statistics to "-"
- Shows welcome message

### 📥 Load Data
- Only enabled when dates selected
- Shows loading spinner
- Queries Firestore for each selected date
- Updates progress message
- Enables Stage 2 on success
- Populates city filter dropdown

### Apply Filters (Stage 2)
- Filters already-loaded data
- Does NOT query Firestore
- Updates map, list, statistics

### Reset Filters (Stage 2)
- Resets filter controls to defaults
- Keeps loaded data
- Re-applies filters (showing all loaded data)

## Key Code Changes

### Global State
```javascript
// OLD
let allAlerts = [];

// NEW
let allAlerts = [];           // Array of alerts
let alertsMap = new Map();    // UUID -> alert (deduplication)
let selectedDates = [];       // User-selected dates
let flatpickrInstance = null; // Date picker instance
```

### Function Mapping

| Old Function | New Function | Purpose |
|--------------|--------------|---------|
| `loadAlertsFromFirestore()` | `loadAlertsForSelectedDates()` | Load data |
| `setInitialDateRange()` | `initDatePicker()` | Setup date selection |
| N/A | `disableStage2UI()` | Lock filters |
| N/A | `enableStage2UI()` | Unlock filters |
| N/A | `resetDashboard()` | Full reset |
| `resetFilters()` | `resetFilters()` | Same (modified) |

## Date Picker Configuration

```javascript
flatpickr('#date-picker', {
  mode: 'multiple',              // Select multiple dates
  dateFormat: 'Y-m-d',           // 2025-10-03
  minDate: '2025-10-03',         // Hardcoded start
  maxDate: new Date(),           // Today
  inline: false,                 // Dropdown (not always visible)
  onChange: updateSelectedDatesDisplay
});
```

## Firestore Query Example

**Single Day Query:**
```javascript
const dayStart = new Date('2025-10-03T00:00:00');
const dayEnd = new Date('2025-10-03T23:59:59.999');

db.collection('police_alerts')
  .where('publish_time', '<=', dayEnd)
  .where('expire_time', '>=', dayStart)
  .get();
```

**What this captures:**
- Alert published on Oct 3 (any time) → ✅
- Alert published on Oct 2, expires Oct 3 → ✅
- Alert published on Oct 1, expires Oct 5 → ✅
- Alert published on Oct 5 → ❌
- Alert published on Oct 2, expires Oct 2 → ❌

## Testing Commands

Open browser console and test:

```javascript
// Check global state
console.log('Alerts loaded:', allAlerts.length);
console.log('Map size:', alertsMap.size);
console.log('Selected dates:', selectedDates);

// Manually trigger functions
resetDashboard();
enableStage2UI();
disableStage2UI();

// Check for duplicates
const uuids = allAlerts.map(a => a.UUID);
const duplicates = uuids.filter((u, i) => uuids.indexOf(u) !== i);
console.log('Duplicate UUIDs:', duplicates); // Should be []
```

## Common Issues & Solutions

### Issue: "Load Data" button stays disabled
**Solution**: Check that dates are selected in date picker

### Issue: Firestore index error
**Solution**: Click the link in console to create composite index

### Issue: No alerts loaded
**Possible causes**:
- Selected dates have no data
- Firestore rules prevent read access
- Not authenticated anonymously

### Issue: Stage 2 filters don't work
**Check**:
- Are they enabled? (opacity should be 1)
- Is data loaded? (check allAlerts.length)

### Issue: Seeing duplicate alerts
**Debug**:
```javascript
// Check map vs array length
console.log('Map:', alertsMap.size, 'Array:', allAlerts.length);
// They should match!
```

## Browser Compatibility

- ✅ Chrome/Edge (Chromium)
- ✅ Firefox
- ✅ Safari
- ✅ Mobile browsers

Flatpickr has excellent browser support.

## Performance Tips

1. **Don't select too many days at once**: Each day = 1 Firestore query
2. **Consecutive days are fine**: Deduplication handles overlaps efficiently
3. **Use Reset Filters instead of Reset Dashboard**: Keeps data in memory
4. **Consider caching**: Future improvement for frequently selected dates

---

**Need help?** Check `REFACTOR_SUMMARY.md` for detailed documentation.
