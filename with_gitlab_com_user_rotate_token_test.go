//go:build saas

package gitlab_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithGitlabUser_RotateToken(t *testing.T) {
	httpClient, _ := getClient(t, "saas")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = ""

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName),
		Storage:   l,
		Data: map[string]any{
			"token":              gitlabComPersonalAccessToken,
			"base_url":           gitlabComUrl,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
			"type":               gitlab2.TypeSaaS.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	var oldToken, newToken string

	// Rotate the main token
	{
		ctxRotate, _ := ctxTestTime(ctx, t.Name(), tokenName)
		resp, err := b.HandleRequest(ctxRotate, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s/rotate", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEqualValues(t, resp.Data["token"], gitlabComPersonalAccessToken)
		oldToken = gitlabComPersonalAccessToken
		newToken = resp.Data["token"].(string)
		require.Nil(t, resp.Secret) // This must not be a secret
	}

	// Old token should not have access anymore
	{
		c, err := g.NewClient(oldToken, g.WithHTTPClient(httpClient), g.WithBaseURL(gitlabComUrl))
		require.NoError(t, err)
		require.NotNil(t, c)
		pat, r, err := c.PersonalAccessTokens.GetSinglePersonalAccessToken()
		require.Error(t, err)
		require.Nil(t, pat)
		require.NotNil(t, r)
		require.EqualValues(t, r.StatusCode, http.StatusUnauthorized)
	}

	// New token should have access
	{
		c, err := g.NewClient(newToken, g.WithHTTPClient(httpClient), g.WithBaseURL(gitlabComUrl))
		require.NoError(t, err)
		require.NotNil(t, c)
		pat, r, err := c.PersonalAccessTokens.GetSinglePersonalAccessToken()
		require.NoError(t, err)
		require.NotNil(t, pat)
		require.NotNil(t, r)
		require.EqualValues(t, r.StatusCode, http.StatusOK)
	}

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/config-token-rotate"},
	})
}
