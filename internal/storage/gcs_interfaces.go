// Package storage provides data persistence abstractions for Firestore and GCS.
package storage

import (
	"context"
	"io"
)

// GCSBucketHandle represents a handle to a GCS bucket.
// This interface enables dependency injection and mocking for testing.
type GCSBucketHandle interface {
	// Object returns a handle to an object in the bucket.
	Object(name string) GCSObjectHandle
}

// GCSObjectHandle represents a handle to a GCS object.
type GCSObjectHandle interface {
	// NewReader creates a new Reader to read the object's contents.
	// Returns an error if the object does not exist or cannot be read.
	NewReader(ctx context.Context) (io.ReadCloser, error)
}

// GCSClient represents a client for interacting with Google Cloud Storage.
// This interface wraps the bucket access pattern used by the alerts service.
type GCSClient interface {
	// Bucket returns a handle to the specified bucket.
	Bucket(name string) GCSBucketHandle
}

// FirebaseAuthClient represents a Firebase Auth client for token verification.
// This interface enables dependency injection and mocking for testing.
type FirebaseAuthClient interface {
	// VerifyIDToken verifies a Firebase ID token and returns the decoded token.
	VerifyIDToken(ctx context.Context, idToken string) (*FirebaseToken, error)
}

// FirebaseToken represents a decoded Firebase ID token.
type FirebaseToken struct {
	// UID is the user ID from the token.
	UID string
}
