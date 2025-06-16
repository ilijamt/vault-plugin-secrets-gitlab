//go:build selfhosted

package gitlab_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestWithServiceAccountUser(t *testing.T) {
	httpClient, _ := getClient(t, "selfhosted")
	ctx := gitlab.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = ""

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		Data: map[string]any{
			"token":              gitlabServiceAccountToken,
			"base_url":           gitlabServiceAccountUrl,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
			"type":               gitlab2.TypeSelfManaged.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	require.NotNil(t, b.GetClient(gitlab.DefaultConfigName))
	var gClient = b.GetClient(gitlab.DefaultConfigName).GitlabClient(ctx)
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
		Path:      fmt.Sprintf("%s/user-service-account", gitlab.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 usr.Username,
			"name":                 `vault-generated-{{ .token_type }}-token`,
			"token_type":           token2.TypeUserServiceAccount.String(),
			"ttl":                  gitlab.DefaultAccessTokenMinTTL,
			"scopes":               token2.ValidUserServiceAccountTokenScopes,
			"gitlab_revokes_token": false,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.Empty(t, resp.Warnings)
	require.EqualValues(t, resp.Data["config_name"], gitlab.TypeConfigDefault)

	// Get a new token for the service account
	ctxIssueToken, _ := ctxTestTime(ctx, t.Name(), tokenName)
	resp, err = b.HandleRequest(ctxIssueToken, &logical.Request{
		Operation: logical.ReadOperation, Storage: l,
		Path: fmt.Sprintf("%s/user-service-account", gitlab.PathTokenRoleStorage),
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	var token = resp.Data["token"].(string)
	require.NotEmpty(t, token)
	secret := resp.Secret
	require.NotNil(t, secret)

	{
		// Validate that the new token works
		c, err := g.NewClient(token, g.WithHTTPClient(httpClient), g.WithBaseURL(gitlabServiceAccountUrl))
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
