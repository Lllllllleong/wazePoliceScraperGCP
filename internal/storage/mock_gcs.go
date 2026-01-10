// Package storage provides data persistence abstractions for Firestore and GCS.
package storage

import (
	"context"
	"io"
)

// MockGCSClient is a mock implementation of GCSClient for testing.
type MockGCSClient struct {
	// BucketFunc is called when Bucket is invoked.
	// If nil, returns a MockGCSBucketHandle with default behavior.
	BucketFunc func(name string) GCSBucketHandle
}

// Bucket implements GCSClient.Bucket.
func (m *MockGCSClient) Bucket(name string) GCSBucketHandle {
	if m.BucketFunc != nil {
		return m.BucketFunc(name)
	}
	return &MockGCSBucketHandle{}
}

// Ensure MockGCSClient implements GCSClient.
var _ GCSClient = (*MockGCSClient)(nil)

// MockGCSBucketHandle is a mock implementation of GCSBucketHandle for testing.
type MockGCSBucketHandle struct {
	// ObjectFunc is called when Object is invoked.
	// If nil, returns a MockGCSObjectHandle with default behavior.
	ObjectFunc func(name string) GCSObjectHandle
}

// Object implements GCSBucketHandle.Object.
func (m *MockGCSBucketHandle) Object(name string) GCSObjectHandle {
	if m.ObjectFunc != nil {
		return m.ObjectFunc(name)
	}
	return &MockGCSObjectHandle{}
}

// Ensure MockGCSBucketHandle implements GCSBucketHandle.
var _ GCSBucketHandle = (*MockGCSBucketHandle)(nil)

// MockGCSObjectHandle is a mock implementation of GCSObjectHandle for testing.
type MockGCSObjectHandle struct {
	// NewReaderFunc is called when NewReader is invoked.
	// If nil, returns an error indicating object not found.
	NewReaderFunc func(ctx context.Context) (io.ReadCloser, error)

	// AttrsFunc is called when Attrs is invoked.
	// If nil, returns ErrObjectNotExist (object does not exist).
	AttrsFunc func(ctx context.Context) (*GCSObjectAttrs, error)

	// NewWriterFunc is called when NewWriter is invoked.
	// If nil, returns a MockGCSWriter with default behavior.
	NewWriterFunc func(ctx context.Context) GCSWriter
}

// NewReader implements GCSObjectHandle.NewReader.
func (m *MockGCSObjectHandle) NewReader(ctx context.Context) (io.ReadCloser, error) {
	if m.NewReaderFunc != nil {
		return m.NewReaderFunc(ctx)
	}
	return nil, ErrObjectNotExist
}

// Attrs implements GCSObjectHandle.Attrs.
func (m *MockGCSObjectHandle) Attrs(ctx context.Context) (*GCSObjectAttrs, error) {
	if m.AttrsFunc != nil {
		return m.AttrsFunc(ctx)
	}
	return nil, ErrObjectNotExist
}

// NewWriter implements GCSObjectHandle.NewWriter.
func (m *MockGCSObjectHandle) NewWriter(ctx context.Context) GCSWriter {
	if m.NewWriterFunc != nil {
		return m.NewWriterFunc(ctx)
	}
	return &MockGCSWriter{}
}

// Ensure MockGCSObjectHandle implements GCSObjectHandle.
var _ GCSObjectHandle = (*MockGCSObjectHandle)(nil)

// MockGCSWriter is a mock implementation of GCSWriter for testing.
type MockGCSWriter struct {
	// WriteFunc is called when Write is invoked.
	// If nil, buffers the data and returns success.
	WriteFunc func(p []byte) (n int, err error)

	// CloseFunc is called when Close is invoked.
	// If nil, returns nil (success).
	CloseFunc func() error

	// Written stores all data written to this writer.
	Written []byte
}

// Write implements io.Writer.
func (m *MockGCSWriter) Write(p []byte) (n int, err error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	m.Written = append(m.Written, p...)
	return len(p), nil
}

// Close implements io.Closer.
func (m *MockGCSWriter) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Ensure MockGCSWriter implements GCSWriter.
var _ GCSWriter = (*MockGCSWriter)(nil)

// ErrObjectNotExist is a sentinel error indicating the object does not exist in GCS.
// This mirrors cloud.google.com/go/storage.ErrObjectNotExist for testing.
var ErrObjectNotExist = &objectNotExistError{}

type objectNotExistError struct{}

func (e *objectNotExistError) Error() string {
	return "storage: object doesn't exist"
}
