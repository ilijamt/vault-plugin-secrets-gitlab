package backend

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

// Backend implements the logical.Backend interface for the GitLab secrets plugin.
// It manages GitLab client connections, plugin configuration flags, and coordinates
// the registration of API paths.
//
// The Backend is safe for concurrent use through its internal synchronization primitives.
type Backend struct {
	*framework.Backend

	clients   sync.Map
	flags     flags.Flags
	lockFlags sync.RWMutex
}

// UpdateFlags atomically updates the plugin's runtime configuration flags.
// This method is safe for concurrent use.
func (b *Backend) UpdateFlags(f flags.Flags) {
	b.lockFlags.Lock()
	defer b.lockFlags.Unlock()
	b.flags = f
}

// GetFlags returns a copy of the current plugin configuration flags.
// This method is safe for concurrent use.
func (b *Backend) GetFlags() flags.Flags {
	b.lockFlags.RLock()
	defer b.lockFlags.RUnlock()
	return b.flags
}

// Factory creates a logical.Factory function that constructs a new Backend instance.
//
// The factory pattern allows Vault to instantiate the backend with the proper
// lifecycle and configuration.
//
// Parameters:
//   - flags: Initial plugin configuration flags that control runtime behavior
//   - version: Plugin version string to be reported to Vault
//   - paths: Variable number of Path implementations to register with the backend
//
// Returns a logical.Factory function that Vault will invoke to create the backend.
// The factory will return an error if any path registration fails or if the backend
// setup encounters an issue.
func Factory(flags flags.Flags, version string, paths ...Path) logical.Factory {
	return func(ctx context.Context, cfg *logical.BackendConfig) (_ logical.Backend, err error) {
		var b *Backend
		b = &Backend{
			Backend: &framework.Backend{
				BackendType:    logical.TypeLogical,
				Help:           strings.TrimSpace(Help),
				RunningVersion: version,
				PeriodicFunc: func(ctx context.Context, req *logical.Request) error {
					var errs []error
					for _, path := range paths {
						errs = append(errs, path.PeriodicFunc(ctx, b.GetFlags(), path, req))
					}
					return errors.Join(errs...)
				},
				Invalidate: func(ctx context.Context, key string) {
					for _, path := range paths {
						path.InvalidateFunc(ctx, b.GetFlags(), path, key)
					}
				},
			},
		}

		b.UpdateFlags(flags)

		var errs []error
		for _, path := range paths {
			errs = append(errs, path.Register(b))
		}
 
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}

		if err := b.Backend.Setup(ctx, cfg); err != nil {
			return nil, err
		}

		return b, nil
	}
}
