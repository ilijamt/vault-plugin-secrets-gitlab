//go:build selfhosted

package integration_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithServiceAccountProject(t *testing.T) {
	httpClient, _ := getClient(t, "selfhosted")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = ""

	b, l, events, err := getBackendWithEventsAndConfig(ctx, map[string]any{
		"token":              gitlabServiceAccountToken,
		"base_url":           gitlabServiceAccountUrl,
		"auto_rotate_token":  true,
		"auto_rotate_before": "24h",
		"type":               gitlabTypes.TypeSelfManaged.String(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, events)

	require.Nil(t, b.GetClient(backend.DefaultConfigName))
	var client glab.Client
	client, err = b.GetClientByName(ctx, l, backend.DefaultConfigName)
	require.NoError(t, err)
	require.NotNil(t, client)
	var gClient = client.GitlabClient(ctx)
	require.NotNil(t, gClient)

	// Project service accounts require GitLab 18.11+.
	md, err := client.Metadata(ctx)
	require.NoError(t, err)
	require.NotNil(t, md)
	if !gitlabVersionAtLeast(md.Version, projectServiceAccountMinVersion) {
		t.Skipf("project service accounts require GitLab >= %s, instance reports %s", projectServiceAccountMinVersion, md.Version)
	}
	t.Setenv("GITLAB_VERSION", md.Version)

	// Project bootstrapped by local-env/tf/_shared/service_accounts.tf
	projectId, err := client.GetProjectIdByPath(ctx, "service-accounts/project")
	require.NoError(t, err)
	var pid = strconv.FormatInt(projectId, 10)

	// Create a project service account
	sa, _, err := gClient.Projects.CreateProjectServiceAccount(pid, &g.CreateProjectServiceAccountOptions{})
	require.NoError(t, err)
	require.NotNil(t, sa)

	t.Cleanup(func() {
		_, _ = gClient.Projects.DeleteProjectServiceAccount(pid, sa.ID, nil)
	})

	// Create a project service account role
	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/project-service-account", backend.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 fmt.Sprintf("%s/%s", pid, sa.Username),
			"name":                 `vault-generated-{{ .token_type }}-token`,
			"token_type":           token.TypeProjectServiceAccount.String(),
			"ttl":                  backend.DefaultAccessTokenMinTTL,
			"scopes":               validScopesFor(token.TypeProjectServiceAccount),
			"gitlab_revokes_token": false,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.Empty(t, resp.Warnings)
	require.EqualValues(t, resp.Data["config_name"], backend.DefaultConfigName)

	// Get a new token for the service account
	ctxIssueToken, _ := ctxTestTime(ctx, t, tokenName)
	resp, err = b.HandleRequest(ctxIssueToken, &logical.Request{
		Operation: logical.ReadOperation, Storage: l,
		Path: fmt.Sprintf("%s/project-service-account", tokenPaths.PathTokenRoleStorage),
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	var newToken = resp.Data["token"].(string)
	require.NotEmpty(t, newToken)
	secret := resp.Secret
	require.NotNil(t, secret)

	{
		// Validate that the new token works
		c, err := g.NewClient(newToken, g.WithHTTPClient(httpClient), g.WithBaseURL(gitlabServiceAccountUrl))
		require.NoError(t, err)
		require.NotNil(t, c)
		pat, resp, err := c.PersonalAccessTokens.GetSinglePersonalAccessToken()
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, pat)
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

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
