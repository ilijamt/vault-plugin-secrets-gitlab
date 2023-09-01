package gitlab

import (
	"context"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
	"strings"
	"sync"
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
		BackendType: logical.TypeLogical,
		Help:        strings.TrimSpace(backendHelp),
		Invalidate:  b.Invalidate,

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
				pathListRoles(b),
				pathRoles(b),
				pathTokenRoles(b),
			},
		),

		PeriodicFunc: b.periodicFunc,
	}

	b.SetClient(nil)
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
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
	b.Logger().Debug("Periodic action executed")
	return b.rotateConfigToken(ctx, request)
}

// Invalidate invalidates the key if required
func (b *Backend) Invalidate(ctx context.Context, key string) {
	b.Logger().Debug("Backend invalidate", "key", key)
	if key == PathConfigStorage {
		b.Logger().Warn("gitlab config changed, reinitializing the gitlab client")
		b.lockClientMutex.Lock()
		defer b.lockClientMutex.Unlock()
		b.client = nil
	}
}

func (b *Backend) SetClient(client Client) {
	b.client = client
}

func (b *Backend) getClient(ctx context.Context, s logical.Storage) (Client, error) {
	if b.client != nil && b.client.Valid() {
		return b.client, nil
	}

	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()
	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	client, err := NewGitlabClient(config)
	if err == nil {
		b.SetClient(client)
	}
	return client, err
}
