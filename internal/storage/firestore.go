package storage

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

// FirestoreClient handles all Firestore operations
type FirestoreClient struct {
	client         *firestore.Client
	collectionName string
}

// NewFirestoreClient creates a new Firestore client
func NewFirestoreClient(ctx context.Context, projectID, collectionName string) (*FirestoreClient, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	// Default collection name if not provided
	if collectionName == "" {
		collectionName = "police_alerts"
	}

	return &FirestoreClient{
		client:         client,
		collectionName: collectionName,
	}, nil
}

// Close closes the Firestore client
func (fc *FirestoreClient) Close() error {
	return fc.client.Close()
}
