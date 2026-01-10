// Package storage provides data persistence abstractions for Firestore and GCS.
package storage

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

// FirebaseAuthClientAdapter wraps a real Firebase Auth client to implement the FirebaseAuthClient interface.
type FirebaseAuthClientAdapter struct {
	Client *auth.Client
}

// VerifyIDToken implements FirebaseAuthClient.VerifyIDToken.
func (a *FirebaseAuthClientAdapter) VerifyIDToken(ctx context.Context, idToken string) (*FirebaseToken, error) {
	token, err := a.Client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return &FirebaseToken{UID: token.UID}, nil
}

// Ensure FirebaseAuthClientAdapter implements FirebaseAuthClient.
var _ FirebaseAuthClient = (*FirebaseAuthClientAdapter)(nil)
