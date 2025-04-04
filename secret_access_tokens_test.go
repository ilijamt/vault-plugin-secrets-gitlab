//go:build unit

package gitlab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestSecretAccessTokenRevokeToken(t *testing.T) {
	httpClient, url := getClient(t, "unit")
	ctx := gitlab.HttpClientNewContext(t.Context(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	t.Run("nil storage", func(t *testing.T) {
		events.resetEvents(t)
		resp, err := b.Secret(gitlab.SecretAccessTokenType).HandleRevoke(ctx, &logical.Request{})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
		events.expectEvents(t, []expectedEvent{})
	})

	t.Run("nil secret", func(t *testing.T) {
		events.resetEvents(t)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":              getGitlabToken("admin_user_root").Token,
				"base_url":           url,
				"auto_rotate_token":  true,
				"auto_rotate_before": "24h",
				"type":               gitlab.TypeSelfManaged.String(),
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, events)

		resp, err = b.Secret(gitlab.SecretAccessTokenType).HandleRevoke(ctx, &logical.Request{Storage: l})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, gitlab.ErrNilValue)

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
		})

	})

	t.Run("token_id invalid value", func(t *testing.T) {
		events.resetEvents(t)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":              getGitlabToken("admin_user_root").Token,
				"base_url":           url,
				"auto_rotate_token":  true,
				"auto_rotate_before": "24h",
				"type":               gitlab.TypeSelfManaged.String(),
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotEmpty(t, events)

		resp, err = b.Secret(gitlab.SecretAccessTokenType).HandleRevoke(ctx, &logical.Request{
			Storage: l,
			Secret: &logical.Secret{
				InternalData: map[string]interface{}{
					"token_id": "asdf",
				},
			},
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, gitlab.ErrInvalidValue)

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
		})
	})

}
