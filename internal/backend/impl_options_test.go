package backend_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
)

func TestWithVersion(t *testing.T) {
	b := newTestBackend(t, backend.WithVersion("1.2.3"))
	assert.Equal(t, "1.2.3", b.RunningVersion)
}

func TestWithHelp(t *testing.T) {
	b := newTestBackend(t, backend.WithHelp("  help text  "))
	assert.Equal(t, "help text", b.Help)
}

func TestWithProviders(t *testing.T) {
	p := &dummyProvider{
		name:  "test",
		paths: []*framework.Path{{Pattern: "test/path"}},
	}
	b := newTestBackend(t, backend.WithProviders(p))
	assert.NotEmpty(t, b.Paths)
}

func TestWithSecrets(t *testing.T) {
	s := &framework.Secret{Type: "test_secret"}
	b := newTestBackend(t, backend.WithSecrets(s))
	require.Len(t, b.Secrets, 1)
	assert.Equal(t, "test_secret", b.Secrets[0].Type)
}

func TestWithSealWrapStorage(t *testing.T) {
	b := newTestBackend(t, backend.WithSealWrapStorage("secret/path"))
	assert.Contains(t, b.PathsSpecial.SealWrapStorage, "secret/path")
}

func TestWithLocalStorage(t *testing.T) {
	b := newTestBackend(t, backend.WithLocalStorage("local/path"))
	assert.Contains(t, b.PathsSpecial.LocalStorage, "local/path")
	assert.Contains(t, b.PathsSpecial.LocalStorage, framework.WALPrefix)
}
