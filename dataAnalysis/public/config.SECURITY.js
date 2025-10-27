// Frontend Security Configuration Guide
// =====================================

// 1. NEVER expose sensitive API keys in frontend code
// 2. Use environment-specific configs (dev vs prod)
// 3. Implement proper error handling to avoid exposing internals

// ⚠️ IMPORTANT: If API key authentication is enabled on backend,
// update this configuration to include the API key

// Map configuration
window.MAP_CONFIG = {
    center: [-34.5, 150.0], // Sydney-Canberra midpoint
    zoom: 8,
    minZoom: 6,
    maxZoom: 18
};

// Marker colors based on alert subtype (for verified alerts)
window.SUBTYPE_COLORS = {
    'POLICE_WITH_MOBILE_CAMERA': '#8bf5ffff', // Cyan
    'POLICE_VISIBLE': '#0051ffff', // Blue
    'POLICE_HIDING': '#ff0000ff', // Red
    '': '#0051ffff', // Blue for general police alerts (empty subtype)
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


// ============================================
// SECURITY NOTES FOR FRONTEND DEVELOPERS
// ============================================

/*
 * 1. CORS PROTECTION
 *    - Your API only accepts requests from whitelisted domains
 *    - Ensure your domain is added to the backend's allowedOrigins list
 *    - Contact backend admin to whitelist new domains
 *
 * 2. RATE LIMITING
 *    - API allows 30 requests per minute per IP
 *    - Implement request queuing/debouncing in frontend to avoid hitting limits
 *    - Show user-friendly error messages when rate limited
 *
 * 3. REQUEST SIZE LIMITS
 *    - Max 10 dates per request
 *    - Max 50 subtypes per filter
 *    - Max 100 streets per filter
 *    - Frontend already enforces 7-day limit (good!)
 *
 * 4. API KEY AUTHENTICATION (if enabled)
 *    - API key must be sent in X-API-Key header
 *    - Alternative: Authorization: Bearer <key>
 *    - Update loadAlertsFromAPI() to include key in headers
 *
 * 5. ERROR HANDLING
 *    - Handle 429 (Too Many Requests) gracefully
 *    - Handle 401 (Unauthorized) if API key is invalid
 *    - Handle 400 (Bad Request) for validation errors
 *    - Show user-friendly messages, don't expose internals
 *
 * 6. HTTPS ONLY
 *    - Always use HTTPS in production
 *    - Mixed content (HTTP + HTTPS) will be blocked by browsers
 *
 * 7. CSP (Content Security Policy)
 *    - Be aware of CSP headers from API
 *    - May need to adjust if embedding in iframes
 */

// Example: Enhanced fetch with API key support
/*
async function loadAlertsFromAPI() {
    const headers = {
        'Content-Type': 'application/json'
    };
    
    // Add API key if configured
    if (window.API_CONFIG.apiKey) {
        headers['X-API-Key'] = window.API_CONFIG.apiKey;
    }
    
    const response = await fetch(window.API_CONFIG.alertsEndpoint, {
        method: 'POST',
        headers: headers,
        body: JSON.stringify(requestBody),
        signal: AbortSignal.timeout(window.API_CONFIG.timeout)
    });
    
    // Handle rate limiting
    if (response.status === 429) {
        throw new Error('Too many requests. Please wait a moment and try again.');
    }
    
    // Handle authentication errors
    if (response.status === 401) {
        throw new Error('Invalid API key. Please contact support.');
    }
    
    // ... rest of code
}
*/
