package gitlab_test

import (
	"context"
	"github.com/hashicorp/vault/sdk/logical"
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
				"auto_rotate_before": "48h",
			},
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)
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
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Data["auto_rotate_before"])
		assert.EqualValues(t, "10h0m0s", resp.Data["auto_rotate_before"])
	})
}
