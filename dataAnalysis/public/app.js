/**
 * Waze Police Alert Dashboard
 * 
 * A single-page application for visualizing police alert data on an interactive map.
 * Uses Firebase Anonymous Authentication and streams data from the alerts-service API.
 * 
 * Key Features:
 * - Multi-date selection with Flatpickr
 * - Real-time JSONL streaming for efficient data loading
 * - Leaflet.js map with timeline visualization
 * - Dynamic filtering by alert type, street, and verification status
 */

// Global state
let allAlerts = []; // Will store alerts as array
let alertsMap = new Map(); // Map for deduplication using UUID as key
let filteredAlerts = [];
let map = null;
let markers = [];
let selectedDates = []; // Array of selected date strings
let flatpickrInstance = null;
let selectedSubtypes = []; // Array of selected subtypes for filtering
let selectedStreets = []; // Array of selected streets for filtering
let timelineLayer = null; // Timeline layer for temporal visualization
let timelineControl = null; // Timeline slider control

// Firebase Auth state
let firebaseApp = null;
let firebaseAuth = null;
let currentIdToken = null;
let tokenRefreshInterval = null;

// Constants for date range
const MIN_DATE = '2025-09-26';
const MAX_SELECTABLE_DATES = 7; // Maximum number of dates that can be selected (reduced due to large data size)

// Helper function to format dates as dd-mm-yyyy HH:MM:SS
function formatDateDDMMYYYY(date, includeTime = true) {
    const d = new Date(date);
    if (isNaN(d.getTime())) return 'Invalid Date';

    const day = String(d.getDate()).padStart(2, '0');
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const year = d.getFullYear();

    if (!includeTime) {
        return `${day}-${month}-${year}`;
    }

    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    const seconds = String(d.getSeconds()).padStart(2, '0');

    return `${day}-${month}-${year} ${hours}:${minutes}:${seconds}`;
}

/**
 * Parses a timestamp from various formats to milliseconds since epoch.
 * @param {string|number|Date} timestamp - The timestamp to parse.
 * @returns {number} - The timestamp in milliseconds since epoch, or 0 if invalid.
 */
function parseTimestamp(timestamp) {
    if (!timestamp) return 0;

    // If it's already a number (milliseconds), return it
    if (typeof timestamp === 'number') return timestamp;

    // If it's a string (ISO 8601 / RFC3339), parse it
    if (typeof timestamp === 'string') {
        const date = new Date(timestamp);
        return date.getTime();
    }

    // If it's a Date object
    if (timestamp instanceof Date) return timestamp.getTime();

    return 0;
}

/**
 * Initialize Firebase Authentication
 */
async function initFirebaseAuth() {
    try {
        // Initialize Firebase App
        firebaseApp = window.FirebaseApp.initializeApp(window.FIREBASE_CONFIG);
        firebaseAuth = window.FirebaseAuth.getAuth(firebaseApp);
        
        // Connect to emulator in local development
        if (window.USE_FIREBASE_EMULATOR) {
            window.FirebaseAuth.connectAuthEmulator(
                firebaseAuth, 
                `http://${window.FIREBASE_EMULATOR_HOST}`
            );
        }
        
        return true;
    } catch (error) {
        console.error('Firebase initialization failed:', error);
        return false;
    }
}

/**
 * Ensure user is authenticated and return a valid ID token
 */
async function ensureAuthenticated() {
    if (!firebaseAuth) {
        throw new Error('Firebase Auth not initialized');
    }

    // Return cached token if available
    if (currentIdToken) {
        return currentIdToken;
    }

    // Sign in anonymously
    const userCredential = await window.FirebaseAuth.signInAnonymously(firebaseAuth);
    currentIdToken = await userCredential.user.getIdToken();
    
    // Set up token refresh (every 50 minutes, tokens expire in 60)
    if (tokenRefreshInterval) {
        clearInterval(tokenRefreshInterval);
    }
    
    tokenRefreshInterval = setInterval(async () => {
        try {
            const user = firebaseAuth.currentUser;
            if (user) {
                currentIdToken = await user.getIdToken(true);
            }
        } catch (error) {
            console.error('Token refresh failed:', error);
            currentIdToken = null;
        }
    }, 50 * 60 * 1000);
    
    return currentIdToken;
}

// Disclaimer Modal Handling
function initDisclaimerModal() {
    const disclaimerModal = document.getElementById('disclaimer-modal');
    const disclaimerAcceptBtn = document.getElementById('disclaimer-accept-btn');

    // Check if user has already accepted the disclaimer
    const hasAccepted = localStorage.getItem('disclaimerAccepted');

    if (hasAccepted === 'true') {
        // User has already accepted, hide modal
        disclaimerModal.classList.add('hidden');
        document.body.classList.remove('disclaimer-active');
    } else {
        // Show modal and blur background
        disclaimerModal.classList.remove('hidden');
        document.body.classList.add('disclaimer-active');
    }

    // Handle accept button click
    disclaimerAcceptBtn.addEventListener('click', () => {
        // Save acceptance to localStorage
        localStorage.setItem('disclaimerAccepted', 'true');

        // Hide modal with animation
        disclaimerModal.style.opacity = '0';
        disclaimerModal.style.transition = 'opacity 0.3s ease-out';

        setTimeout(() => {
            disclaimerModal.classList.add('hidden');
            document.body.classList.remove('disclaimer-active');
            disclaimerModal.style.opacity = '1';
        }, 300);
    });
}

// Initialize the application
document.addEventListener('DOMContentLoaded', async () => {
    // Initialize Firebase Authentication
    const authInitialized = await initFirebaseAuth();
    if (!authInitialized) {
        const alertList = document.getElementById('alert-list');
        alertList.innerHTML = '<p class="loading-message" style="color: var(--danger-color);">❌ Authentication initialization failed. Please refresh the page.</p>';
        return;
    }

    // Initialize UI components
    initDisclaimerModal();
    initMap();
    initDatePicker();
    initEventListeners();
    disableStage2UI();

    // Welcome message
    const alertList = document.getElementById('alert-list');
    alertList.innerHTML = '<p class="loading-message">👋 Welcome! Please select dates above and click "Load Data" to begin.</p>';

    // Pre-authenticate in background for faster data loading
    ensureAuthenticated().catch(() => {
        // Will retry on demand if pre-auth fails
    });
});

