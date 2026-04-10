package backend_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/mocks"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func TestClientStore(t *testing.T) {
	b := newTestBackend(t, flags.Flags{})
	client := mocks.NewMockClient(t)

	assert.Nil(t, b.GetClient("missing"))

	b.SetClient(client, "cfg")
	assert.Equal(t, client, b.GetClient("cfg"))

	b.SetClient(nil, "niltest")
	assert.Nil(t, b.GetClient("niltest"))

	b.SetClient(client, "")
	assert.Equal(t, client, b.GetClient("default"))

	b.DeleteClient("cfg")
	assert.Nil(t, b.GetClient("cfg"))
}

func TestGetClientByName(t *testing.T) {
	t.Run("cached valid client returned directly", func(t *testing.T) {
		b := newTestBackend(t, flags.Flags{})
		client := mocks.NewMockClient(t)
		client.EXPECT().Valid(mock.Anything).Return(true)
		b.SetClient(client, "test")

		got, err := b.GetClientByName(t.Context(), &logical.InmemStorage{}, "test")
		require.NoError(t, err)
		assert.Equal(t, client, got)
	})

	t.Run("missing config returns error", func(t *testing.T) {
		b := newTestBackend(t, flags.Flags{})
		got, err := b.GetClientByName(t.Context(), &logical.InmemStorage{}, "missing")
		assert.Nil(t, got)
		assert.Error(t, err)
	})

	t.Run("context-injected client used when config exists", func(t *testing.T) {
		b := newTestBackend(t, flags.Flags{})
		s := &logical.InmemStorage{}
		require.NoError(t, b.SaveConfig(t.Context(), &config.EntryConfig{Name: "x", BaseURL: "https://gl.io", Token: "t"}, s))

		injected := mocks.NewMockClient(t)
		got, err := b.GetClientByName(gitlab.ClientNewContext(t.Context(), injected), s, "x")
		require.NoError(t, err)
		assert.Equal(t, injected, got)
	})
}
