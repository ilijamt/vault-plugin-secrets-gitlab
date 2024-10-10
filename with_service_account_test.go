package gitlab_test

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "github.com/xanzy/go-gitlab"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestWithUserServiceAccountFailWithSaas(t *testing.T) {
	url := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_URL"), "https://gitlab.com")
	accountToken := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_TOKEN"), "glpat-invalid-value")

	httpClient, _ := getClient(t)
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigStorage, Storage: l,
		Data: map[string]any{
			"token":              accountToken,
			"base_url":           url,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
			"type":               gitlab.TypeSaaS.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	require.NotNil(t, b.GetClient())
	var gClient = b.GetClient().GitlabClient()
	require.NotNil(t, gClient)

	// Create a service account user
	usr, _, err := gClient.Users.CreateServiceAccountUser()
	require.NoError(t, err)
	require.NotNil(t, usr)

	// Create a user service account role
	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/user-service-account", gitlab.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 usr.Username,
			"name":                 `vault-generated-{{ .token_type }}-token-{{ randHexString 4 }}`,
			"token_type":           gitlab.TokenTypeUserServiceAccount.String(),
			"ttl":                  gitlab.DefaultAccessTokenMinTTL,
			"scopes":               gitlab.ValidUserServiceAccountTokenScopes,
			"gitlab_revokes_token": false,
		},
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Error(t, resp.Error())
	require.Empty(t, resp.Warnings)

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
	})
}

func TestWithUserServiceAccountFailWithDedicated(t *testing.T) {
	url := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_URL"), "https://gitlab.com")
	accountToken := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_TOKEN"), "glpat-invalid-value")

	httpClient, _ := getClient(t)
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigStorage, Storage: l,
		Data: map[string]any{
			"token":              accountToken,
			"base_url":           url,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
			"type":               gitlab.TypeDedicated.String(),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	require.NotNil(t, b.GetClient())
	var gClient = b.GetClient().GitlabClient()
	require.NotNil(t, gClient)

	// Create a service account user
	usr, _, err := gClient.Users.CreateServiceAccountUser()
	require.NoError(t, err)
	require.NotNil(t, usr)

	// Create a user service account role
	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/user-service-account", gitlab.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 usr.Username,
			"name":                 `vault-generated-{{ .token_type }}-token-{{ randHexString 4 }}`,
			"token_type":           gitlab.TokenTypeUserServiceAccount.String(),
			"ttl":                  gitlab.DefaultAccessTokenMinTTL,
			"scopes":               gitlab.ValidUserServiceAccountTokenScopes,
			"gitlab_revokes_token": false,
		},
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Error(t, resp.Error())
	require.Empty(t, resp.Warnings)

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
	})
}

func TestWithUserServiceAccount(t *testing.T) {
	url := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_URL"), "https://gitlab.com")
	accountToken := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_TOKEN"), "glpat-invalid-value")

	httpClient, _ := getClient(t)
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigStorage, Storage: l,
		Data: map[string]any{
			"token":              accountToken,
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

	require.NotNil(t, b.GetClient())
	var gClient = b.GetClient().GitlabClient()
	require.NotNil(t, gClient)

	// Create a service account user
	usr, _, err := gClient.Users.CreateServiceAccountUser()
	require.NoError(t, err)
	require.NotNil(t, usr)

	// Create a user service account role
	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/user-service-account", gitlab.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 usr.Username,
			"name":                 `vault-generated-{{ .token_type }}-token-{{ randHexString 4 }}`,
			"token_type":           gitlab.TokenTypeUserServiceAccount.String(),
			"ttl":                  gitlab.DefaultAccessTokenMinTTL,
			"scopes":               gitlab.ValidUserServiceAccountTokenScopes,
			"gitlab_revokes_token": false,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.Empty(t, resp.Warnings)
	require.EqualValues(t, resp.Data["config"], gitlab.TypeConfigDefault)

	// Get a new token for the service account
	resp, err = b.HandleRequest(ctx, &logical.Request{
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
		c, err := g.NewClient(token, g.WithHTTPClient(httpClient), g.WithBaseURL(url))
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

func TestWithGroupServiceAccount(t *testing.T) {
	url := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_URL"), "https://gitlab.com")
	accountToken := cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_TOKEN"), "glpat-invalid-value")

	httpClient, _ := getClient(t)
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigStorage, Storage: l,
		Data: map[string]any{
			"token":              accountToken,
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

	require.NotNil(t, b.GetClient())
	var gClient = b.GetClient().GitlabClient()
	require.NotNil(t, gClient)

	// Create a group service account
	var gid = strconv.Itoa(265)
	sa, _, err := gClient.Groups.CreateServiceAccount(gid, &g.CreateServiceAccountOptions{})
	require.NoError(t, err)
	require.NotNil(t, sa)

	// Create a group service account role
	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("%s/group-service-account", gitlab.PathRoleStorage), Storage: l,
		Data: map[string]any{
			"path":                 fmt.Sprintf("%s/%s", gid, sa.UserName),
			"name":                 `vault-generated-{{ .token_type }}-token-{{ randHexString 4 }}`,
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
	require.EqualValues(t, resp.Data["config"], gitlab.TypeConfigDefault)

	// Get a new token for the service account
	resp, err = b.HandleRequest(ctx, &logical.Request{
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
		c, err := g.NewClient(token, g.WithHTTPClient(httpClient), g.WithBaseURL(url))
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
