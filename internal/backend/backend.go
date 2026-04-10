package backend

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
)

// Logging provides access to the backend logger.
type Logging interface {
	Logger() hclog.Logger
}

// FlagsProvider provides read and update access to runtime flags.
type FlagsProvider interface {
	Flags() flags.Flags
	UpdateFlags(fn func(*flags.Flags))
}

// ClientProvider abstracts obtaining a gitlab client by config name.
type ClientProvider interface {
	GetClient(name string) gitlab.Client
	SetClient(client gitlab.Client, name string)
	DeleteClient(name string)
	GetClientByName(ctx context.Context, s logical.Storage, name string) (gitlab.Client, error)
}

// ClientLocker provides client-level read/write locking.
type ClientLocker interface {
	ClientLock()
	ClientUnlock()
	ClientRLock()
	ClientRUnlock()
}

// RoleLocker provides per-role key locking.
type RoleLocker interface {
	RoleLockForKey(key string) *locksutil.LockEntry
}

// ConfigStore provides config CRUD operations.
type ConfigStore interface {
	GetConfig(ctx context.Context, s logical.Storage, name string) (*config.EntryConfig, error)
	SaveConfig(ctx context.Context, cfg *config.EntryConfig, s logical.Storage) error
}

// RoleStore provides role read operations.
type RoleStore interface {
	GetRole(ctx context.Context, name string, s logical.Storage) (*role.Role, error)
}

// EventSender abstracts sending audit/events from the backend.
type EventSender interface {
	SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error
}

// Backend defines the full contract that composes all sub-interfaces.
type Backend interface {
	Logging
	FlagsProvider
	ClientProvider
	ClientLocker
	RoleLocker
	ConfigStore
	RoleStore
	EventSender
}