// Initialize Leaflet map
function initMap() {
    map = L.map('map').setView(MAP_CONFIG.center, MAP_CONFIG.zoom);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors',
        minZoom: MAP_CONFIG.minZoom,
        maxZoom: MAP_CONFIG.maxZoom
    }).addTo(map);
}

// Initialize Flatpickr date picker
function initDatePicker() {
    // Determine the max selectable date (current day)
    const today = new Date();

    flatpickrInstance = flatpickr('#date-picker', {
        mode: 'multiple',
        dateFormat: 'Y-m-d',
        minDate: MIN_DATE,
        maxDate: today, 
        inline: true,
        disable: [
            '2025-10-03', // Hardcoded unavailable date
            '2025-09-27'
        ],
        onDayCreate: function (dObj, dStr, fp, dayElem) {
            // Disable unselected dates when limit is reached
            if (fp.selectedDates.length >= MAX_SELECTABLE_DATES &&
                !fp.selectedDates.some(d => d.toDateString() === dayElem.dateObj.toDateString())) {
                dayElem.classList.add('flatpickr-disabled');
            }
        },
        onChange: function (selectedDatesArray, dateStr, instance) {
            // Check if user exceeded the limit
            if (selectedDatesArray.length > MAX_SELECTABLE_DATES) {
                // Remove the last selected date
                selectedDatesArray.pop();
                instance.setDate(selectedDatesArray, false); // false prevents triggering onChange again
                alert(`You can only select up to ${MAX_SELECTABLE_DATES} dates at a time.`);
                return;
            }

            selectedDates = selectedDatesArray.map(d => {
                const date = new Date(d);
                // Ensure we get YYYY-MM-DD format in local timezone
                return date.getFullYear() + '-' +
                    String(date.getMonth() + 1).padStart(2, '0') + '-' +
                    String(date.getDate()).padStart(2, '0');
            });
            updateSelectedDatesDisplay();

            // Redraw calendar to update disabled states
            instance.redraw();
        }
    });
}

// Update the display of selected dates count
function updateSelectedDatesDisplay() {
    const countDisplay = document.getElementById('selected-dates-count');
    const loadBtn = document.getElementById('load-data-btn');

    if (selectedDates.length === 0) {
        countDisplay.textContent = 'No dates selected';
        countDisplay.style.color = '';
        loadBtn.disabled = true;
    } else {
        countDisplay.textContent = `${selectedDates.length} day${selectedDates.length > 1 ? 's' : ''} selected`;
        countDisplay.style.color = 'var(--success-color, #28a745)';
        loadBtn.disabled = false;
    }
}

// Load alerts from API for selected dates
async function loadAlertsForSelectedDates() {
    if (selectedDates.length === 0) {
        alert('Please select at least one date');
        return;
    }

    // Show loading UI
    const loadingStatus = document.getElementById('loading-status');
    const loadingMessage = document.getElementById('loading-message');
    const alertList = document.getElementById('alert-list');
    const loadBtn = document.getElementById('load-data-btn');

    loadingStatus.style.display = 'block';
    loadBtn.disabled = true;
    alertList.innerHTML = '<p class="loading-message">📡 Loading alerts from API...</p>';

    try {
        await loadAlertsFromAPI();

        if (allAlerts.length === 0) {
            alertList.innerHTML = '<p class="loading-message" style="color: var(--warning-color);">⚠️ No alerts found for selected dates.</p>';
            loadingStatus.style.display = 'none';
            // Keep button disabled even on empty results - user must reset to try different dates
            loadBtn.disabled = true;
            return;
        }

        // Enable Stage 2 UI
        enableStage2UI();

        // Initialize filteredAlerts with all loaded data
        filteredAlerts = [...allAlerts];

        // Populate filter dropdowns
        populateSubtypeFilter();
        populateStreetFilter();

        // Apply filters (this will apply the default "verified only" filter if checked)
        applyFilters();

        loadingStatus.style.display = 'none';
        // Keep button disabled after successful load - user must reset to change dates
        loadBtn.disabled = true;

    } catch (error) {
        console.error('❌ Error loading alerts:', error);
        let errorMessage = `Error loading alerts: ${error.message}`;

        alertList.innerHTML = `<p class="loading-message" style="color: var(--danger-color);">❌ ${errorMessage}</p>`;
        loadingStatus.style.display = 'none';
        loadBtn.disabled = false;
    }
}

