package tester

import (
	"context"
)

// PackerCheck represents a check for a scaleway resource
type PackerCheck interface {
	Check(ctx context.Context) error
}
