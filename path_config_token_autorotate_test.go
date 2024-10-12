package gitlab_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathConfig_AutoRotate(t *testing.T) {
	t.Run("auto_rotate_token should be false if not specified", func(t *testing.T) {
		ctx, url := getCtxGitlabClientWithUrl(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":    "glpat-secret-random-token",
				"base_url": url,
				"type":     gitlab.TypeSelfManaged.String(),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.Data["auto_rotate_token"].(bool))
	})

	t.Run("auto_rotate_before cannot be more than the minimal value", func(t *testing.T) {
		ctx, url := getCtxGitlabClientWithUrl(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":              "glpat-secret-random-token",
				"base_url":           url,
				"auto_rotate_before": "2h",
				"type":               gitlab.TypeSelfManaged.String(),
			},
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
	})

	t.Run("auto_rotate_before should be less than the maximal limit", func(t *testing.T) {
		ctx, url := getCtxGitlabClientWithUrl(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":              "glpat-secret-random-token",
				"base_url":           url,
				"auto_rotate_before": (gitlab.DefaultAutoRotateBeforeMaxTTL + time.Hour).String(),
				"type":               gitlab.TypeSelfManaged.String(),
			},
		})
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, resp)
	})

	t.Run("auto_rotate_before should be set to correct value", func(t *testing.T) {
		ctx, url := getCtxGitlabClientWithUrl(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":              "glpat-secret-random-token",
				"base_url":           url,
				"auto_rotate_before": "48h",
				"type":               gitlab.TypeSelfManaged.String(),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.EqualValues(t, "48h0m0s", resp.Data["auto_rotate_before"])
	})

	t.Run("auto_rotate_before should be more than the minimal limit", func(t *testing.T) {
		ctx, url := getCtxGitlabClientWithUrl(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":              "glpat-secret-random-token",
				"base_url":           url,
				"auto_rotate_before": (gitlab.DefaultAutoRotateBeforeMinTTL - time.Hour).String(),
				"type":               gitlab.TypeSelfManaged.String(),
			},
		})
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, resp)
	})

	t.Run("auto_rotate_before should be set to min if not specified", func(t *testing.T) {
		ctx, url := getCtxGitlabClientWithUrl(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":    "glpat-secret-random-token",
				"base_url": url,
				"type":     gitlab.TypeSelfManaged.String(),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Data["auto_rotate_before"])
	})

	t.Run("auto_rotate_before should be between the min and max value", func(t *testing.T) {
		ctx, url := getCtxGitlabClientWithUrl(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":              "glpat-secret-random-token",
				"base_url":           url,
				"auto_rotate_before": "10h",
				"type":               gitlab.TypeSelfManaged.String(),
			},
		})
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
		require.Nil(t, resp)
	})
}

func TestPathConfig_AutoRotateToken(t *testing.T) {
	t.Run("no error when auto rotate is disabled and config is not set", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)

		err = b.PeriodicFunc(ctx, &logical.Request{Storage: l})
		require.NoError(t, err)
	})

	t.Run("no error when auto rotate is disabled and config is set", func(t *testing.T) {
		var client = newInMemoryClient(true)
		ctx, url := getCtxGitlabClientWithUrl(t)
		ctx = gitlab.GitlabClientNewContext(ctx, client)
		b, l, err := getBackendWithConfig(ctx, map[string]any{
			"token":    "glpat-secret-token",
			"base_url": url,
			"type":     gitlab.TypeSelfManaged.String(),
		})
		require.NoError(t, err)

		b.SetClient(newInMemoryClient(true), gitlab.DefaultConfigName)
		err = b.PeriodicFunc(ctx, &logical.Request{Storage: l})
		require.NoError(t, err)
	})

	t.Run("call auto rotate the main token and rotate the token", func(t *testing.T) {
		var client = newInMemoryClient(true)
		ctx, url := getCtxGitlabClientWithUrl(t)
		ctx = gitlab.GitlabClientNewContext(ctx, newInMemoryClient(true))
		b, l, events, err := getBackendWithEventsAndConfig(ctx, map[string]any{
			"token":              "token",
			"base_url":           url,
			"auto_rotate_token":  true,
			"auto_rotate_before": "360h",
			"type":               gitlab.TypeSelfManaged.String(),
		})
		require.NoError(t, err)

		client.rotateMainToken.Token = "new token"
		b.SetClient(client, gitlab.DefaultConfigName)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.NotEmpty(t, resp.Data["token_expires_at"])

		err = b.PeriodicFunc(ctx, &logical.Request{Storage: l})
		require.NoError(t, err)
		assert.Greater(t, client.calledRotateMainToken, 0)

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.NotEmpty(t, resp.Data["token_sha1_hash"])
		require.NotEmpty(t, resp.Data["token_expires_at"])

		events.expectEvents(t, []expectedEvent{
			{
				eventType: "gitlab/config-write",
			},
			{
				eventType: "gitlab/config-token-rotate",
			},
		})
	})

	t.Run("call auto rotate the main token but the token is still valid", func(t *testing.T) {
		var client = newInMemoryClient(true)
		ctx, url := getCtxGitlabClientWithUrl(t)
		ctx = gitlab.GitlabClientNewContext(ctx, newInMemoryClient(true))
		b, l, err := getBackendWithConfig(ctx, map[string]any{
			"token":              "token",
			"base_url":           url,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
			"type":               gitlab.TypeSelfManaged.String(),
		})
		require.NoError(t, err)

		var expiresAt = time.Now().Add(100 * 24 * time.Hour)
		client.mainTokenInfo.ExpiresAt = &expiresAt
		b.SetClient(client, gitlab.DefaultConfigName)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.NotEmpty(t, resp.Data["token_expires_at"])

		err = b.PeriodicFunc(ctx, &logical.Request{Storage: l})
		require.NoError(t, err)
		assert.EqualValues(t, 1, client.calledRotateMainToken)

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, resp.Data)
		require.NotEmpty(t, resp.Data["token_expires_at"])
	})

}
