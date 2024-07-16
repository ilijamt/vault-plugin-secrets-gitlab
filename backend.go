package gitlab

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	// operationPrefixGitlabAccessTokens is used as expected prefix for OpenAPI operation id's.
	operationPrefixGitlabAccessTokens = "gitlab"

	backendHelp = `
The Gitlab Access token auth Backend dynamically generates private 
and group tokens.

After mounting this Backend, credentials to manage Gitlab tokens must be configured 
with the "config/" endpoints.
`
)

// Factory returns expected new Backend as logical.Backend
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	var b = &Backend{
		roleLocks: locksutil.CreateLocks(),
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
				pathConfig(b),
				pathConfigTokenRotate(b),
				pathListRoles(b),
				pathRoles(b),
				pathTokenRoles(b),
			},
		),

		PeriodicFunc: b.periodicFunc,
	}

	b.SetClient(nil)
	var err = b.Setup(ctx, conf)
	return b, err
}

type Backend struct {
	*framework.Backend

	// The client that we can use to create and revoke the access tokens
	client Client

	// Mutex to protect access to gitlab clients and client configs, a change to the gitlab client config
	// would invalidate the gitlab client, so it will need to be reinitialized
	lockClientMutex sync.RWMutex

	// roleLocks to protect access for roles, during modifications, deletion
	roleLocks []*locksutil.LockEntry
}

func (b *Backend) periodicFunc(ctx context.Context, request *logical.Request) error {
	b.Logger().Debug("Periodic action executing")

	if !b.WriteSafeReplicationState() {
		return nil
	}

	var config *EntryConfig
	var err error

	b.lockClientMutex.Lock()
	if config, err = getConfig(ctx, request.Storage); err != nil {
		b.lockClientMutex.Unlock()
		return err
	}
	b.lockClientMutex.Unlock()

	if config == nil {
		return nil
	}

	// If we need to autorotate the token, initiate the procedure to autorotate the token
	if config.AutoRotateToken {
		err = errors.Join(err, b.checkAndRotateConfigToken(ctx, request, config))
	}

	return err
}

// Invalidate invalidates the key if required
func (b *Backend) Invalidate(ctx context.Context, key string) {
	b.Logger().Debug("Backend invalidate", "key", key)
	if key == PathConfigStorage {
		b.Logger().Warn("Gitlab config changed, reinitializing the gitlab client")
		b.lockClientMutex.Lock()
		defer b.lockClientMutex.Unlock()
		b.client = nil
	}
}

func (b *Backend) SetClient(client Client) {
	if client == nil {
		b.Logger().Debug("Setting a nil client")
		return
	}
	b.Logger().Debug("Setting a new client")
	b.client = client
}

func (b *Backend) getClient(ctx context.Context, s logical.Storage) (client Client, err error) {
	if b.client != nil && b.client.Valid() {
		b.Logger().Debug("Returning existing gitlab client")
		return b.client, nil
	}

	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()
	var config *EntryConfig
	config, err = getConfig(ctx, s)
	if err != nil {
		b.Logger().Error("Failed to retrieve configuration", "error", err.Error())
		return nil, err
	}

	var httpClient *http.Client
	httpClient, _ = HttpClientFromContext(ctx)
	if client, _ = GitlabClientFromContext(ctx); client == nil {
		if client, err = NewGitlabClient(config, httpClient, b.Logger()); err == nil {
			b.SetClient(client)
		}
	}
	return client, err
}
