package backend

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// PathProvider provides framework paths to register with the backend.
type PathProvider interface {
	Name() string
	Paths() []*framework.Path
}

// PeriodicHandler is optionally implemented by PathProviders that need periodic work.
// The backend checks WriteSafeReplicationState() centrally before dispatching —
// handlers are only called when writes are safe.
type PeriodicHandler interface {
	PeriodicFunc(ctx context.Context, req *logical.Request) error
}

// InvalidateHandler is optionally implemented by PathProviders that need
// to react to storage key invalidation events.
type InvalidateHandler interface {
	Invalidate(ctx context.Context, key string)
}
