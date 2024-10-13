package gitlab_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "github.com/xanzy/go-gitlab"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathTokenRolesMultipleConfigs(t *testing.T) {
	httpClient, gitlabUrl := getClient(t)
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.NotNil(t, b)
	require.NotNil(t, l)

	var configs = map[string]string{"root": "glpat-secret-random-token", "admin": "glpat-secret-admin-token", "normal": "glpat-secret-normal-token"}
	for name, token := range configs {
		require.NoError(t,
			writeBackendConfigWithName(ctx, b, l,
				map[string]any{
					"token":    token,
					"base_url": gitlabUrl,
					"type":     gitlab.TypeSelfManaged.String(),
				},
				name,
			),
		)
	}

	type roleData struct {
		rn, path string
		tt       gitlab.TokenType
		al       gitlab.AccessLevel
		scopes   []string
	}
	var roles = map[string][]roleData{
		"root": {
			{rn: "root-root", path: "root", tt: gitlab.TokenTypePersonal, scopes: gitlab.ValidPersonalTokenScopes},
			{rn: "root-normal-user", path: "normal-user", tt: gitlab.TokenTypePersonal, scopes: gitlab.ValidPersonalTokenScopes},
		},
		"admin": {
			{rn: "admin-example-example", path: "example/example", tt: gitlab.TokenTypeProject, al: gitlab.AccessLevelGuestPermissions, scopes: []string{gitlab.TokenScopeApi.String()}},
		},
		"normal": {
			{rn: "normal-example", path: "example", tt: gitlab.TokenTypeGroup, al: gitlab.AccessLevelGuestPermissions, scopes: []string{gitlab.TokenScopeApi.String()}},
		},
	}

	for cfg, rds := range roles {
		for _, rd := range rds {
			var data = map[string]any{
				"name":       fmt.Sprintf("%s-{{ .role_name }}-{{ .config_name }}-{{ .token_type }}", rd.path),
				"token_type": rd.tt.String(), "path": rd.path, "config_name": cfg, "ttl": gitlab.DefaultAccessTokenMinTTL,
			}

			switch rd.tt {
			case gitlab.TokenTypePersonal:
				data["access_level"] = rd.al.String()
				data["scopes"] = rd.scopes
			case gitlab.TokenTypeGroup:
				data["access_level"] = rd.al.String()
				data["scopes"] = rd.scopes
			case gitlab.TokenTypeProject:
				data["access_level"] = rd.al.String()
				data["scopes"] = rd.scopes
			}

			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%s", gitlab.PathRoleStorage, rd.rn), Storage: l,
				Data: data,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.Empty(t, resp.Warnings)
			require.EqualValues(t, cfg, resp.Data["config_name"])

			ctxIssueToken, _ := ctxTestTime(ctx, t.Name())
			resp, err = b.HandleRequest(ctxIssueToken, &logical.Request{
				Operation: logical.ReadOperation, Storage: l,
				Path: fmt.Sprintf("%s/%s", gitlab.PathTokenRoleStorage, rd.rn),
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Secret)
			require.NoError(t, resp.Error())

			var token = resp.Data["token"].(string)
			require.NotEmpty(t, token)
			var secret = resp.Secret
			require.NotNil(t, secret)

			// verify token that it works
			var c *g.Client
			c, err = g.NewClient(token, g.WithHTTPClient(httpClient), g.WithBaseURL(gitlabUrl))
			require.NoError(t, err)
			require.NotNil(t, c)

			pat, r, err := c.PersonalAccessTokens.GetSinglePersonalAccessToken()
			require.NoError(t, err)
			require.NotNil(t, r)
			require.NotNil(t, pat)

			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.RevokeOperation,
				Path:      "/",
				Storage:   l,
				Secret:    secret,
			})
			require.NoError(t, err)
			require.Nil(t, resp)

		}
	}

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.ListOperation,
		Path:      gitlab.PathRoleStorage, Storage: l,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, resp.Data)

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
