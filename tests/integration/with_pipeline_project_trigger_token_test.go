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
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithPipelineProjectTriggerAccessToken(t *testing.T) {
	httpClient, url := getClient(t, "e2e")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)

	b, l, events, err := getBackendWithEventsAndConfig(ctx,
		standardConfig(gitlabTypes.TypeSelfManaged, url, getGitlabToken("normal_user_initial_token").Token))
	require.NoError(t, err)

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

	_, secret := issueToken(ctx, t, b, l, "e2e", "pptat")

	c := b.GetClient(backend.DefaultConfigName).GitlabClient(ctx)
	require.NotNil(t, c)

	livePipelineTriggers := func() int {
		tt, _, err := c.PipelineTriggers.ListPipelineTriggers("example/example", &g.ListPipelineTriggersOptions{})
		require.NoError(t, err)
		return len(tt)
	}

	require.Equal(t, 1, livePipelineTriggers())
	revokeSecret(ctx, t, b, l, secret)
	require.Equal(t, 0, livePipelineTriggers())

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
