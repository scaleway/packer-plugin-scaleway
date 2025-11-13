package tester

import (
	"context"
)

// PackerCheck represents a check for a scaleway resource
type PackerCheck interface {
	Check(ctx context.Context) error
}

// PackerCleanup represents a cleanup function for a scaleway resource
type PackerCleanup interface {
	Cleanup(ctx context.Context) error
}
