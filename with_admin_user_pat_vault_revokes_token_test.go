//go:build local

package gitlab_test

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestWithAdminUser_PAT_AdminUser_VaultRevokesToken(t *testing.T) {
	httpClient, url := getClient(t, "local")
	ctx := gitlab.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = "admin_user_initial_token"

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		Data: map[string]any{
			"token":              getGitlabToken(tokenName).Token,
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

	var c *g.Client
	var token string
	var secret *logical.Secret

	// create the role
	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation, Storage: l,
			Path: fmt.Sprintf("%s/admin-user", gitlab.PathRoleStorage),
			Data: map[string]any{
				"path":       "admin-user",
				"name":       gitlab.TokenTypePersonal.String(),
				"token_type": gitlab.TokenTypePersonal.String(),
				"scopes": strings.Join(
					[]string{
						token2.TokenScopeReadApi.String(),
					},
					","),
				"ttl":                  time.Hour,
				"gitlab_revokes_token": strconv.FormatBool(false),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
	}

	// issue a personal access token
	{
		ctxIssueToken, _ := ctxTestTime(ctx, t.Name(), tokenName)
		resp, err := b.HandleRequest(ctxIssueToken, &logical.Request{
			Operation: logical.ReadOperation, Storage: l,
			Path: fmt.Sprintf("%s/admin-user", gitlab.PathTokenRoleStorage),
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		token = resp.Data["token"].(string)
		require.NotEmpty(t, token)
		secret = resp.Secret
		require.NotNil(t, secret)
	}

	c, err = g.NewClient(token, g.WithHTTPClient(httpClient), g.WithBaseURL(url))
	require.NoError(t, err)
	require.NotNil(t, c)

	// should have access with token to Gitlab
	{
		var pat *g.PersonalAccessToken
		var r *g.Response
		pat, r, err = c.PersonalAccessTokens.GetSinglePersonalAccessToken()
		require.NoError(t, err)
		require.NotNil(t, pat)
		require.NotNil(t, r)
		require.EqualValues(t, r.StatusCode, http.StatusOK)
	}

	// revoke the token
	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.RevokeOperation,
			Path:      "/",
			Storage:   l,
			Secret:    secret,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	}

	// no longer has access with token to Gitlab
	{
		var pat *g.PersonalAccessToken
		var r *g.Response
		pat, r, err = c.PersonalAccessTokens.GetSinglePersonalAccessToken()
		require.Error(t, err)
		require.Nil(t, pat)
		require.NotNil(t, r)
		require.EqualValues(t, r.StatusCode, http.StatusUnauthorized)
	}

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
