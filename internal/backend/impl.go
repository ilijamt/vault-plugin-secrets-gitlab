package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	g "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

var (
	_ Logging                   = (*Impl)(nil)
	_ FlagsProvider             = (*Impl)(nil)
	_ ClientReader              = (*Impl)(nil)
	_ ClientGetter              = (*Impl)(nil)
	_ ClientSetter              = (*Impl)(nil)
	_ ClientDeleter             = (*Impl)(nil)
	_ ClientLocker              = (*Impl)(nil)
	_ RoleLocker                = (*Impl)(nil)
	_ ConfigStore               = (*Impl)(nil)
	_ RoleStore                 = (*Impl)(nil)
	_ EventSender               = (*Impl)(nil)
	_ WriteSafeReplicationState = (*Impl)(nil)
	_ Backend                   = (*Impl)(nil)
)

// Impl is the concrete implementation of the Backend interface.
type Impl struct {
	*framework.Backend

	flags          flags.Flags
	lockFlagsMutex sync.RWMutex

	// The client that we can use to create and revoke the access tokens
	clients sync.Map

	// Mutex to protect access to gitlab clients, a change to the gitlab client config
	// would invalidate the gitlab client, so it will need to be reinitialized
	// a change to the config should delete the client from the map, and the next request will
	// create a new client with the new config
	lockClientMutex sync.Mutex

	// roleLocks to protect access for roles, during modifications, deletion
	roleLocks []*locksutil.LockEntry

	// pathProviders holds the registered path providers with their optional hooks
	pathProviders []PathProvider
}

// New creates a new BackendImpl with the given flags. Call Init to complete setup.
func New(f flags.Flags) *Impl {
	return &Impl{
		roleLocks: locksutil.CreateLocks(),
		clients:   sync.Map{},
		flags:     f,
	}
}

// Init wires up the framework.Backend with paths from the registered providers,
// secrets, special paths, and periodic/invalidate dispatchers.
func (b *Impl) Init(ctx context.Context, conf *logical.BackendConfig, opts ...InitOption) error {
	cfg := &initConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	b.pathProviders = cfg.providers

	var allPaths []*framework.Path
	for _, p := range cfg.providers {
		allPaths = append(allPaths, p.Paths()...)
	}

	b.Backend = &framework.Backend{
		BackendType:    logical.TypeLogical,
		Help:           strings.TrimSpace(cfg.help),
		RunningVersion: cfg.version,
		Invalidate:     b.invalidate,
		PathsSpecial: &logical.Paths{
			LocalStorage:    append([]string{framework.WALPrefix}, cfg.localStorage...),
			SealWrapStorage: cfg.sealWrapStorage,
		},
		Secrets:      cfg.secrets,
		Paths:        framework.PathAppend(allPaths),
		PeriodicFunc: b.periodicFunc,
	}

	return b.Setup(ctx, conf)
}

// periodicFunc dispatches to all registered PeriodicHandlers.
// Only called when WriteSafeReplicationState() is true.
func (b *Impl) periodicFunc(ctx context.Context, req *logical.Request) error {
	b.Logger().Debug("Periodic action executing")
	if !b.WriteSafeReplicationState() {
		return nil
	}
	var errs error
	for _, p := range b.pathProviders {
		if ph, ok := p.(PeriodicHandler); ok {
			b.Logger().Debug("Periodic handler dispatching", "provider", p.Name())
			errs = errors.Join(errs, ph.PeriodicFunc(ctx, req))
		}
	}
	return errs
}

// invalidate dispatches to all registered InvalidateHandlers.
func (b *Impl) invalidate(ctx context.Context, key string) {
	b.Logger().Debug("Backend invalidate", "key", key)
	for _, p := range b.pathProviders {
		if ih, ok := p.(InvalidateHandler); ok {
			b.Logger().Debug("Invalidate handler dispatching", "provider", p.Name(), "key", key)
			ih.Invalidate(ctx, key)
		}
	}
}

func (b *Impl) GetClient(name string) g.Client {
	if client, ok := b.clients.Load(configName(name)); ok {
		return client.(g.Client)
	}
	return nil
}

func (b *Impl) SetClient(client g.Client, name string) {
	name = configName(name)
	if client == nil {
		b.Logger().Debug("Setting a nil client", "name", name)
		b.DeleteClient(name)
		return
	}
	b.Logger().Debug("Setting a new client", "name", name)
	b.clients.Store(name, client)
}

func (b *Impl) DeleteClient(name string) {
	name = configName(name)
	b.Logger().Debug("Removing client", "name", name)
	b.clients.Delete(name)
}

func (b *Impl) GetClientByName(ctx context.Context, s logical.Storage, name string) (client g.Client, err error) {
	name = configName(name)

	if client = b.GetClient(name); client != nil && client.Valid(ctx) {
		b.Logger().Debug("Returning existing gitlab client")
		return client, nil
	}

	b.ClientLock()
	defer b.ClientUnlock()

	// Re-check after lock acquisition — another goroutine may have created the client while we waited.
	if client = b.GetClient(name); client != nil && client.Valid(ctx) {
		return client, nil
	}

	var config *modelConfig.EntryConfig
	config, err = b.GetConfig(ctx, s, name)
	if err != nil {
		b.Logger().Error("Failed to retrieve configuration", "error", err.Error())
		return nil, err
	}

	if config == nil {
		return nil, fmt.Errorf("configuration %q not found", name)
	}

	var httpClient *http.Client
	httpClient, _ = utils.HttpClientFromContext(ctx)
	if client, _ = g.ClientFromContext(ctx); client == nil {
		if client, err = g.NewGitlabClient(config, httpClient, b.Logger()); err == nil {
			b.SetClient(client, name)
		}
	}
	return client, err
}

func (b *Impl) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	return event.Event(ctx, b.Backend, eventType, metadata)
}

func (b *Impl) Flags() flags.Flags {
	b.lockFlagsMutex.RLock()
	defer b.lockFlagsMutex.RUnlock()
	return b.flags
}

func (b *Impl) UpdateFlags(fn func(*flags.Flags)) {
	b.lockFlagsMutex.Lock()
	defer b.lockFlagsMutex.Unlock()
	fn(&b.flags)
}

func (b *Impl) ClientLock() {
	b.lockClientMutex.Lock()
}

func (b *Impl) ClientUnlock() {
	b.lockClientMutex.Unlock()
}

func (b *Impl) RoleLockForKey(key string) *locksutil.LockEntry {
	return locksutil.LockForKey(b.roleLocks, key)
}

func (b *Impl) GetConfig(ctx context.Context, s logical.Storage, name string) (*modelConfig.EntryConfig, error) {
	return model.Get[modelConfig.EntryConfig](ctx, s, fmt.Sprintf("%s/%s", PathConfigStorage, name))
}

func (b *Impl) SaveConfig(ctx context.Context, s logical.Storage, config *modelConfig.EntryConfig) error {
	return model.Save(ctx, s, PathConfigStorage, config)
}

func (b *Impl) GetRole(ctx context.Context, s logical.Storage, name string) (*role.Role, error) {
	return model.Get[role.Role](ctx, s, fmt.Sprintf("%s/%s", PathRoleStorage, name))
}