// Load alerts from API
async function loadAlertsFromAPI() {
    const loadingMessage = document.getElementById('loading-message');
    const totalAlertsCounter = document.getElementById('total-alerts');
    
    try {
        loadingMessage.textContent = 'Authenticating...';
        totalAlertsCounter.textContent = '0';

        // Get authentication token
        const token = await ensureAuthenticated();
        loadingMessage.textContent = `Fetching data for ${selectedDates.length} date(s)...`;

        // Clear existing data
        alertsMap.clear();
        allAlerts = [];

        const datesParam = selectedDates.join(',');
        const url = `${window.API_CONFIG.alertsEndpoint}?dates=${datesParam}`;

        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            signal: AbortSignal.timeout(window.API_CONFIG.timeout)
        });

        if (response.status === 429) {
            throw new Error('⏱️ Too many requests. Please wait a moment and try again.');
        }
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        loadingMessage.textContent = 'Processing data stream...';
        
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        while (true) {
            const { done, value } = await reader.read();
            if (done) {
                break;
            }

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop(); // Keep the last partial line in the buffer

            for (const line of lines) {
                if (line.trim() === '') continue;

                try {
                    const rawAlert = JSON.parse(line);
                    const processedAlert = {
                        id: rawAlert.UUID,
                        UUID: rawAlert.UUID,
                        Type: rawAlert.Type,
                        Subtype: rawAlert.Subtype || '',
                        Street: rawAlert.Street || '',
                        City: rawAlert.City || '',
                        Country: rawAlert.Country,
                        LocationGeo: rawAlert.LocationGeo || { latitude: 0, longitude: 0 },
                        Reliability: rawAlert.Reliability,
                        Confidence: rawAlert.Confidence,
                        PublishTime: new Date(rawAlert.PublishTime).getTime(),
                        ExpireTime: new Date(rawAlert.ExpireTime).getTime(),
                        ScrapeTime: new Date(rawAlert.ScrapeTime).getTime(),
                        ActiveMillis: rawAlert.ActiveMillis,
                        LastVerificationMillis: rawAlert.LastVerificationMillis ? new Date(rawAlert.LastVerificationMillis).getTime() : null,
                        NThumbsUpLast: rawAlert.NThumbsUpLast,
                        ReportRating: rawAlert.ReportRating
                    };

                    // Deduplication
                    const existingAlert = alertsMap.get(processedAlert.UUID);
                    if (!existingAlert || processedAlert.ExpireTime > existingAlert.ExpireTime) {
                        alertsMap.set(processedAlert.UUID, processedAlert);
                    }
                    
                    // Update counter
                    totalAlertsCounter.textContent = alertsMap.size;

                } catch (e) {
                    console.warn('Failed to parse a line of JSONL:', e, 'Line:', line);
                }
            }
        }

        // Process any remaining data in the buffer
        if (buffer.trim() !== '') {
            try {
                const rawAlert = JSON.parse(buffer);
                const processedAlert = {
                    id: rawAlert.UUID,
                    UUID: rawAlert.UUID,
                    Type: rawAlert.Type,
                    Subtype: rawAlert.Subtype || '',
                    Street: rawAlert.Street || '',
                    City: rawAlert.City || '',
                    Country: rawAlert.Country,
                    LocationGeo: rawAlert.LocationGeo || { latitude: 0, longitude: 0 },
                    Reliability: rawAlert.Reliability,
                    Confidence: rawAlert.Confidence,
                    PublishTime: new Date(rawAlert.PublishTime).getTime(),
                    ExpireTime: new Date(rawAlert.ExpireTime).getTime(),
                    ScrapeTime: new Date(rawAlert.ScrapeTime).getTime(),
                    ActiveMillis: rawAlert.ActiveMillis,
                    LastVerificationMillis: rawAlert.LastVerificationMillis ? new Date(rawAlert.LastVerificationMillis).getTime() : null,
                    NThumbsUpLast: rawAlert.NThumbsUpLast,
                    ReportRating: rawAlert.ReportRating
                };
                const existingAlert = alertsMap.get(processedAlert.UUID);
                if (!existingAlert || processedAlert.ExpireTime > existingAlert.ExpireTime) {
                    alertsMap.set(processedAlert.UUID, processedAlert);
                }
                totalAlertsCounter.textContent = alertsMap.size;
            } catch (e) {
                console.warn('Failed to parse final buffer content:', e, 'Buffer:', buffer);
            }
        }

        allAlerts = Array.from(alertsMap.values());
        console.log(`📦 Processed ${allAlerts.length} unique alerts.`);
        totalAlertsCounter.textContent = allAlerts.length;

    } catch (error) {
        console.error('Failed to fetch or process alerts:', error);
        loadingMessage.textContent = `Error: ${error.message}`;
        throw error;
    }
}

// Disable Stage 2 UI elements (Alert Filters)
function disableStage2UI() {
    const filterControls = document.getElementById('filter-controls');
    if (filterControls) {
        const inputs = filterControls.querySelectorAll('input, select, button');
        inputs.forEach(input => {
            input.disabled = true;
            input.style.opacity = '0.5';
        });
    }

    // Disable render button
    const renderBtn = document.getElementById('render-map-btn');
    if (renderBtn) {
        renderBtn.disabled = true;
    }
    const renderSingleDayBtn = document.getElementById('render-map-single-day-btn');
    if (renderSingleDayBtn) {
        renderSingleDayBtn.disabled = true;
    }
}

// Enable Stage 2 UI elements (Alert Filters)
function enableStage2UI() {
    const filterControls = document.getElementById('filter-controls');
    if (filterControls) {
        const inputs = filterControls.querySelectorAll('input, select, button');
        inputs.forEach(input => {
            input.disabled = false;
            input.style.opacity = '1';
        });
    }

    // Enable render button
    const renderBtn = document.getElementById('render-map-btn');
    if (renderBtn) {
        renderBtn.disabled = false;
    }
    const renderSingleDayBtn = document.getElementById('render-map-single-day-btn');
    if (renderSingleDayBtn) {
        renderSingleDayBtn.disabled = false;
    }
}

// Populate subtype filter dropdown
function populateSubtypeFilter() {
    const subtypeDropdown = document.getElementById('subtype-dropdown');

    // Clear existing options except the first placeholder
    subtypeDropdown.innerHTML = '<option value="" disabled selected>Select a subtype to add...</option>';

    // Get unique subtypes from allAlerts
    const subtypes = [...new Set(allAlerts.map(a => a.Subtype))].sort();

    subtypes.forEach(subtype => {
        const option = document.createElement('option');
        option.value = subtype;

        // Special case for empty subtype
        if (subtype === '' || subtype === null || subtype === undefined) {
            option.textContent = "General Police Alert";
            option.value = ''; // Ensure it's an empty string
        } else {
            option.textContent = subtype;
        }

        subtypeDropdown.appendChild(option);
    });
}

// Handle subtype selection from dropdown
function onSubtypeSelected(e) {
    const selectedValue = e.target.value;

    // Check if this subtype is already selected
    if (!selectedSubtypes.includes(selectedValue)) {
        selectedSubtypes.push(selectedValue);
        updateSubtypeTags();
        applyFilters(); // Auto-apply filters when subtype is selected
    }

    // Reset dropdown to placeholder
    e.target.value = '';
}

// Remove a subtype from selection
function removeSubtype(subtype) {
    selectedSubtypes = selectedSubtypes.filter(s => s !== subtype);
    updateSubtypeTags();
    applyFilters(); // Auto-apply filters when subtype is removed
}

// Make removeSubtype available globally for onclick handlers
window.removeSubtype = removeSubtype;

// Update the visual display of selected subtype tags
function updateSubtypeTags() {
    const tagsContainer = document.getElementById('subtype-tags');

    if (selectedSubtypes.length === 0) {
        tagsContainer.innerHTML = '<span class="tag-placeholder">No subtypes selected (showing all)</span>';
    } else {
        tagsContainer.innerHTML = '';

        selectedSubtypes.forEach(subtype => {
            const tag = document.createElement('span');
            tag.className = 'filter-tag';

            // Display text with special handling for empty subtype
            const displayText = (subtype === '' || subtype === null || subtype === undefined)
                ? "General Police"
                : subtype;

            tag.innerHTML = `
                ${displayText}
                <button class="tag-remove" onclick="removeSubtype('${subtype}')" aria-label="Remove ${displayText}">
                    ×
                </button>
            `;

            tagsContainer.appendChild(tag);
        });
    }
}

