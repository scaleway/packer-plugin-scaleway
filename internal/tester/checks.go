package tester

import (
	"context"
	"testing"
)

// PackerCheck represents a check for a scaleway resource
type PackerCheck interface {
	Check(ctx context.Context) error
	CheckName() string
}

// PackerCleanup represents a cleanup function for a scaleway resource
type PackerCleanup interface {
	Cleanup(ctx context.Context, t *testing.T) error
}
