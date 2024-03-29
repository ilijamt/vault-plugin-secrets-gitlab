package gitlab_test

import (
	"context"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathConfig(t *testing.T) {
	t.Run("initial config should be empty", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, resp.Error(), gitlab.ErrBackendNotConfigured)
	})

	t.Run("deleting uninitialized config should return error", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.True(t, resp.IsError())
		require.EqualValues(t, resp.Error(), gitlab.ErrBackendNotConfigured)
	})

	t.Run("write, read, delete and read config", func(t *testing.T) {
		b, l, events, err := getBackendWithEvents()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]any{
				"token":    "token",
				"base_url": "https://gitlab.com",
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		assert.NotEmpty(t, resp.Data["token_sha1_hash"])
		assert.NotEmpty(t, resp.Data["base_url"])
		require.Len(t, events.eventsProcessed, 1)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
		require.Len(t, events.eventsProcessed, 2)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())

		events.expectEvents(t, []expectedEvent{
			{
				eventType: "gitlab/config-write",
			},
			{
				eventType: "gitlab/config-delete",
			},
		})
	})

	t.Run("missing token from the request", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]any{},
		})

		require.Error(t, err)
		require.Nil(t, resp)

		var errorMap = countErrByName(err.(*multierror.Error))
		assert.EqualValues(t, 1, errorMap[gitlab.ErrFieldRequired.Error()])
		require.Len(t, errorMap, 1)
	})
}
