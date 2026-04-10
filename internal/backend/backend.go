package backend

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
)

// ClientProvider abstracts obtaining a gitlab client by config name.
type ClientProvider interface {
	GetClient(name string) gitlab.Client
	SetClient(client gitlab.Client, name string)
	GetClientByName(ctx context.Context, s logical.Storage, name string) (gitlab.Client, error)
}

// EventSender abstracts sending audit/events from the backend.
type EventSender interface {
	SendEvent(ctx context.Context, eventType string, metadata map[string]string) error
}

// Backend defines the contract that path handlers and cross-package consumers depend on.
type Backend interface {
	ClientProvider
	EventSender

	// Logger returns the backend logger.
	Logger() hclog.Logger

	// Flags returns a copy of the current flags (thread-safe read).
	Flags() flags.Flags

	// UpdateFlags atomically updates the flags under a write lock.
	UpdateFlags(fn func(*flags.Flags))

	// Client config locking
	ClientLock()
	ClientUnlock()
	ClientRLock()
	ClientRUnlock()

	// Role locking
	RoleLockForKey(key string) *locksutil.LockEntry

	// SecretForType returns the framework.Secret for the given secret type.
	SecretForType(secretType string) *framework.Secret

	// GetConfig retrieves a config entry from storage by name.
	GetConfig(ctx context.Context, s logical.Storage, name string) (*config.EntryConfig, error)

	// SaveConfig persists a config entry to storage.
	SaveConfig(ctx context.Context, cfg *config.EntryConfig, s logical.Storage) error

	// GetRole retrieves a role entry from storage by name.
	GetRole(ctx context.Context, name string, s logical.Storage) (*role.Role, error)
}
