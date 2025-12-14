// Environment detection
const isLocalhost = window.location.hostname === 'localhost' || 
                    window.location.hostname === '127.0.0.1' || 
                    window.location.port === '5000';

// API Configuration
window.API_CONFIG = {
    alertsEndpoint: isLocalhost 
        ? "http://localhost:8080/police_alerts"
        : "https://alerts-service-807773831037.us-central1.run.app/police_alerts",
    timeout: 60000
};

// Firebase Configuration
window.FIREBASE_CONFIG = {
    apiKey: "AIzaSyBi6kHNLu8BM2CjObGtj5YOwLBjFaLW9xI",
    authDomain: "wazepolicescrapergcp.firebaseapp.com",
    projectId: "wazepolicescrapergcp",
    storageBucket: "wazepolicescrapergcp.firebasestorage.app",
    messagingSenderId: "807773831037",
    appId: "1:807773831037:web:b26c7a5bd2e2b4b3f1c4d9"
};

// Emulator configuration (only for local development)
window.USE_FIREBASE_EMULATOR = isLocalhost;
window.FIREBASE_EMULATOR_HOST = "localhost:9099";

// Environment logging
if (isLocalhost) {
    console.log('🔧 Development Mode: localhost:8080 + Firebase Emulator');
} else {
    console.log('🌐 Production Mode: Cloud Run + Firebase Auth');
}

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
