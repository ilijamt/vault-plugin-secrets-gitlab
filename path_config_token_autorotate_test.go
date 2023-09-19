package gitlab_test

import (
	"context"
	"github.com/hashicorp/vault/sdk/logical"
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPathConfig_AutoRotate(t *testing.T) {
	t.Run("auto_rotate_token should be false if not specified", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token": "super-secret-token",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.Data["auto_rotate_token"].(bool))
	})

	t.Run("auto_rotate_before cannot be more than the minimal value", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":              "super-secret-token",
				"max_ttl":            "48h",
				"auto_rotate_before": "2h",
			},
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
	})

	t.Run("auto_rotate_before should be less than the maximal limit", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":              "super-secret-token",
				"max_ttl":            "48h",
				"auto_rotate_before": (gitlab.DefaultAutoRotateBeforeMaxTTL + time.Hour).String(),
			},
		})
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, resp)
	})

	t.Run("auto_rotate_before should be set to correct value", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":              "super-secret-token",
				"max_ttl":            "48h",
				"auto_rotate_before": "48h",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.EqualValues(t, "48h0m0s", resp.Data["auto_rotate_before"])
	})

	t.Run("auto_rotate_before should be more than the minimal limit", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":              "super-secret-token",
				"max_ttl":            "48h",
				"auto_rotate_before": (gitlab.DefaultAutoRotateBeforeMinTTL - time.Hour).String(),
			},
		})
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, resp)
	})

	t.Run("auto_rotate_before should be set to min if not specified", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":   "super-secret-token",
				"max_ttl": "48h",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Data["auto_rotate_before"])
	})

	t.Run("auto_rotate_before should be between the min and max value", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
			Data: map[string]interface{}{
				"token":              "super-secret-token",
				"max_ttl":            "48h",
				"auto_rotate_before": "10h",
			},
		})
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, resp)
	})
}

func TestPathConfig_AutoRotateToken(t *testing.T) {

	t.Run("no error when auto rotate is disabled and config is not set", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)

		err = b.PeriodicFunc(context.Background(), &logical.Request{Storage: l})
		require.NoError(t, err)
	})

	t.Run("no error when auto rotate is disabled and config is set", func(t *testing.T) {
		b, l, err := getBackendWithConfig(map[string]any{"token": "token"})
		require.NoError(t, err)

		err = b.PeriodicFunc(context.Background(), &logical.Request{Storage: l})
		require.NoError(t, err)
	})

	t.Run("call auto rotate the main token and rotate the token", func(t *testing.T) {
		b, l, events, err := getBackendWithEventsAndConfig(map[string]any{"token": "token", "revoke_auto_rotated_token": true, "auto_rotate_token": true})
		require.NoError(t, err)

		var client = newInMemoryClient(true)
		client.rotateMainToken.Token = "new token"
		b.SetClient(client)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.Empty(t, resp.Data["token_expires_at"])

		err = b.PeriodicFunc(context.Background(), &logical.Request{Storage: l})
		require.NoError(t, err)
		assert.Greater(t, client.calledMainToken, 0)
		assert.Greater(t, client.calledRotateMainToken, 0)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.EqualValues(t, "new token", resp.Data["token"])
		require.NotEmpty(t, resp.Data["token_expires_at"])

		events.expectEvents(t, []expectedEvent{
			{
				eventType: "gitlab/config-write",
			},
			{
				eventType: "gitlab/config-token-rotate",
			},
			{
				eventType: "gitlab/config-token-revoke",
			},
		})

	})

	t.Run("call auto rotate the main token but the token is still valid", func(t *testing.T) {
		b, l, err := getBackendWithConfig(map[string]any{"token": "token", "auto_rotate_token": true})
		require.NoError(t, err)

		var client = newInMemoryClient(true)
		var expiresAt = time.Now().Add(100 * 24 * time.Hour)
		client.mainTokenInfo.ExpiresAt = &expiresAt
		b.SetClient(client)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.Empty(t, resp.Data["token_expires_at"])

		err = b.PeriodicFunc(context.Background(), &logical.Request{Storage: l})
		require.NoError(t, err)
		assert.Greater(t, client.calledMainToken, 0)
		assert.EqualValues(t, client.calledRotateMainToken, 0)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.NotEmpty(t, resp.Data["token_expires_at"])
	})

}