// Populate street filter dropdown
function populateStreetFilter() {
    const streetDropdown = document.getElementById('street-dropdown');

    // Clear existing options except the first placeholder
    streetDropdown.innerHTML = '<option value="" disabled selected>Select a street to add...</option>';

    // Get unique streets from allAlerts, including empty streets
    const streets = [...new Set(allAlerts.map(a => a.Street || ''))].sort();

    streets.forEach(street => {
        const option = document.createElement('option');
        option.value = street;

        // Special case for empty street
        if (street === '' || street === null || street === undefined) {
            option.textContent = "(No Street)";
            option.value = ''; // Ensure it's an empty string
        } else {
            option.textContent = street;
        }

        streetDropdown.appendChild(option);
    });
}

// Handle street selection from dropdown
function onStreetSelected(e) {
    const selectedValue = e.target.value;

    // Check if this street is already selected
    if (!selectedStreets.includes(selectedValue)) {
        selectedStreets.push(selectedValue);
        updateStreetTags();
        applyFilters(); // Auto-apply filters when street is selected
    }

    // Reset dropdown to placeholder
    e.target.value = '';
}

// Remove a street from selection
function removeStreet(street) {
    selectedStreets = selectedStreets.filter(s => s !== street);
    updateStreetTags();
    applyFilters(); // Auto-apply filters when street is removed
}

// Make removeStreet available globally for onclick handlers
window.removeStreet = removeStreet;

// Handle Hume Highway filter checkbox change
function onHumeHighwayFilterChange(e) {
    const isChecked = e.target.checked;

    if (isChecked) {
        // Find all streets containing "hume" or "federal" (case-insensitive) from loaded alerts
        const humeStreets = [...new Set(
            allAlerts
                .map(a => a.Street || '')
                .filter(street => {
                    const lowerStreet = street.toLowerCase();
                    return lowerStreet.includes('hume') || lowerStreet.includes('federal');
                })
        )];

        if (humeStreets.length === 0) {
            alert('No streets containing "Hume" or "Federal" found in loaded alerts.');
            e.target.checked = false;
            return;
        }

        // Clear existing street selections and add Hume streets
        selectedStreets = [...humeStreets];
        updateStreetTags();

        console.log(`Hume Highway filter activated. Found ${humeStreets.length} streets:`, humeStreets);
    } else {
        // Clear street filters when unchecked
        selectedStreets = [];
        updateStreetTags();
    }

    applyFilters();
}

// Update the visual display of selected street tags
function updateStreetTags() {
    const tagsContainer = document.getElementById('street-tags');

    if (selectedStreets.length === 0) {
        tagsContainer.innerHTML = '<span class="tag-placeholder">No streets selected (showing all)</span>';
    } else {
        tagsContainer.innerHTML = '';

        selectedStreets.forEach(street => {
            const tag = document.createElement('span');
            tag.className = 'filter-tag';

            // Display text with special handling for empty street
            const displayText = (street === '' || street === null || street === undefined)
                ? "(No Street)"
                : street;

            tag.innerHTML = `
                ${displayText}
                <button class="tag-remove" onclick="removeStreet('${street.replace(/'/g, "\'")}')" aria-label="Remove ${displayText}">
                    ×
                </button>
            `;

            tagsContainer.appendChild(tag);
        });
    }
}

// Reset dashboard and clear all data
function resetDashboard() {
    // Clear all data
    allAlerts = [];
    alertsMap.clear();
    filteredAlerts = [];
    selectedDates = [];
    selectedSubtypes = [];
    selectedStreets = [];

    // Clear date picker
    if (flatpickrInstance) {
        flatpickrInstance.clear();
    }

    // Reset UI
    updateSelectedDatesDisplay();
    disableStage2UI();
    clearMap();

    // Clear subtype tags
    const subtypeTags = document.getElementById('subtype-tags');
    if (subtypeTags) {
        subtypeTags.innerHTML = '<span class="tag-placeholder">No subtypes selected (showing all)</span>';
    }

    // Clear subtype dropdown
    const subtypeDropdown = document.getElementById('subtype-dropdown');
    if (subtypeDropdown) {
        subtypeDropdown.innerHTML = '<option value="" disabled selected>Select a subtype to add...</option>';
    }

    // Clear street tags
    const streetTags = document.getElementById('street-tags');
    if (streetTags) {
        streetTags.innerHTML = '<span class="tag-placeholder">No streets selected (showing all)</span>';
    }

    // Clear street dropdown
    const streetDropdown = document.getElementById('street-dropdown');
    if (streetDropdown) {
        streetDropdown.innerHTML = '<option value="" disabled selected>Select a street to add...</option>';
    }

    // Reset verified checkbox to default (checked)
    document.getElementById('verified-only-filter').checked = true;

    // Update statistics
    document.getElementById('total-alerts').textContent = '-';
    document.getElementById('filtered-alerts').textContent = '-';
    document.getElementById('date-range').textContent = '-';
    document.getElementById('avg-reliability').textContent = '-';
    document.getElementById('avg-confidence').textContent = '-';
    document.getElementById('top-city').textContent = '-';

    // Clear alert list
    const alertList = document.getElementById('alert-list');
    alertList.innerHTML = '<p class="loading-message">👋 Dashboard reset. Please select dates and click "Load Data" to begin.</p>';

    console.log('✅ Dashboard reset complete');
}

// Event listeners
function initEventListeners() {
    // Stage 1: Date selection and loading
    document.getElementById('load-data-btn').addEventListener('click', loadAlertsForSelectedDates);
    document.getElementById('reset-dashboard-btn').addEventListener('click', resetDashboard);

    // Stage 2: Filter controls
    document.getElementById('subtype-dropdown').addEventListener('change', onSubtypeSelected);
    document.getElementById('street-dropdown').addEventListener('change', onStreetSelected);
    document.getElementById('verified-only-filter').addEventListener('change', applyFilters);
    document.getElementById('hume-highway-filter').addEventListener('change', onHumeHighwayFilterChange);
    document.getElementById('reset-filters').addEventListener('click', resetFilters);

    // Render button
    document.getElementById('render-map-btn').addEventListener('click', renderAlertsToMap);
    document.getElementById('render-map-single-day-btn').addEventListener('click', renderAlertsToMapSingleDay);

    // Search
    document.getElementById('search-box').addEventListener('input', onSearchInput);
}

