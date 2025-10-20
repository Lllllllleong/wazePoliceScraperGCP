// Initialize Firebase
firebase.initializeApp(firebaseConfig);
const db = firebase.firestore();
const auth = firebase.auth();

// Global state
let allAlerts = []; // Will store alerts as array
let alertsMap = new Map(); // Map for deduplication using UUID as key
let filteredAlerts = [];
let map = null;
let markers = [];
let timelineInterval = null;
let currentTimelineIndex = 0;
let isPlaying = false;
let selectedDates = []; // Array of selected date strings
let flatpickrInstance = null;
let selectedSubtypes = []; // Array of selected subtypes for filtering
let selectedCities = []; // Array of selected cities for filtering

// Constants for date range
// TODO: Ideally, MIN_DATE should be derived by querying the earliest "publish_time" 
// Timestamp field in the entire Firestore collection, but we've hardcoded it for now
const MIN_DATE = '2025-10-03';
const MAX_DATE = new Date().toISOString().split('T')[0]; // Current date

// Initialize the application
document.addEventListener('DOMContentLoaded', async () => {
    console.log('Initializing Waze Police Alert Analysis app...');

    initMap();
    initDatePicker();
    initEventListeners();
    disableStage2UI(); // Disable alert filters until data is loaded

    // Show welcome message
    const alertList = document.getElementById('alert-list');
    alertList.innerHTML = '<p class="loading-message">� Welcome! Please select dates above and click "Load Data" to begin.</p>';

    // Sign in anonymously for Firestore read access
    try {
        await auth.signInAnonymously();
        console.log('✅ Authenticated with Firebase');
    } catch (error) {
        console.error('❌ Authentication error:', error);
        showError(`Failed to authenticate: ${error.message}<br><br>Please check your Firebase configuration and ensure Anonymous authentication is enabled.`);
    }
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
    flatpickrInstance = flatpickr('#date-picker', {
        mode: 'multiple',
        dateFormat: 'Y-m-d',
        minDate: MIN_DATE,
        maxDate: MAX_DATE,
        inline: false,
        onChange: function (selectedDatesArray, dateStr, instance) {
            selectedDates = selectedDatesArray.map(d => {
                const date = new Date(d);
                // Ensure we get YYYY-MM-DD format in local timezone
                return date.getFullYear() + '-' +
                    String(date.getMonth() + 1).padStart(2, '0') + '-' +
                    String(date.getDate()).padStart(2, '0');
            });
            updateSelectedDatesDisplay();
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

// Load alerts from Firestore for selected dates
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
    alertList.innerHTML = '<p class="loading-message">📡 Loading alerts from Firestore...</p>';

    try {
        // Sort dates chronologically
        const sortedDates = [...selectedDates].sort();
        console.log('Loading alerts for dates:', sortedDates);

        let totalFetched = 0;
        let duplicatesSkipped = 0;

        // Iterate through each selected date
        for (let i = 0; i < sortedDates.length; i++) {
            const dateStr = sortedDates[i];
            loadingMessage.textContent = `Loading day ${i + 1} of ${sortedDates.length}: ${dateStr}...`;

            // Create start and end timestamps for the day
            const dayStart = new Date(dateStr + 'T00:00:00');
            const dayEnd = new Date(dateStr + 'T23:59:59.999');

            console.log(`Querying alerts for ${dateStr} (${dayStart.toISOString()} to ${dayEnd.toISOString()})`);

            // Query alerts where:
            // - expire_time >= start of day (alert is still active at start of day)
            // - publish_time <= end of day (alert was published by end of day)
            const snapshot = await db.collection(COLLECTION_NAME)
                .where('publish_time', '<=', dayEnd)
                .where('expire_time', '>=', dayStart)
                .get();

            console.log(`📦 Received ${snapshot.docs.length} documents for ${dateStr}`);

            // Process documents and deduplicate using Map
            snapshot.docs.forEach(doc => {
                const data = doc.data();
                const uuid = data.uuid || data.UUID;

                if (!uuid) {
                    console.warn('Alert without UUID:', doc.id);
                    return;
                }

                // Check if we already have this alert
                if (alertsMap.has(uuid)) {
                    duplicatesSkipped++;
                    return;
                }

                // Add new alert to map
                const alert = {
                    id: doc.id,
                    UUID: uuid,
                    Type: data.type || data.Type,
                    Subtype: data.subtype || data.Subtype || '',
                    Street: data.street || data.Street || '',
                    City: data.city || data.City || '',
                    Country: data.country || data.Country || '',
                    LocationGeo: data.location || data.LocationGeo || { latitude: 0, longitude: 0 },
                    Reliability: data.reliability || data.Reliability || 0,
                    Confidence: data.confidence || data.Confidence || 0,
                    PublishTime: data.publish_time || data.PublishTime,
                    ExpireTime: data.expire_time || data.ExpireTime,
                    ScrapeTime: data.scrape_time || data.ScrapeTime,
                    NThumbsUpLast: data.n_thumbs_up_last || data.NThumbsUpLast || 0,
                    ReportRating: data.report_rating || data.ReportRating || 0
                };

                alertsMap.set(uuid, alert);
                totalFetched++;
            });
        }

        // Convert map to array
        allAlerts = Array.from(alertsMap.values());

        console.log(`✅ Loaded ${totalFetched} new alerts, skipped ${duplicatesSkipped} duplicates`);
        console.log(`📊 Total unique alerts in memory: ${allAlerts.length}`);

        if (allAlerts.length === 0) {
            alertList.innerHTML = '<p class="loading-message" style="color: var(--warning-color);">⚠️ No alerts found for selected dates.</p>';
            loadingStatus.style.display = 'none';
            loadBtn.disabled = false;
            return;
        }

        // Enable Stage 2 UI
        enableStage2UI();

        // Initialize filteredAlerts with all loaded data
        filteredAlerts = [...allAlerts];

        // Populate filter dropdowns
        populateSubtypeFilter();
        populateCityFilter();

        // Update display with all alerts initially
        updateDisplay();

        loadingStatus.style.display = 'none';
        loadBtn.disabled = false;

    } catch (error) {
        console.error('❌ Error loading alerts:', error);
        let errorMessage = `Error loading alerts: ${error.message}`;

        if (error.code === 'permission-denied') {
            errorMessage += '<br><br>Please ensure:<br>1. Firestore rules are deployed<br>2. Anonymous authentication is enabled';
        } else if (error.code === 'failed-precondition') {
            errorMessage += '<br><br>Missing Firestore index. Check console for index creation link.';
        }

        alertList.innerHTML = `<p class="loading-message" style="color: var(--danger-color);">❌ ${errorMessage}</p>`;
        loadingStatus.style.display = 'none';
        loadBtn.disabled = false;
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
            option.textContent = "'' (General Police Alert)";
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
                ? "'' (General Police)"
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

// Populate city filter dropdown
function populateCityFilter() {
    const cityDropdown = document.getElementById('city-dropdown');

    // Clear existing options except the first placeholder
    cityDropdown.innerHTML = '<option value="" disabled selected>Select a city to add...</option>';

    // Get unique cities from allAlerts
    const cities = [...new Set(allAlerts.map(a => a.City).filter(c => c))].sort();

    cities.forEach(city => {
        const option = document.createElement('option');
        option.value = city;
        option.textContent = city;
        cityDropdown.appendChild(option);
    });
}

// Handle city selection from dropdown
function onCitySelected(e) {
    const selectedValue = e.target.value;

    // Check if this city is already selected
    if (!selectedCities.includes(selectedValue)) {
        selectedCities.push(selectedValue);
        updateCityTags();
        applyFilters(); // Auto-apply filters when city is selected
    }

    // Reset dropdown to placeholder
    e.target.value = '';
}

// Remove a city from selection
function removeCity(city) {
    selectedCities = selectedCities.filter(c => c !== city);
    updateCityTags();
    applyFilters(); // Auto-apply filters when city is removed
}

// Make removeCity available globally for onclick handlers
window.removeCity = removeCity;

// Update the visual display of selected city tags
function updateCityTags() {
    const tagsContainer = document.getElementById('city-tags');

    if (selectedCities.length === 0) {
        tagsContainer.innerHTML = '<span class="tag-placeholder">No cities selected (showing all)</span>';
    } else {
        tagsContainer.innerHTML = '';

        selectedCities.forEach(city => {
            const tag = document.createElement('span');
            tag.className = 'filter-tag';

            tag.innerHTML = `
                ${city}
                <button class="tag-remove" onclick="removeCity('${city.replace(/'/g, "\\'")}'')" aria-label="Remove ${city}">
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
    selectedCities = [];

    // Clear date picker
    if (flatpickrInstance) {
        flatpickrInstance.clear();
    }

    // Reset UI
    updateSelectedDatesDisplay();
    disableStage2UI();
    clearMap();
    resetTimeline();

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

    // Clear city tags
    const cityTags = document.getElementById('city-tags');
    if (cityTags) {
        cityTags.innerHTML = '<span class="tag-placeholder">No cities selected (showing all)</span>';
    }

    // Clear city dropdown
    const cityDropdown = document.getElementById('city-dropdown');
    if (cityDropdown) {
        cityDropdown.innerHTML = '<option value="" disabled selected>Select a city to add...</option>';
    }

    // Reset verified checkbox to default (checked)
    document.getElementById('verified-only-filter').checked = true;

    // Update statistics
    document.getElementById('total-alerts').textContent = '-';
    document.getElementById('filtered-alerts').textContent = '-';
    document.getElementById('date-range').textContent = '-';
    document.getElementById('avg-reliability').textContent = '-';
    document.getElementById('mobile-cameras').textContent = '-';
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
    document.getElementById('city-dropdown').addEventListener('change', onCitySelected);
    document.getElementById('verified-only-filter').addEventListener('change', applyFilters);
    document.getElementById('reset-filters').addEventListener('click', resetFilters);

    // Timeline controls
    document.getElementById('play-pause-btn').addEventListener('click', togglePlayPause);
    document.getElementById('reset-timeline-btn').addEventListener('click', resetTimeline);
    document.getElementById('timeline-range').addEventListener('input', onTimelineSliderChange);

    // Map controls
    document.getElementById('reset-zoom-btn').addEventListener('click', resetMapZoom);

    // Search
    document.getElementById('search-box').addEventListener('input', onSearchInput);

    // Export
    document.getElementById('export-btn').addEventListener('click', exportFilteredData);
}

// Apply filters
function applyFilters() {
    // Start with all loaded alerts
    filteredAlerts = [...allAlerts];

    // Filter by selected subtypes (if any are selected)
    if (selectedSubtypes.length > 0) {
        filteredAlerts = filteredAlerts.filter(a => selectedSubtypes.includes(a.Subtype));
    }

    // Filter by selected cities (if any are selected)
    if (selectedCities.length > 0) {
        filteredAlerts = filteredAlerts.filter(a => selectedCities.includes(a.City));
    }

    // Filter by verified status (if checkbox is checked)
    const showVerifiedOnly = document.getElementById('verified-only-filter').checked;
    if (showVerifiedOnly) {
        filteredAlerts = filteredAlerts.filter(a => {
            const thumbsUp = a.NThumbsUpLast || a.n_thumbs_up_last || 0;
            return thumbsUp !== null && thumbsUp > 0;
        });
    }

    // Sort by publish time
    filteredAlerts.sort((a, b) => new Date(a.PublishTime) - new Date(b.PublishTime));

    updateDisplay();
}

// Reset filters (Stage 2 only, doesn't clear loaded data)
function resetFilters() {
    // Clear subtype selections
    selectedSubtypes = [];
    updateSubtypeTags();

    // Clear city selections
    selectedCities = [];
    updateCityTags();

    // Reset verified checkbox to default (checked)
    document.getElementById('verified-only-filter').checked = true;

    applyFilters();
}

// Update display
function updateDisplay() {
    updateStatistics();
    updateMap();
    updateAlertList();
    initializeTimeline();
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
                `${minDate.toLocaleDateString()} - ${maxDate.toLocaleDateString()}`;
        }

        const avgReliability = (filteredAlerts.reduce((sum, a) => sum + a.Reliability, 0) / filteredAlerts.length).toFixed(1);
        document.getElementById('avg-reliability').textContent = avgReliability;

        // Count mobile cameras
        const mobileCameras = filteredAlerts.filter(a => a.Subtype === 'POLICE_WITH_MOBILE_CAMERA').length;
        document.getElementById('mobile-cameras').textContent = mobileCameras.toLocaleString();

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
        document.getElementById('mobile-cameras').textContent = '-';
        document.getElementById('top-city').textContent = '-';
    }
}

// Clear all markers from map
function clearMap() {
    markers.forEach(marker => map.removeLayer(marker));
    markers = [];
}

// Update map
function updateMap() {
    // Clear existing markers
    markers.forEach(marker => map.removeLayer(marker));
    markers = [];

    // Add markers for filtered alerts
    filteredAlerts.forEach(alert => {
        const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
        const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;

        if (lat && lng) {
            const color = MARKER_COLORS[alert.Subtype] || MARKER_COLORS.default;

            const marker = L.circleMarker([lat, lng], {
                radius: 8,
                fillColor: color,
                color: '#fff',
                weight: 2,
                opacity: 1,
                fillOpacity: 0.8
            }).addTo(map);

            const popupContent = `
                <div class="popup-content">
                    <div class="popup-title">${alert.Street || 'Unknown Street'}</div>
                    <div class="popup-details">
                        <div><strong>City:</strong> ${alert.City}</div>
                        <div><strong>Type:</strong> ${alert.Subtype || 'Standard Police'}</div>
                        <div><strong>Time:</strong> ${new Date(alert.PublishTime).toLocaleString()}</div>
                        <div><strong>Reliability:</strong> ${alert.Reliability}/10</div>
                        <div><strong>Thumbs Up:</strong> 👍 ${alert.NThumbsUpLast}</div>
                    </div>
                </div>
            `;

            marker.bindPopup(popupContent);
            markers.push(marker);
        }
    });

    // Fit map to markers if there are any
    if (markers.length > 0) {
        const group = L.featureGroup(markers);
        map.fitBounds(group.getBounds(), { padding: [50, 50] });
    }
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
            <div class="alert-detail"><strong>Time:</strong> ${new Date(alert.PublishTime).toLocaleString()}</div>
            <div class="alert-detail"><strong>Reliability:</strong> ${alert.Reliability}/10</div>
            <div class="alert-detail"><strong>Thumbs Up:</strong> 👍 ${alert.NThumbsUpLast}</div>
        </div>
    `;

    div.addEventListener('click', () => {
        // Highlight selected alert
        document.querySelectorAll('.alert-item').forEach(item => item.classList.remove('selected'));
        div.classList.add('selected');

        // Pan to marker on map
        const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
        const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;
        if (lat && lng) {
            map.setView([lat, lng], 14);
            markers[index].openPopup();
        }
    });

    return div;
}

// Timeline functions
function initializeTimeline() {
    const slider = document.getElementById('timeline-range');
    if (filteredAlerts.length === 0) {
        slider.max = 0;
        slider.value = 0;
        currentTimelineIndex = 0;
        updateTimelineDisplay();
        return;
    }

    slider.max = filteredAlerts.length - 1;
    slider.value = 0;
    currentTimelineIndex = 0;

    // Reset any playing state
    if (isPlaying) {
        togglePlayPause();
    }

    updateTimelineDisplay();
    console.log(`📅 Timeline initialized with ${filteredAlerts.length} alerts`);
}

function togglePlayPause() {
    if (filteredAlerts.length === 0) return;

    isPlaying = !isPlaying;
    const btn = document.getElementById('play-pause-btn');

    if (isPlaying) {
        btn.textContent = '⏸ Pause';
        btn.classList.add('playing');
        playTimeline();
    } else {
        btn.textContent = '▶ Play';
        btn.classList.remove('playing');
        stopTimeline();
    }
}

function playTimeline() {
    const speed = parseFloat(document.getElementById('playback-speed').value);
    const interval = 1000 / speed; // milliseconds per step

    timelineInterval = setInterval(() => {
        if (currentTimelineIndex < filteredAlerts.length - 1) {
            currentTimelineIndex++;
            document.getElementById('timeline-range').value = currentTimelineIndex;
            updateTimelineDisplay();
            highlightAlertInList(currentTimelineIndex);
        } else {
            // Reached the end, stop playback
            togglePlayPause();
        }
    }, interval);
}

function stopTimeline() {
    if (timelineInterval) {
        clearInterval(timelineInterval);
        timelineInterval = null;
    }
}

function resetTimeline() {
    stopTimeline();
    if (isPlaying) {
        isPlaying = false;
        const btn = document.getElementById('play-pause-btn');
        btn.textContent = '▶ Play';
        btn.classList.remove('playing');
    }
    currentTimelineIndex = 0;
    document.getElementById('timeline-range').value = 0;

    // Clear all existing markers
    clearMarkerHighlights();

    updateTimelineDisplay();
}

function onTimelineSliderChange(e) {
    currentTimelineIndex = parseInt(e.target.value);
    updateTimelineDisplay();
    highlightAlertInList(currentTimelineIndex);
}

function updateTimelineDisplay() {
    if (filteredAlerts.length === 0 || currentTimelineIndex >= filteredAlerts.length) {
        document.getElementById('current-time').textContent = '-';
        document.getElementById('timeline-progress').textContent = '0 / 0 alerts';
        return;
    }

    const currentAlert = filteredAlerts[currentTimelineIndex];
    const currentTime = new Date(currentAlert.PublishTime).toLocaleString();
    document.getElementById('current-time').textContent = currentTime;
    document.getElementById('timeline-progress').textContent =
        `${currentTimelineIndex + 1} / ${filteredAlerts.length} alerts`;

    // Update map to show cumulative alerts up to current timeline position
    updateTimelineMarkers();
}

function updateTimelineMarkers() {
    // Clear existing markers
    markers.forEach(marker => map.removeLayer(marker));
    markers = [];

    // Show all alerts up to current timeline index
    for (let i = 0; i <= currentTimelineIndex; i++) {
        const alert = filteredAlerts[i];
        const lat = alert.LocationGeo.latitude || alert.LocationGeo.y;
        const lng = alert.LocationGeo.longitude || alert.LocationGeo.x;

        if (lat && lng) {
            const color = MARKER_COLORS[alert.Subtype] || MARKER_COLORS.default;
            const isCurrent = (i === currentTimelineIndex);

            const marker = L.circleMarker([lat, lng], {
                radius: isCurrent ? 12 : 6,
                fillColor: color,
                color: isCurrent ? '#fff' : color,
                weight: isCurrent ? 3 : 1,
                opacity: 1,
                fillOpacity: isCurrent ? 1 : 0.5,
                className: isCurrent ? 'current-marker' : ''
            }).addTo(map);

            const popupContent = `
                <div class="popup-content">
                    <div class="popup-title">${alert.Street || 'Unknown Street'}</div>
                    <div class="popup-details">
                        <div><strong>City:</strong> ${alert.City}</div>
                        <div><strong>Type:</strong> ${alert.Subtype || 'Standard Police'}</div>
                        <div><strong>Time:</strong> ${new Date(alert.PublishTime).toLocaleString()}</div>
                        <div><strong>Reliability:</strong> ${alert.Reliability}/10</div>
                        <div><strong>Thumbs Up:</strong> 👍 ${alert.NThumbsUpLast}</div>
                    </div>
                </div>
            `;

            marker.bindPopup(popupContent);

            // Open popup for current alert
            if (isCurrent) {
                marker.openPopup();
                map.setView([lat, lng], Math.max(map.getZoom(), 11), { animate: true });
            }

            markers.push(marker);
        }
    }
}

function highlightAlertInList(index) {
    // Remove previous highlights
    document.querySelectorAll('.alert-item').forEach(item => {
        item.classList.remove('timeline-active');
    });

    // Highlight current alert
    const alertItems = document.querySelectorAll('.alert-item');
    if (alertItems[index]) {
        alertItems[index].classList.add('timeline-active');
        // Scroll to the highlighted alert
        alertItems[index].scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
}

function clearMarkerHighlights() {
    markers.forEach(marker => {
        marker.setStyle({
            radius: 8,
            weight: 2,
            fillOpacity: 0.8
        });
    });
}

// Search functionality
function onSearchInput(e) {
    // TODO: Implement search logic
    // Placeholder: Filter alert list by search term
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

// Export filtered data
function exportFilteredData() {
    if (filteredAlerts.length === 0) {
        alert('No data to export!');
        return;
    }

    const dataStr = filteredAlerts.map(alert => JSON.stringify(alert)).join('\n');
    const blob = new Blob([dataStr], { type: 'application/jsonl' });
    const url = URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = url;
    a.download = `police_alerts_${new Date().toISOString().split('T')[0]}.jsonl`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

// Error handling
function showError(message) {
    const alertList = document.getElementById('alert-list');
    alertList.innerHTML = `<p class="loading-message" style="color: var(--danger-color);">❌ ${message}</p>`;
}

// Map control functions
function resetMapZoom() {
    if (markers.length > 0) {
        const group = L.featureGroup(markers);
        map.fitBounds(group.getBounds(), { padding: [50, 50] });
    } else {
        map.setView(MAP_CONFIG.center, MAP_CONFIG.zoom);
    }
}
