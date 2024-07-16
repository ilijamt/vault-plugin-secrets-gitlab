package gitlab_test

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "github.com/xanzy/go-gitlab"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

var (
	gitlabComPersonalAccessToken = cmp.Or(os.Getenv("GITLAB_COM_TOKEN"), "glpat-invalid-value")
	gitlabComUrl                 = cmp.Or(os.Getenv("GITLAB_COM_URL"), "https://gitlab.com")
)

func TestWithGitlabUser_RotateToken(t *testing.T) {
	httpClient, _ := getClient(t)
	url := gitlabComUrl
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigStorage, Storage: l,
		Data: map[string]any{
			"token":              gitlabComPersonalAccessToken,
			"base_url":           url,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	var oldToken, newToken string

	// Rotate the main token
	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/rotate", gitlab.PathConfigStorage), Storage: l,
			Data: map[string]any{},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEqualValues(t, resp.Data["token"], gitlabComPersonalAccessToken)
		oldToken = gitlabComPersonalAccessToken
		require.NotNil(t, resp.Secret)
		require.NotNil(t, resp.Secret.InternalData)
		require.NotEmpty(t, resp.Secret.InternalData["token"])
		newToken = resp.Secret.InternalData["token"].(string)
	}

	// Old token should not have access anymore
	{
		c, err := g.NewClient(oldToken, g.WithHTTPClient(httpClient), g.WithBaseURL(url))
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
		c, err := g.NewClient(newToken, g.WithHTTPClient(httpClient), g.WithBaseURL(url))
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
