//go:build serviceaccount

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

// serviceAccountCase captures what differs between the user, group and project
// service-account flows; the rest of the lifecycle is shared.
type serviceAccountCase struct {
	roleName  string
	tokenType token.Type
	// setupSA provisions the service account, registers its teardown, and
	// returns the GitLab path the role should target.
	setupSA func(t *testing.T, ctx context.Context, client glab.Client, gClient *g.Client) string
}

func runServiceAccountTokenTest(t *testing.T, tc serviceAccountCase) {
	t.Helper()
	requireServiceAccounts(t)
	httpClient, url := getClient(t, "serviceaccount")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)

	b, l, events, err := getBackendWithEventsAndConfig(ctx,
		standardConfig(gitlabTypes.TypeSelfManaged, url, getGitlabToken("admin_user_root").Token))
	require.NoError(t, err)

	require.Nil(t, b.GetClient(backend.DefaultConfigName))
	client, err := b.GetClientByName(ctx, l, backend.DefaultConfigName)
	require.NoError(t, err)
	require.NotNil(t, client)
	gClient := client.GitlabClient(ctx)
	require.NotNil(t, gClient)

	path := tc.setupSA(t, ctx, client, gClient)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/%s", backend.PathRoleStorage, tc.roleName), Storage: l,
		Data: map[string]any{
			"path":                 path,
			"name":                 `vault-generated-{{ .token_type }}-token`,
			"token_type":           tc.tokenType.String(),
			"ttl":                  backend.DefaultAccessTokenMinTTL,
			"scopes":               []string{token.ScopeApi.String(), token.ScopeReadApi.String()},
			"gitlab_revokes_token": false,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.Empty(t, resp.Warnings)
	require.EqualValues(t, resp.Data["config_name"], backend.DefaultConfigName)

	newToken, secret := issueToken(ctx, t, b, l, "serviceaccount", tc.roleName)
	requireTokenStatus(t, httpClient, url, newToken, http.StatusOK)
	revokeSecret(ctx, t, b, l, secret)

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
