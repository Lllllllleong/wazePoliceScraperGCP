package storage

import (
	"errors"
	"fmt"
	"testing"

	gcs "cloud.google.com/go/storage"
)

func TestIsObjectNotExist(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "direct SDK sentinel",
			err:  gcs.ErrObjectNotExist,
			want: true,
		},
		{
			name: "wrapped SDK sentinel (GCS SDK v1.57+ regression case)",
			err:  fmt.Errorf("rpc: %w", gcs.ErrObjectNotExist),
			want: true,
		},
		{
			name: "double-wrapped SDK sentinel",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", gcs.ErrObjectNotExist)),
			want: true,
		},
		{
			name: "direct mock sentinel",
			err:  ErrObjectNotExist,
			want: true,
		},
		{
			name: "wrapped mock sentinel",
			err:  fmt.Errorf("wrapped: %w", ErrObjectNotExist),
			want: true,
		},
		{
			name: "unrelated error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "bare string matching old message — no longer matches (string fallback removed)",
			err:  errors.New("storage: object doesn't exist"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsObjectNotExist(tt.err)
			if got != tt.want {
				t.Errorf("IsObjectNotExist(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
