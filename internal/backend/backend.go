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

// ClientReader provides read-only client access by config name.
type ClientReader interface {
	GetClientByName(ctx context.Context, s logical.Storage, name string) (gitlab.Client, error)
}

// ClientGetter provides direct cache access to a stored client.
type ClientGetter interface {
	GetClient(name string) gitlab.Client
}

// ClientSetter stores a client in the cache.
type ClientSetter interface {
	SetClient(client gitlab.Client, name string)
}

// ClientDeleter removes a client from the cache.
type ClientDeleter interface {
	DeleteClient(name string)
}

// Locker provides per-key locking scoped by a path prefix.
type Locker interface {
	LockForKey(path, key string) *locksutil.LockEntry
}

// ConfigStore provides config CRUD operations.
type ConfigStore interface {
	GetConfig(ctx context.Context, s logical.Storage, name string) (*config.EntryConfig, error)
	SaveConfig(ctx context.Context, s logical.Storage, cfg *config.EntryConfig) error
}

// RoleStore provides role read operations.
type RoleStore interface {
	GetRole(ctx context.Context, s logical.Storage, name string) (*role.Role, error)
}

// EventSender abstracts sending audit/events from the backend.
type EventSender interface {
	SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error
}

type WriteSafeReplicationState interface {
	WriteSafeReplicationState() bool
}

type Backend interface {
	Logging
	FlagsProvider
	ClientReader
	ClientGetter
	ClientSetter
	ClientDeleter
	Locker
	ConfigStore
	RoleStore
	EventSender
	WriteSafeReplicationState
}
