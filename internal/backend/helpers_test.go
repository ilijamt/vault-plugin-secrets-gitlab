package backend_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

type dummyClient struct {
	gitlab.Client
	valid bool
}

func (d *dummyClient) Valid(_ context.Context) bool { return d.valid }

func newTestBackend(t *testing.T, opts ...backend.InitOption) *backend.Impl {
	t.Helper()
	b := backend.New(flags.Flags{})
	require.NoError(t, b.Init(t.Context(), &logical.BackendConfig{System: &logical.StaticSystemView{}}, opts...))
	return b
}

// dummyProvider implements PathProvider only.
type dummyProvider struct {
	name  string
	paths []*framework.Path
}

func (d *dummyProvider) Name() string             { return d.name }
func (d *dummyProvider) Paths() []*framework.Path { return d.paths }

// errorStorage is a logical.Storage where Get always returns an error.
type errorStorage struct {
	logical.InmemStorage
	err error
}

func newErrorStorage() *errorStorage {
	return &errorStorage{err: errors.New("storage error")}
}

func (e *errorStorage) Get(_ context.Context, _ string) (*logical.StorageEntry, error) {
	return nil, e.err
}
