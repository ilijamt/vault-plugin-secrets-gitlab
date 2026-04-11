package backend_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
)

// dummyInvalidateProvider implements PathProvider and InvalidateHandler.
type dummyInvalidateProvider struct {
	dummyProvider
	invalidatedKey string
}

func (d *dummyInvalidateProvider) Invalidate(_ context.Context, key string) {
	d.invalidatedKey = key
}

func TestInvalidate_NoProviders(t *testing.T) {
	b := newTestBackend(t)
	b.Invalidate(t.Context(), "some/key")
}

func TestInvalidate_WithInvalidateHandler(t *testing.T) {
	p := &dummyInvalidateProvider{dummyProvider: dummyProvider{name: "inv"}}
	b := newTestBackend(t, backend.WithProviders(p))

	b.Invalidate(t.Context(), "config/default")
	assert.Equal(t, "config/default", p.invalidatedKey)
}

func TestInvalidate_SkipsNonInvalidateProviders(t *testing.T) {
	p := &dummyProvider{name: "plain"}
	b := newTestBackend(t, backend.WithProviders(p))

	b.Invalidate(t.Context(), "some/key")
}
