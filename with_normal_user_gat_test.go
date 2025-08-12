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
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithNormalUser_GAT(t *testing.T) {
	httpClient, url := getClient(t, "local")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = "normal_user_initial_token"

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
			"type":               gitlab2.TypeSelfManaged.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	var token string
	var secret *logical.Secret

	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/gat", gitlab.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "example",
				"name":                 `gat-token`,
				"token_type":           token2.TypeGroup.String(),
				"ttl":                  time.Hour * 120,
				"gitlab_revokes_token": strconv.FormatBool(false),
				"access_level":         token2.AccessLevelMaintainerPermissions.String(),
				"scopes":               strings.Join([]string{token2.ScopeReadApi.String()}, ","),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
	}

	// issue a group access token
	{
		ctxIssueToken, _ := ctxTestTime(ctx, t.Name(), tokenName)
		resp, err := b.HandleRequest(ctxIssueToken, &logical.Request{
			Operation: logical.ReadOperation, Storage: l,
			Path: fmt.Sprintf("%s/gat", gitlab.PathTokenRoleStorage),
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		token = resp.Data["token"].(string)
		require.NotEmpty(t, token)
		secret = resp.Secret
		require.NotNil(t, secret)
	}

	var c *g.Client
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
