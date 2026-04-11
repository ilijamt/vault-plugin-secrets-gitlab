package gitlab

import (
	"context"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	configPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
	flagsPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/flags"
	rolePaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/role"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
)

// Backend is the public type alias for the concrete backend implementation.
// Tests and consumers use *Backend which is equivalent to *backend.BackendImpl.
type Backend = backend.Impl

const (
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

// factory creates and initializes the Backend with all path providers.
func factory(ctx context.Context, conf *logical.BackendConfig, f flags.Flags) (logical.Backend, error) {
	b := backend.New(f)

	s := secret.NewSecret(b, backend.DefaultConfigName)

	err := b.Init(ctx, conf,
		backend.WithVersion(Version),
		backend.WithHelp(backendHelp),
		backend.WithProviders(
			flagsPaths.New(b),
			configPaths.New(b),
			rolePaths.New(b),
			tokenPaths.New(b, s),
		),
		backend.WithSecrets(s),
		backend.WithSealWrapStorage(backend.PathConfigStorage),
		backend.WithLocalStorage(),
	)

	return b, err
}
