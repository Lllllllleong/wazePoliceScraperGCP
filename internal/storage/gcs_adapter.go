// Package storage provides data persistence abstractions for Firestore and GCS.
package storage

import (
	"context"
	"io"

	gcs "cloud.google.com/go/storage"
)

// GCSClientAdapter wraps a real GCS client to implement the GCSClient interface.
type GCSClientAdapter struct {
	Client *gcs.Client
}

// Bucket implements GCSClient.Bucket.
func (a *GCSClientAdapter) Bucket(name string) GCSBucketHandle {
	return &GCSBucketHandleAdapter{Handle: a.Client.Bucket(name)}
}

// Ensure GCSClientAdapter implements GCSClient.
var _ GCSClient = (*GCSClientAdapter)(nil)

// GCSBucketHandleAdapter wraps a real GCS bucket handle to implement the GCSBucketHandle interface.
type GCSBucketHandleAdapter struct {
	Handle *gcs.BucketHandle
}

// Object implements GCSBucketHandle.Object.
func (a *GCSBucketHandleAdapter) Object(name string) GCSObjectHandle {
	return &GCSObjectHandleAdapter{Handle: a.Handle.Object(name)}
}

// Ensure GCSBucketHandleAdapter implements GCSBucketHandle.
var _ GCSBucketHandle = (*GCSBucketHandleAdapter)(nil)

// GCSObjectHandleAdapter wraps a real GCS object handle to implement the GCSObjectHandle interface.
type GCSObjectHandleAdapter struct {
	Handle *gcs.ObjectHandle
}

// NewReader implements GCSObjectHandle.NewReader.
func (a *GCSObjectHandleAdapter) NewReader(ctx context.Context) (io.ReadCloser, error) {
	return a.Handle.NewReader(ctx)
}

// Attrs implements GCSObjectHandle.Attrs.
func (a *GCSObjectHandleAdapter) Attrs(ctx context.Context) (*GCSObjectAttrs, error) {
	attrs, err := a.Handle.Attrs(ctx)
	if err != nil {
		return nil, err
	}
	return &GCSObjectAttrs{
		Name: attrs.Name,
		Size: attrs.Size,
	}, nil
}

// NewWriter implements GCSObjectHandle.NewWriter.
func (a *GCSObjectHandleAdapter) NewWriter(ctx context.Context) GCSWriter {
	return a.Handle.NewWriter(ctx)
}

// Ensure GCSObjectHandleAdapter implements GCSObjectHandle.
var _ GCSObjectHandle = (*GCSObjectHandleAdapter)(nil)

// IsObjectNotExist checks if an error indicates that a GCS object does not exist.
// This works for both the real GCS error and our mock error.
func IsObjectNotExist(err error) bool {
	if err == gcs.ErrObjectNotExist {
		return true
	}
	if err == ErrObjectNotExist {
		return true
	}
	// Check error message as fallback
	if err != nil && err.Error() == "storage: object doesn't exist" {
		return true
	}
	return false
}
