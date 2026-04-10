package backend_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/mocks"
)

func TestPeriodicFunc(t *testing.T) {
	t.Run("no providers", func(t *testing.T) {
		b := newTestBackend(t, flags.Flags{})
		assert.NoError(t, b.PeriodicFunc(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}))
	})

	t.Run("dispatches and joins errors", func(t *testing.T) {
		b := newTestBackend(t, flags.Flags{}, backend.WithProviders(
			newPeriodicProvider(t, "p1", errors.New("err1")),
			newPeriodicProvider(t, "p2", errors.New("err2")),
		))
		err := b.PeriodicFunc(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "err1")
		assert.Contains(t, err.Error(), "err2")
	})
}

func TestInvalidate(t *testing.T) {
	pp := mocks.NewMockPathProvider(t)
	pp.EXPECT().Paths().Return(nil)
	pp.EXPECT().Name().Return("inv")

	ih := mocks.NewMockInvalidateHandler(t)
	ih.EXPECT().Invalidate(mock.Anything, "config/default")

	b := newTestBackend(t, flags.Flags{}, backend.WithProviders(&invalidatePathProvider{pp, ih}))
	b.Invalidate(t.Context(), "config/default")
}
