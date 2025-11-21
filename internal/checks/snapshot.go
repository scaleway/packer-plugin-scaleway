package checks

import "context"

type SnapshotCheck interface {
	Check(context.Context) error
}
