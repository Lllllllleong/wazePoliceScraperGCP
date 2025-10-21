// Firebase Configuration (using compat SDK)
// No imports needed - loaded via script tags in HTML

window.firebaseConfig = {
    apiKey: "AIzaSyAQEJlaNkj9kBmxUU7L1JYTn9LlbgTAWQc",
    authDomain: "wazepolicescrapergcp.firebaseapp.com",
    projectId: "wazepolicescrapergcp",
    storageBucket: "wazepolicescrapergcp.firebasestorage.app",
    messagingSenderId: "807773831037",
    appId: "1:807773831037:web:b80c7ceaa8306276ad5614",
    measurementId: "G-6RXKVX9N9H"
};

// Firestore collection name
window.COLLECTION_NAME = "police_alerts";

// API Configuration
// Set this to your deployed Cloud Function URL
// For local development, use: http://localhost:8080/api/alerts
// For production, use your Cloud Run URL: https://your-service-url.run.app/api/alerts
window.API_CONFIG = {
    alertsEndpoint: "https://waze-alerts-api-u6cjbro2iq-uc.a.run.app/api/alerts",
    useAPI: true, // Set to true to use API, false to use direct Firestore access (legacy)
    timeout: 30000 // Request timeout in milliseconds
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
    'POLICE_WITH_MOBILE_CAMERA': '#8bf5ffff', // Red
    'POLICE_VISIBLE': '#0051ffff', // Blue
    'POLICE_HIDING': '#ff0000ff', // Purple
    '': '#0051ffff', // Blue for general police alerts (empty subtype)
    'default': '#64748b' // Gray fallback
};

// Unverified alert color (grayscale)
window.UNVERIFIED_COLOR = '#9ca3af'; // Gray-400
