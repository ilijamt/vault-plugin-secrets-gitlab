//go:build selfhosted

package gitlab_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestWithServiceAccountGroup(t *testing.T) {
	httpClient, _ := getClient(t)
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

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
			"type":               gitlab.TypeSelfManaged.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	require.NotNil(t, b.GetClient(gitlab.DefaultConfigName))
	var gClient = b.GetClient(gitlab.DefaultConfigName).GitlabClient(ctx)
	require.NotNil(t, gClient)

	// Create a group service account
	var gid = strconv.Itoa(265)
	sa, _, err := gClient.Groups.CreateServiceAccount(gid, &g.CreateServiceAccountOptions{})
	require.NoError(t, err)
	require.NotNil(t, sa)

	t.Cleanup(func() {
		_, _ = gClient.Users.DeleteUser(sa.ID)
	})

	// Create a group service account role
	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/group-service-account", gitlab.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 fmt.Sprintf("%s/%s", gid, sa.UserName),
			"name":                 `vault-generated-{{ .token_type }}-token`,
			"token_type":           gitlab.TokenTypeGroupServiceAccount.String(),
			"ttl":                  gitlab.DefaultAccessTokenMinTTL,
			"scopes":               gitlab.ValidGroupServiceAccountTokenScopes,
			"gitlab_revokes_token": false,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.Empty(t, resp.Warnings)
	require.EqualValues(t, resp.Data["config_name"], gitlab.TypeConfigDefault)

	// Get a new token for the service account
	ctxIssueToken, _ := ctxTestTime(ctx, t.Name())
	resp, err = b.HandleRequest(ctxIssueToken, &logical.Request{
		Operation: logical.ReadOperation, Storage: l,
		Path: fmt.Sprintf("%s/group-service-account", gitlab.PathTokenRoleStorage),
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
