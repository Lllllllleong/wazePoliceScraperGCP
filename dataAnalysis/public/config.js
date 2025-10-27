// API Configuration
window.API_CONFIG = {
    alertsEndpoint: "https://alerts-service-807773831037.us-central1.run.app/police_alerts",
    timeout: 60000 // Increased timeout for streaming
};

// Map configuration
window.MAP_CONFIG = {
    center: [-34.5, 150.0], // Sydney-Canberra midpoint
    zoom: 8,
    minZoom: 6,
    maxZoom: 18
};

// Marker colors based on alert subtype (for verified alerts)
window.SUBTYPE_COLORS = {
    'POLICE_WITH_MOBILE_CAMERA': '#4cccffff', // Cyan
    'POLICE_VISIBLE': '#0004ffff', // Blue
    'POLICE_HIDING': '#ff0000ff', // Red
    '': '#0004ffff', // Blue for general police alerts (empty subtype)
    'default': '#64748b' // Gray fallback
};

// Emoji icons for each subtype (displayed on top of colored circles)
window.SUBTYPE_EMOJIS = {
    'POLICE_WITH_MOBILE_CAMERA': '📷',  // Camera icon
    'POLICE_VISIBLE': '👮',             // Police officer
    'POLICE_HIDING': '🕵️',              // Detective
    '': '',                             // No emoji for general alerts
    'default': '🚓'                     // Fallback police car
};

// Unverified alert color (grayscale)
window.UNVERIFIED_COLOR = '#9ca3af'; // Gray-400
