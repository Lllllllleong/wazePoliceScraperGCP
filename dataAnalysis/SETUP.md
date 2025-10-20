# Firebase Setup Guide for Existing GCP Project

## Prerequisites
- Existing GCP Project: `wazepolicescrapergcp`
- Firestore already enabled and containing `police_alerts` collection
- Firebase CLI installed: `npm install -g firebase-tools`

## Step-by-Step Setup

### 1. Add Firebase to Existing GCP Project

🔗 Go to: https://console.firebase.google.com/

1. Click **"Add project"**
2. **IMPORTANT**: In the dropdown, select your existing project: `wazepolicescrapergcp`
3. Click "Continue" through the setup wizard
4. Firebase will be added to your existing GCP project (no new project created!)

### 2. Enable Firebase Authentication

1. In Firebase Console, navigate to **Build → Authentication**
2. Click **"Get Started"**
3. Go to **"Sign-in method"** tab
4. Click **"Anonymous"**
5. Toggle **"Enable"** and click **"Save"**

   ✅ This allows your web app to authenticate anonymously for read-only Firestore access

### 3. Register Your Web Application

1. In Firebase Console, click the ⚙️ **Settings** icon → **Project settings**
2. Scroll to **"Your apps"** section (at the bottom)
3. Click the **Web** icon `</>`
4. Enter app nickname: `Waze Alert Analyzer`
5. ✅ Check "Also set up Firebase Hosting"
6. Click **"Register app"**

### 4. Get Firebase Configuration

After registering, you'll see a configuration object like this:

```javascript
const firebaseConfig = {
  apiKey: "AIza...",
  authDomain: "wazepolicescrapergcp.firebaseapp.com",
  projectId: "wazepolicescrapergcp",
  storageBucket: "wazepolicescrapergcp.appspot.com",
  messagingSenderId: "123456789",
  appId: "1:123456789:web:abcdef"
};
```

### 5. Update Your Configuration

Copy the Firebase config and update `dataAnalysis/public/config.js`:

```bash
# Open the file
notepad dataAnalysis/public/config.js
```

Replace the placeholder values with your actual Firebase configuration.

### 6. Update Firestore Security Rules

Deploy the security rules to allow authenticated read access:

```bash
cd dataAnalysis
firebase login
firebase deploy --only firestore:rules
```

The rules in `firestore.rules` are already configured to:
- ✅ Allow READ for authenticated users (including anonymous)
- ❌ Deny WRITE (only your backend can write)

### 7. Test Locally

Before deploying, test the app locally:

```bash
cd dataAnalysis
firebase serve
```

Open http://localhost:5000 in your browser.

### 8. Deploy to Firebase Hosting

Once everything works:

```bash
firebase deploy --only hosting
```

Firebase will give you a URL like: `https://wazepolicescrapergcp.web.app`

## Firestore Collection Structure

Your existing Firestore collection should work as-is. The app expects:

**Collection**: `police_alerts`

**Document fields** (from your Go backend):
```
uuid (string)
type (string) - "POLICE"
subtype (string) - "" or "POLICE_WITH_MOBILE_CAMERA"
street (string)
city (string)
country (string)
location (map)
  ├─ latitude (number)
  └─ longitude (number)
reliability (number)
confidence (number)
publish_time (timestamp)
scrape_time (timestamp)
n_thumbs_up_last (number)
report_rating (number)
```

The web app handles both field naming conventions (camelCase and PascalCase).

## Troubleshooting

### "Failed to authenticate"
- Ensure Anonymous auth is enabled in Firebase Console
- Check browser console for detailed errors

### "Permission denied" errors
- Deploy Firestore rules: `firebase deploy --only firestore:rules`
- Verify rules allow read access for authenticated users

### "Collection not found"
- Verify collection name is exactly `police_alerts` in Firestore
- Check GCP Project ID matches in all config files

### Map not loading
- Check that alerts have valid `location.latitude` and `location.longitude`
- Open browser DevTools → Console to see errors

## Project Structure

```
dataAnalysis/
├── public/
│   ├── index.html         # Main app
│   ├── styles.css         # Styles
│   ├── config.js          # ⚠️ UPDATE THIS with Firebase config
│   └── app.js             # App logic
├── firebase.json          # ✅ Already configured
├── firestore.rules        # ✅ Already configured
├── .firebaserc            # ✅ Already has your project ID
└── README.md             # Usage documentation
```

## Next Steps After Deployment

1. Share the Firebase Hosting URL with your team
2. Monitor usage in Firebase Console → Analytics
3. Adjust Firestore rules if you need more granular access control
4. Consider adding Firebase Analytics for user insights

## Cost Considerations

Firebase has a generous free tier:
- **Hosting**: 10 GB storage, 360 MB/day transfer
- **Firestore**: 50K reads/day, 20K writes/day
- **Authentication**: Unlimited anonymous auth

Your use case should stay well within free tier limits! 🎉
