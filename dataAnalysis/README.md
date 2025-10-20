# Waze Police Alert Analysis Web Application

This is a Firebase-hosted web application for analyzing Waze police alert data.

## Features

- **Interactive Map Viewer**: Displays police alerts on a Leaflet.js map (similar to Folium)
- **Time-based Filtering**: Filter alerts by date and time ranges
- **Advanced Filters**: Filter by alert type, city, reliability, and thumbs up count
- **Timeline Player**: Playback alerts chronologically with adjustable speed
- **Search**: Search alerts by street, city, or UUID
- **Export**: Export filtered data to JSONL format
- **Real-time Data**: Reads alert data directly from Firestore

## Setup Instructions

### 1. Install Firebase CLI

```bash
npm install -g firebase-tools
```

### 2. Login to Firebase

```bash
firebase login
```

### 3. Configure Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Select your project (or create a new one)
3. Get your Firebase configuration:
   - Go to Project Settings > General
   - Scroll down to "Your apps" section
   - Click on the web app icon (</>)
   - Copy the Firebase configuration object

4. Update `public/config.js` with your Firebase configuration:
   ```javascript
   const firebaseConfig = {
       apiKey: "YOUR_API_KEY",
       authDomain: "YOUR_PROJECT_ID.firebaseapp.com",
       projectId: "YOUR_PROJECT_ID",
       storageBucket: "YOUR_PROJECT_ID.appspot.com",
       messagingSenderId: "YOUR_MESSAGING_SENDER_ID",
       appId: "YOUR_APP_ID"
   };
   ```

5. Update `.firebaserc` with your project ID:
   ```json
   {
     "projects": {
       "default": "your-actual-project-id"
     }
   }
   ```

### 4. Enable Firebase Authentication

1. In Firebase Console, go to Authentication
2. Click "Get Started"
3. Enable "Anonymous" sign-in method (for read-only access)

### 5. Configure Firestore Security Rules

The `firestore.rules` file is already configured to allow read access for authenticated users. Deploy it:

```bash
firebase deploy --only firestore:rules
```

### 6. Deploy the Application

```bash
cd dataAnalysis
firebase deploy --only hosting
```

After deployment, Firebase will provide a URL where your app is hosted.

## Local Development

To test the application locally before deploying:

```bash
cd dataAnalysis
firebase serve
```

Then open http://localhost:5000 in your browser.

## Application Structure

```
dataAnalysis/
├── public/
│   ├── index.html       # Main HTML page
│   ├── styles.css       # Styling
│   ├── config.js        # Firebase configuration
│   └── app.js           # Main application logic
├── october_alerts.jsonl # Sample exported data
├── firebase.json        # Firebase hosting configuration
├── firestore.rules      # Firestore security rules
├── .firebaserc          # Firebase project reference
└── README.md           # This file
```

## Using the Application

### Time Range Filter
- Set start and end dates/times to analyze specific periods
- Click "Apply Time Filter" to update the display

### Alert Filters
- **Subtype**: Filter by standard police or mobile camera alerts
- **City**: Filter alerts by city
- **Min Reliability**: Set minimum reliability threshold (0-10)
- **Min Thumbs Up**: Filter by minimum community verification

### Timeline Player
- Click "▶ Play" to animate alerts chronologically
- Adjust playback speed (0.5x to 10x)
- Use slider to jump to specific points in time
- Alerts are highlighted on the map as they appear

### Map Interaction
- Click on markers to view alert details
- Click on alert items in the list to jump to their location on the map
- Red markers = Mobile camera alerts
- Blue markers = Standard police alerts

### Search & Export
- Use the search box to find specific alerts
- Click "Export Filtered Data" to download current filtered results as JSONL

## Firestore Data Structure

The application expects alerts in Firestore with the following structure:

```javascript
{
  uuid: "string",
  type: "POLICE",
  subtype: "POLICE_WITH_MOBILE_CAMERA" | "",
  street: "string",
  city: "string",
  location: {
    latitude: number,
    longitude: number
  },
  reliability: number (0-10),
  confidence: number,
  publish_time: timestamp,
  n_thumbs_up_last: number
}
```

## Troubleshooting

### "Failed to authenticate" error
- Ensure Anonymous authentication is enabled in Firebase Console
- Check that your Firebase configuration is correct in `config.js`

### "No alerts match the current filters"
- Verify that your Firestore collection name is "police_alerts"
- Check Firestore rules allow read access
- Ensure data exists in Firestore

### Map not displaying
- Check browser console for errors
- Ensure internet connection (Leaflet tiles require internet)
- Verify alert data has valid latitude/longitude coordinates

## Next Steps

The filtering logic placeholders in `app.js` are marked with `// TODO:` comments. The basic filtering is implemented, but you can enhance it further:

- Add more sophisticated time interval analysis
- Implement heatmap visualization
- Add clustering for better performance with many markers
- Create statistical charts and graphs
- Add export to other formats (CSV, GeoJSON)
