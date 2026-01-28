package backend

import (
	"context"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

// Path defines the interface for registering and managing API paths in the GitLab secrets backend.
//
// Each implementation of this interface represents a specific API endpoint or group of related
// endpoints in the plugin. Implementations are responsible for defining their path structure,
// registration logic, and periodic/invalidation behavior.
//
// The interface supports the plugin's lifecycle management by providing hooks for periodic
// operations (like token rotation) and cache invalidation.
type Path interface {
	// Path returns the path.
	Path() string

	// Register registers this path with the backend.
	//
	// This method is called during backend initialization to add the path's
	// operations (read, write, delete, list) to the backend's routing table.
	//
	// Parameters:
	//   - b: The Backend instance to register with
	//
	// Returns an error if registration fails, which will prevent the backend
	// from starting.
	Register(b *Backend) error

	// Name returns a human-readable name for this path.
	//
	// The name is used for logging and diagnostic purposes to identify
	// the path in error messages and debug output.
	Name() string

	// HelpSynopsis returns a brief one-line description of the path's purpose.
	//
	// This text is displayed in Vault's API documentation and CLI help output
	// to give users a quick overview of what the endpoint does.
	HelpSynopsis() string

	// HelpDescription returns a detailed description of the path's functionality.
	//
	// This text provides comprehensive documentation about the endpoint, including
	// what operations it supports, what parameters it accepts, and how to use it.
	// It is displayed in Vault's API documentation and CLI help output.
	HelpDescription() string

	// PeriodicFunc is called periodically by Vault to perform background maintenance tasks.
	//
	// This method is invoked at regular intervals (configured via Vault's PeriodicFunc)
	// and should be used for tasks like:
	//   - Token rotation checks
	//   - Cleanup of expired credentials
	//   - Cache updates
	//
	// Implementations should be efficient and avoid long-running operations,
	// as they may block other periodic tasks.
	//
	// Parameters:
	//   - ctx: Context for the operation, may be used for cancellation
	//   - flags: Current plugin configuration flags
	//   - path: The Path instance (typically self)
	//   - req: The Vault request context containing storage and other utilities
	//
	// Returns an error if the periodic operation fails. Errors are logged but
	// typically do not prevent future periodic invocations.
	PeriodicFunc(ctx context.Context, flags flags.Flags, path Path, req *logical.Request) error

	// InvalidateFunc is called when a key in storage is updated or deleted.
	//
	// This method provides a cache invalidation mechanism. When data changes in
	// Vault's storage layer, this function is called to allow the path to update
	// any cached state (like GitLab client connections).
	//
	// Common use cases:
	//   - Invalidating GitLab client connections when config changes
	//   - Clearing role caches when role definitions are updated
	//   - Resetting internal state when related data changes
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - flags: Current plugin configuration flags
	//   - path: The Path instance (typically self)
	//   - key: The storage key that was invalidated (e.g., "config/default")
	//
	// This method should not return an error; it should handle failures gracefully
	// since invalidation is advisory and should not break the system.
	InvalidateFunc(ctx context.Context, flags flags.Flags, path Path, key string)
}
