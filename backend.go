package gitlab

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

const (
	// operationPrefixGitlabAccessTokens is used as expected prefix for OpenAPI operation id's.
	operationPrefixGitlabAccessTokens = "gitlab"

	backendHelp = `
The Gitlab Access token auth Backend dynamically generates private 
and group tokens.

After mounting this Backend, credentials to manage Gitlab tokens must be configured 
with the "^config/(?P<config_name>\w(([\w-.@]+)?\w)?)$" endpoints.
`
)

func Factory(flags flags.Flags) logical.Factory {
	return func(ctx context.Context, config *logical.BackendConfig) (logical.Backend, error) {
		return factory(ctx, config, flags)
	}
}

// Factory returns expected new Backend as logical.Backend
func factory(ctx context.Context, conf *logical.BackendConfig, flags flags.Flags) (logical.Backend, error) {
	var b = &Backend{
		roleLocks: locksutil.CreateLocks(),
		clients:   sync.Map{},
		flags:     flags,
	}

	b.Backend = &framework.Backend{
		BackendType:    logical.TypeLogical,
		Help:           strings.TrimSpace(backendHelp),
		RunningVersion: Version,
		Invalidate:     b.Invalidate,

		PathsSpecial: &logical.Paths{
			LocalStorage: []string{
				framework.WALPrefix,
			},
			SealWrapStorage: []string{
				PathConfigStorage,
			},
		},

		Secrets: []*framework.Secret{
			secretAccessTokens(b),
		},

		Paths: framework.PathAppend(
			[]*framework.Path{
				pathFlags(b),
				pathConfig(b),
				pathListConfig(b),
				pathConfigTokenRotate(b),
				pathListRoles(b),
				pathRoles(b),
				pathTokenRoles(b),
			},
		),

		PeriodicFunc: b.periodicFunc,
	}

	var err = b.Setup(ctx, conf)
	return b, err
}

type Backend struct {
	*framework.Backend

	flags flags.Flags

	// The client that we can use to create and revoke the access tokens
	clients sync.Map

	// Mutex to protect access to gitlab clients and client configs, a change to the gitlab client config
	// would invalidate the gitlab client, so it will need to be reinitialized
	lockClientMutex sync.RWMutex

	// Mutex to protect flags change, this is required as it could require reinitialization of some components
	// of the plugin
	lockFlagsMutex sync.RWMutex

	// roleLocks to protect access for roles, during modifications, deletion
	roleLocks []*locksutil.LockEntry
}

func (b *Backend) periodicFunc(ctx context.Context, req *logical.Request) (err error) {
	b.Logger().Debug("Periodic action executing")

	if b.WriteSafeReplicationState() {
		var config *EntryConfig

		b.lockClientMutex.Lock()
		unlockLockClientMutex := sync.OnceFunc(func() { b.lockClientMutex.Unlock() })
		defer unlockLockClientMutex()

		// @TODO: Check and fix this is not correct, the locking mechanism doesn't make sense

		var configs []string
		configs, err = req.Storage.List(ctx, fmt.Sprintf("%s/", PathConfigStorage))

		for _, name := range configs {
			if config, err = getConfig(ctx, req.Storage, name); err == nil {
				b.Logger().Debug("Trying to rotate the config", "name", name)
				unlockLockClientMutex()
				if config != nil {
					// If we need to autorotate the token, initiate the procedure to autorotate the token
					if config.AutoRotateToken {
						err = errors.Join(err, b.checkAndRotateConfigToken(ctx, req, config))
					}
				}
			}
		}
	}

	return err
}

// Invalidate invalidates the key if required
func (b *Backend) Invalidate(ctx context.Context, key string) {
	b.Logger().Debug("Backend invalidate", "key", key)
	if strings.HasPrefix(key, PathConfigStorage) {
		parts := strings.SplitN(key, "/", 2)
		var name = parts[1]
		b.Logger().Warn(fmt.Sprintf("Gitlab config for %s changed, reinitializing the gitlab client", name))
		b.lockClientMutex.Lock()
		defer b.lockClientMutex.Unlock()
		b.clients.Delete(name)
	}
}

func (b *Backend) GetClient(name string) Client {
	if client, ok := b.clients.Load(cmp.Or(name, DefaultConfigName)); ok {
		return client.(Client)
	}
	return nil
}

func (b *Backend) SetClient(client Client, name string) {
	name = cmp.Or(name, DefaultConfigName)
	if client == nil {
		b.Logger().Debug("Setting a nil client")
		return
	}
	b.Logger().Debug("Setting a new client")
	b.clients.Store(name, client)
}

func (b *Backend) getClient(ctx context.Context, s logical.Storage, name string) (client Client, err error) {
	if c, ok := b.clients.Load(cmp.Or(name, DefaultConfigName)); ok {
		client = c.(Client)
	}
	if client != nil && client.Valid(ctx) {
		b.Logger().Debug("Returning existing gitlab client")
		return client, nil
	}

	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()
	var config *EntryConfig
	config, err = getConfig(ctx, s, name)
	if err != nil {
		b.Logger().Error("Failed to retrieve configuration", "error", err.Error())
		return nil, err
	}

	var httpClient *http.Client
	httpClient, _ = utils.HttpClientFromContext(ctx)
	if client, _ = ClientFromContext(ctx); client == nil {
		if client, err = NewGitlabClient(config, httpClient, b.Logger()); err == nil {
			b.SetClient(client, name)
		}
	}
	return client, err
}
