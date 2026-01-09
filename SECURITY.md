# Security Considerations

This document explains the security model of this project and clarifies which configuration values are safe to be public.

## Firebase Configuration (Public by Design)

The Firebase configuration in `dataAnalysis/public/config.js` contains API keys and project identifiers that may appear sensitive but are **intentionally public**:

```javascript
window.FIREBASE_CONFIG = {
    apiKey: "...",
    authDomain: "...",
    projectId: "...",
    // ...
};
```

**Why this is safe:**
- Firebase API keys are designed to be embedded in client-side code
- They identify your Firebase project to Google's servers, similar to a public identifier
- Security is enforced through Firebase Security Rules, not by hiding the API key
- The `apiKey` restricts which Firebase services can be accessed, not who can access them

**Security controls in place:**
- Firebase Anonymous Authentication with per-user rate limiting (30 req/min)
- CORS restrictions on the backend API
- Read-only public access (no write operations exposed)

## GCP Project Identifiers

The following identifiers appear throughout the codebase and are **not secrets**:
- **Project ID** (`wazepolicescrapergcp`): A public identifier for the GCP project
- **Cloud Run URLs**: Public endpoints by design
- **Service account names**: Identity references, not credentials

## Sensitive Items (Not in Repository)

The following are **not** committed to this repository:
- Service account key files (`.json` credentials)
- Environment files with secrets (`.env`)
- Terraform state files containing sensitive outputs
- Personal access tokens or API secrets

## Reporting Security Issues

If you discover a security vulnerability, please open an issue or contact the maintainer directly.

