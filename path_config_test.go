package gitlab_test

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/logical"
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

	t.Run("write, read, delete and read config", func(t *testing.T) {
		b, l, events, err := getBackendWithEvents()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":   "token",
				"max_ttl": int((32 * time.Hour).Seconds()),
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
		assert.EqualValues(t, "token", resp.Data["token"])
		assert.NotEmpty(t, resp.Data["base_url"])
		assert.EqualValues(t, int((32 * time.Hour).Seconds()), resp.Data["max_ttl"])
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
			Data: map[string]interface{}{},
		})

		require.Error(t, err)
		require.Nil(t, resp)

		var errorMap = countErrByName(err.(*multierror.Error))
		assert.EqualValues(t, 1, errorMap[gitlab.ErrFieldRequired.Error()])
		require.Len(t, errorMap, 1)
	})

	t.Run("if max_ttl is less than 24h, should be set to 24h", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":   "token",
				"max_ttl": time.Hour * 23,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)
		require.NoError(t, resp.Error())

		assert.EqualValues(t, (time.Hour * 24).Seconds(), resp.Data["max_ttl"])
	})

	t.Run("if max_ttl is 0 or less than 0 set max_ttl to 8670 hours", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":   "token",
				"max_ttl": 0,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)
		require.NoError(t, resp.Error())

		assert.EqualValues(t, (365 * 24 * time.Hour).Seconds(), resp.Data["max_ttl"])
	})

	t.Run("if max_ttl is 0 or less than 0 set max_ttl to 8670 hours", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":   "token",
				"max_ttl": 0,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)
		require.NoError(t, resp.Error())

		assert.EqualValues(t, (365 * 24 * time.Hour).Seconds(), resp.Data["max_ttl"])
	})

	t.Run("if max_ttl is more than 8670 hours set max_ttl to 8670 hours", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":   "token",
				"max_ttl": 366 * 24 * time.Hour,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)
		require.NoError(t, resp.Error())

		assert.EqualValues(t, (365 * 24 * time.Hour).Seconds(), resp.Data["max_ttl"])
	})

	t.Run("if max_ttl is between 24h and 8670h", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":   "token",
				"max_ttl": 14 * 24 * time.Hour,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		assert.EqualValues(t, (14 * 24 * time.Hour).Seconds(), resp.Data["max_ttl"])
	})

}
