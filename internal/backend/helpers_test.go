package backend_test

import (
	"context"
	"io"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestBackend(t *testing.T, f flags.Flags, opts ...backend.InitOption) *backend.BackendImpl {
	t.Helper()
	b := backend.New(f)
	ctx := t.Context()
	err := b.Init(ctx, &logical.BackendConfig{
		Logger:      hclog.New(&hclog.LoggerOptions{Output: io.Discard, Level: hclog.NoLevel}),
		System:      &logical.StaticSystemView{},
		StorageView: &logical.InmemStorage{},
	}, opts...)
	require.NoError(t, err)
	return b
}

func newPeriodicProvider(t *testing.T, name string, err error) *periodicPathProvider {
	t.Helper()
	pp := mocks.NewMockPathProvider(t)
	pp.EXPECT().Paths().Return(nil)
	pp.EXPECT().Name().Return(name)
	ph := mocks.NewMockPeriodicHandler(t)
	ph.EXPECT().PeriodicFunc(mock.Anything, mock.Anything).Return(err)
	return &periodicPathProvider{pp, ph}
}

type periodicPathProvider struct {
	*mocks.MockPathProvider
	*mocks.MockPeriodicHandler
}

func (p *periodicPathProvider) Name() string             { return p.MockPathProvider.Name() }
func (p *periodicPathProvider) Paths() []*framework.Path { return p.MockPathProvider.Paths() }
func (p *periodicPathProvider) PeriodicFunc(ctx context.Context, req *logical.Request) error {
	return p.MockPeriodicHandler.PeriodicFunc(ctx, req)
}

type invalidatePathProvider struct {
	*mocks.MockPathProvider
	*mocks.MockInvalidateHandler
}

func (p *invalidatePathProvider) Name() string             { return p.MockPathProvider.Name() }
func (p *invalidatePathProvider) Paths() []*framework.Path { return p.MockPathProvider.Paths() }
func (p *invalidatePathProvider) Invalidate(ctx context.Context, key string) {
	p.MockInvalidateHandler.Invalidate(ctx, key)
}