// Apply filters
function applyFilters() {
    // Clear map and re-enable render buttons when filters change
    clearMapAndEnableRenderButtons();

    // Start with all loaded alerts
    filteredAlerts = [...allAlerts];

    // Filter by selected subtypes (if any are selected)
    if (selectedSubtypes.length > 0) {
        filteredAlerts = filteredAlerts.filter(a => selectedSubtypes.includes(a.Subtype));
    }

    // Filter by selected streets (if any are selected)
    if (selectedStreets.length > 0) {
        filteredAlerts = filteredAlerts.filter(a => selectedStreets.includes(a.Street || ''));
    }

    // Filter by verified status (if checkbox is checked)
    const showVerifiedOnly = document.getElementById('verified-only-filter').checked;
    if (showVerifiedOnly) {
        filteredAlerts = filteredAlerts.filter(a => {
            const thumbsUp = a.NThumbsUpLast || a.n_thumbs_up_last || 0;
            return thumbsUp !== null && thumbsUp > 0;
        });
    }

    // Sort by publish time (PublishTime is now in milliseconds)
    filteredAlerts.sort((a, b) => a.PublishTime - b.PublishTime);

    updateDisplay();
}

// Helper function to convert alerts to GeoJSON features
// If normalizeDates is true, all dates will be normalized to the same day while preserving hours
function createGeoJSONFromAlerts(alerts, normalizeDates = false) {
    const geojsonFeatures = [];

    // If normalizing dates, use a reference date (e.g., today at midnight)
    let referenceDate = null;
    if (normalizeDates) {
        referenceDate = new Date();
        referenceDate.setHours(0, 0, 0, 0);
    }

    // Helper to normalize a timestamp to the reference date while preserving time
    const normalizeTimestamp = (timestamp) => {
        if (!normalizeDates) return timestamp;

        const date = new Date(timestamp);
        const normalized = new Date(referenceDate);
        normalized.setHours(date.getHours(), date.getMinutes(), date.getSeconds(), date.getMilliseconds());
        return normalized.getTime();
    };

    alerts.forEach(alert => {
        const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
        const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;

        if (!lat || !lng || isNaN(lat) || isNaN(lng)) {
            return; // Skip alerts without valid coordinates
        }

        const publishTime = normalizeTimestamp(alert.PublishTime);
        const expireTime = normalizeTimestamp(alert.ExpireTime);
        const lastVerificationMillis = alert.LastVerificationMillis
            ? normalizeTimestamp(alert.LastVerificationMillis)
            : null;
        const isVerified = lastVerificationMillis !== null && lastVerificationMillis !== undefined;

        if (isVerified) {
            // Create two features: verified period and unverified period

            // 1. Verified period: PublishTime to LastVerificationMillis
            geojsonFeatures.push({
                type: 'Feature',
                properties: {
                    start: publishTime,
                    end: lastVerificationMillis,
                    verified: true,
                    subtype: alert.Subtype,
                    uuid: alert.UUID,
                    street: alert.Street || 'Unknown Location',
                    city: alert.City || 'Unknown City',
                    reliability: alert.Reliability,
                    thumbsUp: alert.NThumbsUpLast,
                    publishTime: formatDateDDMMYYYY(publishTime),
                    expireTime: formatDateDDMMYYYY(expireTime),
                    lastVerification: formatDateDDMMYYYY(lastVerificationMillis)
                },
                geometry: {
                    type: 'Point',
                    coordinates: [lng, lat]
                }
            });

            // 2. Unverified period: LastVerificationMillis to ExpireTime
            geojsonFeatures.push({
                type: 'Feature',
                properties: {
                    start: lastVerificationMillis,
                    end: expireTime,
                    verified: false,
                    subtype: alert.Subtype,
                    uuid: alert.UUID,
                    street: alert.Street || 'Unknown Location',
                    city: alert.City || 'Unknown City',
                    reliability: alert.Reliability,
                    thumbsUp: alert.NThumbsUpLast,
                    publishTime: formatDateDDMMYYYY(publishTime),
                    expireTime: formatDateDDMMYYYY(expireTime),
                    lastVerification: formatDateDDMMYYYY(lastVerificationMillis)
                },
                geometry: {
                    type: 'Point',
                    coordinates: [lng, lat]
                }
            });
        } else {
            // Single feature: PublishTime to ExpireTime (unverified)
            geojsonFeatures.push({
                type: 'Feature',
                properties: {
                    start: publishTime,
                    end: expireTime,
                    verified: false,
                    subtype: alert.Subtype,
                    uuid: alert.UUID,
                    street: alert.Street || 'Unknown Location',
                    city: alert.City || 'Unknown City',
                    reliability: alert.Reliability,
                    thumbsUp: alert.NThumbsUpLast,
                    publishTime: formatDateDDMMYYYY(publishTime),
                    expireTime: formatDateDDMMYYYY(expireTime),
                    lastVerification: null
                },
                geometry: {
                    type: 'Point',
                    coordinates: [lng, lat]
                }
            });
        }
    });

    return geojsonFeatures;
}

