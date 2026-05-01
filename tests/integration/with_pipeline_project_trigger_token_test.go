//go:build e2e

package integration_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithPipelineProjectTriggerAccessToken(t *testing.T) {
	httpClient, url := getClient(t, "e2e")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = "normal_user_initial_token"

	b, l, events, err := getBackendWithEventsAndConfig(ctx, map[string]any{
		"token":              getGitlabToken(tokenName).Token,
		"base_url":           url,
		"auto_rotate_token":  true,
		"auto_rotate_before": "24h",
		"type":               gitlabTypes.TypeSelfManaged.String(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, events)

	var c *g.Client
	var token string
	var secret *logical.Secret

	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/pptat", backend.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "example/example",
				"name":                 token2.TypePipelineProjectTrigger.String(),
				"token_type":           token2.TypePipelineProjectTrigger.String(),
				"gitlab_revokes_token": strconv.FormatBool(false),
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
			Path: fmt.Sprintf("%s/pptat", tokenPaths.PathTokenRoleStorage),
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
		tt, _, err := c.PipelineTriggers.ListPipelineTriggers("example/example", &g.ListPipelineTriggersOptions{})
		require.NoError(t, err)
		require.Len(t, tt, 1)
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
		tt, _, err := c.PipelineTriggers.ListPipelineTriggers("example/example", &g.ListPipelineTriggersOptions{})
		require.NoError(t, err)
		require.Len(t, tt, 0)
	}

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
