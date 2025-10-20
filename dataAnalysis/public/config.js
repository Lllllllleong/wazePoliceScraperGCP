// Firebase Configuration (using compat SDK)
// No imports needed - loaded via script tags in HTML

const firebaseConfig = {
    apiKey: "AIzaSyAQEJlaNkj9kBmxUU7L1JYTn9LlbgTAWQc",
    authDomain: "wazepolicescrapergcp.firebaseapp.com",
    projectId: "wazepolicescrapergcp",
    storageBucket: "wazepolicescrapergcp.firebasestorage.app",
    messagingSenderId: "807773831037",
    appId: "1:807773831037:web:b80c7ceaa8306276ad5614",
    measurementId: "G-6RXKVX9N9H"
};

// Firestore collection name
const COLLECTION_NAME = "police_alerts";

// Map configuration
const MAP_CONFIG = {
    center: [-34.5, 150.0], // Sydney-Canberra midpoint
    zoom: 8,
    minZoom: 6,
    maxZoom: 18
};

// Marker colors based on alert type
const MARKER_COLORS = {
    'POLICE_WITH_MOBILE_CAMERA': '#ef4444', // Red
    'POLICE': '#2563eb', // Blue
    'default': '#64748b' // Gray
};
