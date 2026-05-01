//go:build selfhosted

package integration_test

import (
	"fmt"
	"net/http"
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

func TestWithServiceAccountUser(t *testing.T) {
	httpClient, _ := getClient(t, "selfhosted")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = ""

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/%s", backend.PathConfigStorage, backend.DefaultConfigName), Storage: l,
		Data: map[string]any{
			"token":              gitlabServiceAccountToken,
			"base_url":           gitlabServiceAccountUrl,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
			"type":               gitlabTypes.TypeSelfManaged.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	require.Nil(t, b.GetClient(backend.DefaultConfigName))
	var client glab.Client
	client, err = b.GetClientByName(ctx, l, backend.DefaultConfigName)
	require.NoError(t, err)
	require.NotNil(t, client)
	var gClient = client.GitlabClient(ctx)
	require.NotNil(t, gClient)

	// Create a service account user
	usr, _, err := gClient.Users.CreateServiceAccountUser(&g.CreateServiceAccountUserOptions{})
	require.NoError(t, err)
	require.NotNil(t, usr)

	t.Cleanup(func() {
		_, _ = gClient.Users.DeleteUser(usr.ID)
	})

	// Create a user service account role
	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/user-service-account", backend.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 usr.Username,
			"name":                 `vault-generated-{{ .token_type }}-token`,
			"token_type":           token.TypeUserServiceAccount.String(),
			"ttl":                  backend.DefaultAccessTokenMinTTL,
			"scopes":               validScopesFor(token.TypeUserServiceAccount),
			"gitlab_revokes_token": false,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.Empty(t, resp.Warnings)
	require.EqualValues(t, resp.Data["config_name"], backend.DefaultConfigName)

	// Get a new token for the service account
	ctxIssueToken, _ := ctxTestTime(ctx, t.Name(), tokenName)
	resp, err = b.HandleRequest(ctxIssueToken, &logical.Request{
		Operation: logical.ReadOperation, Storage: l,
		Path: fmt.Sprintf("%s/user-service-account", tokenPaths.PathTokenRoleStorage),
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
