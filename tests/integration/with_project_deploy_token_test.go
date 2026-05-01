//go:build e2e

package integration_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithProjectDeployToken(t *testing.T) {
	httpClient, url := getClient(t, "e2e")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = "normal_user_initial_token"

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/%s", backend.PathConfigStorage, backend.DefaultConfigName), Storage: l,
		Data: map[string]any{
			"token":              getGitlabToken(tokenName).Token,
			"base_url":           url,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
			"type":               gitlabTypes.TypeSelfManaged.String(),
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
			Path:      fmt.Sprintf("%s/role", backend.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "example/example",
				"name":                 token2.TypeProjectDeploy.String(),
				"token_type":           token2.TypeProjectDeploy.String(),
				"gitlab_revokes_token": strconv.FormatBool(false),
				"ttl":                  120 * time.Hour,
				"scopes":               []string{token2.ScopeReadRepository.String()},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
	}

	{
		ctxIssueToken, _ := ctxTestTime(ctx, t, tokenName)
		resp, err := b.HandleRequest(ctxIssueToken, &logical.Request{
			Operation: logical.ReadOperation, Storage: l,
			Path: fmt.Sprintf("%s/role", tokenPaths.PathTokenRoleStorage),
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		token = resp.Data["token"].(string)
		require.NotEmpty(t, token)
		secret = resp.Secret
		require.NotNil(t, secret)
	}

	c = b.GetClient(backend.DefaultConfigName).GitlabClient(ctx)
	require.NotNil(t, c)

	{
		tt, _, err := c.DeployTokens.ListProjectDeployTokens("example/example", &g.ListProjectDeployTokensOptions{})
		require.NoError(t, err)
		out := filterSlice(tt, func(item *g.DeployToken, index int64) bool { return !item.Expired && !item.Revoked })
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
		out := filterSlice(tt, func(item *g.DeployToken, index int64) bool { return !item.Expired && !item.Revoked })
		require.Len(t, out, 0)
	}

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
