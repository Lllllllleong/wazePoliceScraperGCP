// Package storage provides data persistence abstractions for Firestore and GCS.
package storage

import (
	"context"
	"errors"
)

// MockFirebaseAuthClient is a mock implementation of FirebaseAuthClient for testing.
type MockFirebaseAuthClient struct {
	// VerifyIDTokenFunc is called when VerifyIDToken is invoked.
	// If nil, returns an error indicating invalid token.
	VerifyIDTokenFunc func(ctx context.Context, idToken string) (*FirebaseToken, error)

	// CallLog tracks calls made to the mock for verification.
	CallLog struct {
		VerifyIDTokenCalls int
		LastToken          string
	}
}

// VerifyIDToken implements FirebaseAuthClient.VerifyIDToken.
func (m *MockFirebaseAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*FirebaseToken, error) {
	m.CallLog.VerifyIDTokenCalls++
	m.CallLog.LastToken = idToken

	if m.VerifyIDTokenFunc != nil {
		return m.VerifyIDTokenFunc(ctx, idToken)
	}
	return nil, errors.New("mock: token verification not configured")
}

// Ensure MockFirebaseAuthClient implements FirebaseAuthClient.
var _ FirebaseAuthClient = (*MockFirebaseAuthClient)(nil)
