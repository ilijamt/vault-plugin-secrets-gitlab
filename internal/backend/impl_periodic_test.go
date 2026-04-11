package backend_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
)

// dummyPeriodicProvider implements PathProvider and PeriodicHandler.
type dummyPeriodicProvider struct {
	dummyProvider
	periodicErr error
	called      bool
}

func (d *dummyPeriodicProvider) PeriodicFunc(_ context.Context, _ *logical.Request) error {
	d.called = true
	return d.periodicErr
}

func TestPeriodicFunc_NoProviders(t *testing.T) {
	b := newTestBackend(t)
	err := b.PeriodicFunc(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}})
	assert.NoError(t, err)
}

func TestPeriodicFunc_WithPeriodicHandler(t *testing.T) {
	p := &dummyPeriodicProvider{dummyProvider: dummyProvider{name: "periodic"}}
	b := newTestBackend(t, backend.WithProviders(p))

	err := b.PeriodicFunc(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}})
	require.NoError(t, err)
	assert.True(t, p.called)
}

func TestPeriodicFunc_SkipsNonPeriodicProviders(t *testing.T) {
	p := &dummyProvider{name: "plain"}
	b := newTestBackend(t, backend.WithProviders(p))

	err := b.PeriodicFunc(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}})
	assert.NoError(t, err)
}

func TestPeriodicFunc_ErrorAggregation(t *testing.T) {
	p1 := &dummyPeriodicProvider{dummyProvider: dummyProvider{name: "p1"}, periodicErr: errors.New("err1")}
	p2 := &dummyPeriodicProvider{dummyProvider: dummyProvider{name: "p2"}, periodicErr: errors.New("err2")}
	b := newTestBackend(t, backend.WithProviders(p1, p2))

	err := b.PeriodicFunc(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}})
	require.Error(t, err)
	assert.ErrorContains(t, err, "err1")
	assert.ErrorContains(t, err, "err2")
	assert.True(t, p1.called)
	assert.True(t, p2.called)
}
