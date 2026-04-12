package backend_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func TestNew(t *testing.T) {
	b := backend.New(flags.Flags{ShowConfigToken: true})
	assert.True(t, b.Flags().ShowConfigToken)
}

func TestFlags(t *testing.T) {
	b := backend.New(flags.Flags{})
	assert.False(t, b.Flags().ShowConfigToken)

	b.UpdateFlags(func(f *flags.Flags) { f.ShowConfigToken = true })
	assert.True(t, b.Flags().ShowConfigToken)
}

func TestClientCRUD(t *testing.T) {
	b := newTestBackend(t)
	c := &dummyClient{valid: true}

	assert.Nil(t, b.GetClient("x"))

	b.SetClient(c, "x")
	assert.Same(t, c, b.GetClient("x"))

	b.DeleteClient("x")
	assert.Nil(t, b.GetClient("x"))

	b.SetClient(c, "y")
	b.SetClient(nil, "y")
	assert.Nil(t, b.GetClient("y"))
}

func TestClientDefaultName(t *testing.T) {
	b := newTestBackend(t)
	c := &dummyClient{valid: true}
	b.SetClient(c, "")
	assert.Same(t, c, b.GetClient(backend.DefaultConfigName))
}

func TestLockForKey(t *testing.T) {
	b := backend.New(flags.Flags{})

	l := b.LockForKey("role", "k")
	require.NotNil(t, l)
	assert.Same(t, l, b.LockForKey("role", "k"))

	// Different paths with the same key produce different locks.
	assert.NotSame(t, b.LockForKey("role", "x"), b.LockForKey("config", "x"))
}

func TestGetConfig(t *testing.T) {
	b := backend.New(flags.Flags{})
	ctx, s := t.Context(), &logical.InmemStorage{}

	cfg, err := b.GetConfig(ctx, s, "missing")
	require.NoError(t, err)
	assert.Nil(t, cfg)

	require.NoError(t, b.SaveConfig(ctx, s, &config.EntryConfig{Name: "a", BaseURL: "https://gitlab.com"}))
	cfg, err = b.GetConfig(ctx, s, "a")
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.com", cfg.BaseURL)
}

func TestGetRole(t *testing.T) {
	b := backend.New(flags.Flags{})
	r, err := b.GetRole(t.Context(), &logical.InmemStorage{}, "missing")
	require.NoError(t, err)
	assert.Nil(t, r)
}

func TestGetClientByName(t *testing.T) {
	b := newTestBackend(t)
	ctx, s := t.Context(), &logical.InmemStorage{}

	// cached valid client returned directly
	c := &dummyClient{valid: true}
	b.SetClient(c, "ok")
	got, err := b.GetClientByName(ctx, s, "ok")
	require.NoError(t, err)
	assert.Same(t, c, got)

	// stale client triggers config lookup — no config means error
	b.SetClient(&dummyClient{valid: false}, "stale")
	_, err = b.GetClientByName(ctx, s, "stale")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not found")

	// client injected via context is used
	ctxC := &dummyClient{valid: true}
	require.NoError(t, b.SaveConfig(t.Context(), s, &config.EntryConfig{Name: "ctx-cfg"}))
	got, err = b.GetClientByName(gitlab.ClientNewContext(ctx, ctxC), s, "ctx-cfg")
	require.NoError(t, err)
	assert.Same(t, ctxC, got)
}

func TestGetClientByName_GetConfigError(t *testing.T) {
	b := newTestBackend(t)
	b.SetClient(&dummyClient{valid: false}, "fail")

	_, err := b.GetClientByName(t.Context(), newErrorStorage(), "fail")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "storage error")
}

func TestGetClientByName_NewGitlabClientError(t *testing.T) {
	b := newTestBackend(t)
	s := &logical.InmemStorage{}

	require.NoError(t, b.SaveConfig(t.Context(), s, &config.EntryConfig{Name: "bad", BaseURL: "", Token: ""}))

	_, err := b.GetClientByName(t.Context(), s, "bad")
	assert.Error(t, err)
}

func TestGetClientByName_NewGitlabClientSuccess(t *testing.T) {
	b := newTestBackend(t)
	s := &logical.InmemStorage{}

	require.NoError(t, b.SaveConfig(t.Context(), s, &config.EntryConfig{Name: "good", BaseURL: "https://gitlab.com", Token: "glpat-test"}))

	got, err := b.GetClientByName(t.Context(), s, "good")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Same(t, got, b.GetClient("good"))
}

func TestGetClientByName_ConcurrentAccess(t *testing.T) {
	b := newTestBackend(t)
	s := &logical.InmemStorage{}

	require.NoError(t, b.SaveConfig(t.Context(), s, &config.EntryConfig{Name: "race", BaseURL: "https://gitlab.com", Token: "glpat-test"}))

	const goroutines = 10
	errs := make(chan error, goroutines)

	for range goroutines {
		go func() {
			_, err := b.GetClientByName(t.Context(), s, "race")
			errs <- err
		}()
	}

	for range goroutines {
		require.NoError(t, <-errs)
	}
	assert.NotNil(t, b.GetClient("race"))
}
