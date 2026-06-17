//go:build e2e

package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithProjectDeployToken(t *testing.T) {
	httpClient, url := getClient(t, "e2e")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)

	b, l, events, err := getBackendWithEventsAndConfig(ctx,
		standardConfig(gitlabTypes.TypeSelfManaged, url, getGitlabToken("normal_user_initial_token").Token))
	require.NoError(t, err)

	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/role", backend.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "example/example",
				"name":                 token2.TypeProjectDeploy.String(),
				"token_type":           token2.TypeProjectDeploy.String(),
				"gitlab_revokes_token": false,
				"ttl":                  120 * time.Hour,
				"scopes":               []string{token2.ScopeReadRepository.String()},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
	}

	_, secret := issueToken(ctx, t, b, l, "e2e", "role")

	c := b.GetClient(backend.DefaultConfigName).GitlabClient(ctx)
	require.NotNil(t, c)

	liveProjectDeployTokens := func() int {
		tt, _, err := c.DeployTokens.ListProjectDeployTokens("example/example", &g.ListProjectDeployTokensOptions{})
		require.NoError(t, err)
		return len(filterSlice(tt, func(item *g.DeployToken, index int64) bool { return !item.Expired && !item.Revoked }))
	}

	require.Equal(t, 1, liveProjectDeployTokens())
	revokeSecret(ctx, t, b, l, secret)
	require.Equal(t, 0, liveProjectDeployTokens())

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
