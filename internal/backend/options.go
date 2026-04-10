package backend

import "github.com/hashicorp/vault/sdk/framework"

// initConfig holds the configuration assembled from InitOption functions.
type initConfig struct {
	version         string
	help            string
	providers       []PathProvider
	secrets         []*framework.Secret
	sealWrapStorage []string
	localStorage    []string
}

// InitOption is a functional option for configuring backend initialization.
type InitOption func(*initConfig)

// WithVersion sets the running version of the backend.
func WithVersion(v string) InitOption {
	return func(c *initConfig) { c.version = v }
}

// WithHelp sets the help text for the backend.
func WithHelp(h string) InitOption {
	return func(c *initConfig) { c.help = h }
}

// WithProviders registers path providers with the backend.
func WithProviders(p ...PathProvider) InitOption {
	return func(c *initConfig) { c.providers = append(c.providers, p...) }
}

// WithSecrets registers framework secrets with the backend.
func WithSecrets(s ...*framework.Secret) InitOption {
	return func(c *initConfig) { c.secrets = append(c.secrets, s...) }
}

// WithSealWrapStorage specifies storage paths that should be seal-wrapped.
func WithSealWrapStorage(paths ...string) InitOption {
	return func(c *initConfig) { c.sealWrapStorage = append(c.sealWrapStorage, paths...) }
}

// WithLocalStorage specifies storage paths that should be stored locally.
func WithLocalStorage(paths ...string) InitOption {
	return func(c *initConfig) { c.localStorage = append(c.localStorage, paths...) }
}
