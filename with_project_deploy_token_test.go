//go:build local

package gitlab_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestWithProjectDeployToken(t *testing.T) {
	httpClient, url := getClient(t, "local")
	ctx := gitlab.HttpClientNewContext(t.Context(), httpClient)
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

	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/role", gitlab.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "example/example",
				"name":                 gitlab.TokenTypeProjectDeploy.String(),
				"token_type":           gitlab.TokenTypeProjectDeploy.String(),
				"gitlab_revokes_token": strconv.FormatBool(false),
				"ttl":                  120 * time.Hour,
				"scopes":               []string{token2.TokenScopeReadRepository.String()},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
	}

	{
		ctxIssueToken, _ := ctxTestTime(ctx, t.Name(), tokenName)
		resp, err := b.HandleRequest(ctxIssueToken, &logical.Request{
			Operation: logical.ReadOperation, Storage: l,
			Path: fmt.Sprintf("%s/role", gitlab.PathTokenRoleStorage),
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		token = resp.Data["token"].(string)
		require.NotEmpty(t, token)
		secret = resp.Secret
		require.NotNil(t, secret)
	}

	c = b.GetClient(gitlab.DefaultConfigName).GitlabClient(ctx)
	require.NotNil(t, c)

	{
		tt, _, err := c.DeployTokens.ListProjectDeployTokens("example/example", &g.ListProjectDeployTokensOptions{})
		require.NoError(t, err)
		out := filterSlice(tt, func(item *g.DeployToken, index int) bool { return !item.Expired && !item.Revoked })
		require.Len(t, out, 1)
	}

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

	{
		tt, _, err := c.DeployTokens.ListProjectDeployTokens("example/example", &g.ListProjectDeployTokensOptions{})
		require.NoError(t, err)
		out := filterSlice(tt, func(item *g.DeployToken, index int) bool { return !item.Expired && !item.Revoked })
		require.Len(t, out, 0)
	}

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