// Helper function to scroll to the Alert Map section
function scrollToAlertMap() {
    const mapHeader = document.getElementById('map-header');
    if (mapHeader) {
        mapHeader.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
}

// Render alerts to map with dates normalized to a single day (preserving hours)
function renderAlertsToMapSingleDay() {
    // Disable this button and enable the other render option
    const renderSingleDayBtn = document.getElementById('render-map-single-day-btn');
    const renderTimelineBtn = document.getElementById('render-map-btn');
    if (renderSingleDayBtn) {
        renderSingleDayBtn.disabled = true;
    }
    if (renderTimelineBtn) {
        renderTimelineBtn.disabled = false;
    }

    // Clear existing markers and timeline
    clearMap();

    if (filteredAlerts.length === 0) {
        alert('No alerts to render. Please adjust your filters.');
        return;
    }

    // Step 1: Calculate bounds for centering and zooming
    const coordinates = [];
    filteredAlerts.forEach(alert => {
        const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
        const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;

        if (lat && lng && !isNaN(lat) && !isNaN(lng)) {
            coordinates.push([lat, lng]);
        }
    });

    if (coordinates.length === 0) {
        alert('No valid coordinates found in filtered alerts.');
        return;
    }

    // Step 2: Convert alerts to GeoJSON format with normalized dates
    const geojsonFeatures = createGeoJSONFromAlerts(filteredAlerts, true);

    const geojsonData = {
        type: 'FeatureCollection',
        features: geojsonFeatures
    };

    if (geojsonFeatures.length === 0) {
        alert('No valid features created. Check if alerts have valid timestamps and coordinates.');
        return;
    }

    // Step 3: Create timeline layer with custom styling
    const getInterval = function (feature) {
        return {
            start: feature.properties.start,
            end: feature.properties.end
        };
    };

    timelineLayer = L.timeline(geojsonData, {
        getInterval: getInterval,
        pointToLayer: function (feature, latlng) {
            const props = feature.properties;
            const isVerified = props.verified;
            const subtype = props.subtype || '';

            // Determine color based on verification status
            let color;
            if (isVerified) {
                // Use subtype-specific color for verified alerts
                color = SUBTYPE_COLORS[subtype] || SUBTYPE_COLORS.default;
            } else {
                // Use grayscale for unverified alerts
                color = UNVERIFIED_COLOR;
            }

            // Get emoji for this subtype
            const emoji = SUBTYPE_EMOJIS[subtype] || SUBTYPE_EMOJIS.default;

            // Create custom div icon with colored circle background and emoji on top
            const iconHtml = `
                <div style="
                    width: 32px;
                    height: 32px;
                    border-radius: 50%;
                    background-color: ${color};
                    border: 2px solid ${isVerified ? '#fff' : '#6b7280'};
                    opacity: ${isVerified ? '0.95' : '0.7'};
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 16px;
                    box-shadow: 0 2px 6px rgba(0,0,0,0.3);
                ">
                    ${emoji}
                </div>
            `;

            const customIcon = L.divIcon({
                html: iconHtml,
                className: 'custom-emoji-marker',
                iconSize: [32, 32],
                iconAnchor: [16, 16],
                popupAnchor: [0, -16]
            });

            const marker = L.marker(latlng, { icon: customIcon });

            // Create popup content
            const verifiedBadge = isVerified
                ? '<span style="color: #10b981;">✓ Verified</span>'
                : '<span style="color: #9ca3af;">⊘ Unverified</span>';

            const subtypeBadge = subtype === 'POLICE_WITH_MOBILE_CAMERA'
                ? '📷 Mobile Camera'
                : subtype === 'POLICE_VISIBLE'
                    ? '👮 Visible'
                    : subtype === 'POLICE_HIDING'
                        ? '🕵️ Hiding'
                        : subtype === 'POLICE_ON_BRIDGE'
                            ? '🌉 On Bridge'
                            : subtype === 'POLICE_MOTORCYCLIST'
                                ? '🏍️ Motorcyclist'
                                : '🚓 Police';

            const popupContent = `
                <div style="min-width: 200px;">
                    <h4 style="margin: 0 0 8px 0;">${props.street}</h4>
                    <p style="margin: 4px 0;"><strong>${subtypeBadge}</strong></p>
                    <p style="margin: 4px 0;">${verifiedBadge}</p>
                    <p style="margin: 4px 0; font-size: 12px;"><strong>City:</strong> ${props.city}</p>
                    <p style="margin: 4px 0; font-size: 12px;"><strong>Reliability:</strong> ${props.reliability}/10</p>
                    <p style="margin: 4px 0; font-size: 12px;"><strong>Thumbs Up:</strong> 👍 ${props.thumbsUp}</p>
                    <hr style="margin: 8px 0;">
                    <p style="margin: 4px 0; font-size: 11px;"><strong>Published:</strong> ${props.publishTime}</p>
                    ${props.lastVerification ? `<p style="margin: 4px 0; font-size: 11px;"><strong>Last Verified:</strong> ${props.lastVerification}</p>` : ''}
                    <p style="margin: 4px 0; font-size: 11px;"><strong>Expires:</strong> ${props.expireTime}</p>
                </div>
            `;

            marker.bindPopup(popupContent);
            return marker;
        }
    });

    // Step 4: Create timeline control with the timeline layer
    timelineControl = L.timelineSliderControl({
        formatOutput: function (date) {
            return new Date(date).toLocaleString('en-AU', {
                hour: '2-digit',
                minute: '2-digit'
            });
        },
        enablePlayback: true,
        enableKeyboardControls: true,
        autoPlay: false,
        waitToUpdateMap: false,
        steps: 1000,
        duration: 30000,
        showTicks: true,
        position: 'bottomright'
    });

    timelineControl.addTo(map);
    timelineControl.addTimelines(timelineLayer);
    timelineLayer.addTo(map);

    // Get the time range from the features
    const times = geojsonFeatures.map(f => f.properties.start).filter(t => t && !isNaN(t));
    const minTime = Math.min(...times);
    const maxTime = Math.max(...geojsonFeatures.map(f => f.properties.end).filter(t => t && !isNaN(t)));

    // Set the timeline to start at the earliest time
    if (minTime && !isNaN(minTime)) {
        setTimeout(() => {
            timelineLayer.setTime(minTime);
        }, 200);
    }

    // Step 5: Fit map to show all alerts with padding
    setTimeout(() => {
        // Invalidate map size to ensure proper rendering after expansion
        map.invalidateSize();

        // Calculate bounds for all coordinates
        const bounds = L.latLngBounds(coordinates);

        // Fit map to show all alerts with padding
        map.fitBounds(bounds, {
            padding: [50, 50],
            maxZoom: 15 // Don't zoom in too close even for single alerts
        });

        // Scroll to map section
        scrollToAlertMap();
    }, 100);
}

// Render alerts to map
function renderAlertsToMap() {
    // Disable this button and enable the other render option
    const renderTimelineBtn = document.getElementById('render-map-btn');
    const renderSingleDayBtn = document.getElementById('render-map-single-day-btn');
    if (renderTimelineBtn) {
        renderTimelineBtn.disabled = true;
    }
    if (renderSingleDayBtn) {
        renderSingleDayBtn.disabled = false;
    }

    // Clear existing markers and timeline
    clearMap();

    if (filteredAlerts.length === 0) {
        alert('No alerts to render. Please adjust your filters.');
        return;
    }

    // Step 1: Calculate bounds for centering and zooming
    const coordinates = [];
    filteredAlerts.forEach(alert => {
        const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
        const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;

        if (lat && lng && !isNaN(lat) && !isNaN(lng)) {
            coordinates.push([lat, lng]);
        }
    });

    if (coordinates.length === 0) {
        alert('No valid coordinates found in filtered alerts.');
        return;
    }

    // Step 2: Convert alerts to GeoJSON format (without date normalization)
    const geojsonFeatures = createGeoJSONFromAlerts(filteredAlerts, false);

    const geojsonData = {
        type: 'FeatureCollection',
        features: geojsonFeatures
    };

    if (geojsonFeatures.length === 0) {
        alert('No valid features created. Check if alerts have valid timestamps and coordinates.');
        return;
    }

    // Step 3: Create timeline layer with custom styling
    const getInterval = function (feature) {
        return {
            start: feature.properties.start,
            end: feature.properties.end
        };
    };

    timelineLayer = L.timeline(geojsonData, {
        getInterval: getInterval,
        pointToLayer: function (feature, latlng) {
            const props = feature.properties;
            const isVerified = props.verified;
            const subtype = props.subtype || '';

            // Determine color based on verification status
            let color;
            if (isVerified) {
                // Use subtype-specific color for verified alerts
                color = SUBTYPE_COLORS[subtype] || SUBTYPE_COLORS.default;
            } else {
                // Use grayscale for unverified alerts
                color = UNVERIFIED_COLOR;
            }

            // Get emoji for this subtype
            const emoji = SUBTYPE_EMOJIS[subtype] || SUBTYPE_EMOJIS.default;

            // Create custom div icon with colored circle background and emoji on top
            const iconHtml = `
                <div style="
                    width: 32px;
                    height: 32px;
                    border-radius: 50%;
                    background-color: ${color};
                    border: 2px solid ${isVerified ? '#fff' : '#6b7280'};
                    opacity: ${isVerified ? '0.95' : '0.7'};
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 16px;
                    box-shadow: 0 2px 6px rgba(0,0,0,0.3);
                ">
                    ${emoji}
                </div>
            `;

            const customIcon = L.divIcon({
                html: iconHtml,
                className: 'custom-emoji-marker',
                iconSize: [32, 32],
                iconAnchor: [16, 16],
                popupAnchor: [0, -16]
            });

            const marker = L.marker(latlng, { icon: customIcon });

            // Create popup content
            const verifiedBadge = isVerified
                ? '<span style="color: #10b981;">✓ Verified</span>'
                : '<span style="color: #9ca3af;">⊘ Unverified</span>';

            const subtypeBadge = subtype === 'POLICE_WITH_MOBILE_CAMERA'
                ? '📷 Mobile Camera'
                : subtype === 'POLICE_VISIBLE'
                    ? '👮 Visible'
                    : subtype === 'POLICE_HIDING'
                        ? '🕵️ Hiding'
                        : subtype === 'POLICE_ON_BRIDGE'
                            ? '🌉 On Bridge'
                            : subtype === 'POLICE_MOTORCYCLIST'
                                ? '🏍️ Motorcyclist'
                                : '🚓 Police';

            const popupContent = `
                <div style="min-width: 200px;">
                    <h4 style="margin: 0 0 8px 0;">${props.street}</h4>
                    <p style="margin: 4px 0;"><strong>${subtypeBadge}</strong></p>
                    <p style="margin: 4px 0;">${verifiedBadge}</p>
                    <p style="margin: 4px 0; font-size: 12px;"><strong>City:</strong> ${props.city}</p>
                    <p style="margin: 4px 0; font-size: 12px;"><strong>Reliability:</strong> ${props.reliability}/10</p>
                    <p style="margin: 4px 0; font-size: 12px;"><strong>Thumbs Up:</strong> 👍 ${props.thumbsUp}</p>
                    <hr style="margin: 8px 0;">
                    <p style="margin: 4px 0; font-size: 11px;"><strong>Published:</strong> ${props.publishTime}</p>
                    ${props.lastVerification ? `<p style="margin: 4px 0; font-size: 11px;"><strong>Last Verified:</strong> ${props.lastVerification}</p>` : ''}
                    <p style="margin: 4px 0; font-size: 11px;"><strong>Expires:</strong> ${props.expireTime}</p>
                </div>
            `;

            marker.bindPopup(popupContent);
            return marker;
        }
    });

    // Step 4: Create timeline control with the timeline layer
    timelineControl = L.timelineSliderControl({
        formatOutput: function (date) {
            return new Date(date).toLocaleString('en-AU', {
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        },
        enablePlayback: true,
        enableKeyboardControls: true,
        autoPlay: false,
        waitToUpdateMap: false,
        steps: 1000,
        duration: 30000,
        showTicks: true,
        position: 'bottomright'
    });

    timelineControl.addTo(map);
    timelineControl.addTimelines(timelineLayer);
    timelineLayer.addTo(map);

    // Get the time range from the features
    const times = geojsonFeatures.map(f => f.properties.start).filter(t => t && !isNaN(t));
    const minTime = Math.min(...times);
    const maxTime = Math.max(...geojsonFeatures.map(f => f.properties.end).filter(t => t && !isNaN(t)));

    // Set the timeline to start at the earliest time
    if (minTime && !isNaN(minTime)) {
        setTimeout(() => {
            timelineLayer.setTime(minTime);
        }, 200);
    }

    // Step 5: Fit map to show all alerts with padding
    setTimeout(() => {
        // Invalidate map size to ensure proper rendering after expansion
        map.invalidateSize();

        // Calculate bounds for all coordinates
        const bounds = L.latLngBounds(coordinates);

        // Fit map to show all alerts with padding
        map.fitBounds(bounds, {
            padding: [50, 50],
            maxZoom: 15 // Don't zoom in too close even for single alerts
        });

        // Scroll to map section
        scrollToAlertMap();
    }, 100);
}

// Reset filters (Stage 2 only, doesn't clear loaded data)
function resetFilters() {
    // Clear subtype selections
    selectedSubtypes = [];
    updateSubtypeTags();

    // Clear street selections
    selectedStreets = [];
    updateStreetTags();

    // Reset verified checkbox to default (checked)
    document.getElementById('verified-only-filter').checked = true;

    // Reset Hume Highway filter to default (unchecked)
    document.getElementById('hume-highway-filter').checked = false;

    applyFilters();
}

// Update display
function updateDisplay() {
    updateStatistics();
    updateAlertList();
    // Note: Map is no longer automatically updated - use "Render Alerts to Map" button
}

// Update statistics
function updateStatistics() {
    document.getElementById('total-alerts').textContent = allAlerts.length.toLocaleString();
    document.getElementById('filtered-alerts').textContent = filteredAlerts.length.toLocaleString();

    if (filteredAlerts.length > 0) {
        const dates = filteredAlerts.map(a => new Date(a.PublishTime)).filter(d => !isNaN(d));
        if (dates.length > 0) {
            const minDate = new Date(Math.min(...dates));
            const maxDate = new Date(Math.max(...dates));
            document.getElementById('date-range').textContent =
                `${formatDateDDMMYYYY(minDate, false)} - ${formatDateDDMMYYYY(maxDate, false)}`;
        }

        const avgReliability = (filteredAlerts.reduce((sum, a) => sum + a.Reliability, 0) / filteredAlerts.length).toFixed(1);
        document.getElementById('avg-reliability').textContent = avgReliability;

        // Calculate average confidence
        const avgConfidence = (filteredAlerts.reduce((sum, a) => sum + (a.Confidence || 0), 0) / filteredAlerts.length).toFixed(1);
        document.getElementById('avg-confidence').textContent = avgConfidence;

        // Find top city
        const cityCounts = {};
        filteredAlerts.forEach(a => {
            if (a.City) {
                cityCounts[a.City] = (cityCounts[a.City] || 0) + 1;
            }
        });
        const topCity = Object.entries(cityCounts).sort((a, b) => b[1] - a[1])[0];
        document.getElementById('top-city').textContent = topCity ? `${topCity[0]} (${topCity[1]})` : '-';
    } else {
        document.getElementById('date-range').textContent = '-';
        document.getElementById('avg-reliability').textContent = '-';
        document.getElementById('avg-confidence').textContent = '-';
        document.getElementById('top-city').textContent = '-';
    }
}

// Clear map and re-enable render buttons (called when filters change)
function clearMapAndEnableRenderButtons() {
    clearMap();

    // Re-enable both render buttons
    const renderBtn = document.getElementById('render-map-btn');
    if (renderBtn) {
        renderBtn.disabled = false;
    }
    const renderSingleDayBtn = document.getElementById('render-map-single-day-btn');
    if (renderSingleDayBtn) {
        renderSingleDayBtn.disabled = false;
    }
}

// Clear all markers from map
function clearMap() {
    // Remove old markers
    markers.forEach(marker => map.removeLayer(marker));
    markers = [];

    // Remove timeline layer if it exists
    if (timelineLayer) {
        map.removeLayer(timelineLayer);
        timelineLayer = null;
    }

    // Remove timeline control if it exists
    if (timelineControl) {
        map.removeControl(timelineControl);
        timelineControl = null;
    }
}

// Update map (no longer automatically renders markers)
function updateMap() {
    // Map rendering is now manual via the "Render Alerts to Map" button
    // This function is kept for potential future updates
}

// Update alert list
function updateAlertList() {
    const alertList = document.getElementById('alert-list');

    if (filteredAlerts.length === 0) {
        alertList.innerHTML = '<p class="loading-message">No alerts match the current filters.</p>';
        return;
    }

    alertList.innerHTML = '';

    filteredAlerts.forEach((alert, index) => {
        const alertItem = createAlertItem(alert, index);
        alertList.appendChild(alertItem);
    });
}

// Create alert list item
function createAlertItem(alert, index) {
    const div = document.createElement('div');
    div.className = 'alert-item';
    div.dataset.index = index;

    const badgeClass = alert.Subtype === 'POLICE_WITH_MOBILE_CAMERA' ? 'mobile-camera' : 'standard';
    const badgeText = alert.Subtype === 'POLICE_WITH_MOBILE_CAMERA' ? '📷 Mobile Camera' : '🚓 Police';

    div.innerHTML = `
        <div class="alert-header">
            <div class="alert-title">${alert.Street || 'Unknown Location'}</div>
            <span class="alert-badge ${badgeClass}">${badgeText}</span>
        </div>
        <div class="alert-details">
            <div class="alert-detail"><strong>City:</strong> ${alert.City}</div>
            <div class="alert-detail"><strong>Time:</strong> ${formatDateDDMMYYYY(alert.PublishTime)}</div>
            <div class="alert-detail"><strong>Reliability:</strong> ${alert.Reliability}/10</div>
            <div class="alert-detail"><strong>Thumbs Up:</strong> 👍 ${alert.NThumbsUpLast}</div>
        </div>
    `;

    div.addEventListener('click', () => {
        // Highlight selected alert
        document.querySelectorAll('.alert-item').forEach(item => item.classList.remove('selected'));
        div.classList.add('selected');

        // Pan to location on map only if map has been rendered (has markers)
        if (markers.length > 0 && markers[index]) {
            const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
            const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;
            if (lat && lng) {
                map.setView([lat, lng], 14);
                markers[index].openPopup();
            }
        }
    });

    return div;
}



// Search functionality - filters alert list by street, city, or UUID
function onSearchInput(e) {
    const searchTerm = e.target.value.toLowerCase();

    if (!searchTerm) {
        updateAlertList();
        return;
    }

    const searchResults = filteredAlerts.filter(alert =>
        (alert.Street && alert.Street.toLowerCase().includes(searchTerm)) ||
        (alert.City && alert.City.toLowerCase().includes(searchTerm)) ||
        (alert.UUID && alert.UUID.toLowerCase().includes(searchTerm))
    );

    const alertList = document.getElementById('alert-list');
    alertList.innerHTML = '';

    if (searchResults.length === 0) {
        alertList.innerHTML = '<p class="loading-message">No alerts match your search.</p>';
        return;
    }

    searchResults.forEach((alert, index) => {
        const originalIndex = filteredAlerts.indexOf(alert);
        const alertItem = createAlertItem(alert, originalIndex);
        alertList.appendChild(alertItem);
    });
}

// Map control functions