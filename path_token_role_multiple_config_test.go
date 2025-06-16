//go:build unit

package gitlab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestPathTokenRolesMultipleConfigs(t *testing.T) {
	httpClient, gitlabUrl := getClient(t, "unit")
	ctx := gitlab.HttpClientNewContext(t.Context(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.NotNil(t, b)
	require.NotNil(t, l)

	var configs = map[string]string{
		"root":   getGitlabToken("admin_user_root").Token,
		"admin":  getGitlabToken("admin_user_initial_token").Token,
		"normal": getGitlabToken("normal_user_initial_token").Token,
	}
	for name, token := range configs {
		require.NoError(t,
			writeBackendConfigWithName(ctx, b, l,
				map[string]any{
					"token":    token,
					"base_url": gitlabUrl,
					"type":     gitlab2.TypeSelfManaged.String(),
				},
				name,
			),
		)
	}

	type roleData struct {
		roleName, path, tokenName string
		tokenType                 token.Type
		accessLevel               gitlab.AccessLevel
		scopes                    []string
	}
	var roles = map[string][]roleData{
		"root": {
			{
				roleName:  "root-root",
				path:      "root",
				tokenType: token.TypePersonal,
				scopes:    []string{token.ScopeApi.String(), token.ScopeSelfRotate.String()},
				tokenName: "admin_user_root",
			},
			{
				roleName:  "root-normal-user",
				path:      "normal-user",
				tokenType: token.TypePersonal,
				scopes:    []string{token.ScopeApi.String(), token.ScopeSelfRotate.String()},
				tokenName: "admin_user_root",
			},
		},
		"admin": {
			{
				roleName:    "admin-example-example",
				path:        "example/example",
				tokenType:   token.TypeProject,
				accessLevel: gitlab.AccessLevelGuestPermissions,
				scopes:      []string{token.ScopeApi.String(), token.ScopeSelfRotate.String()},
				tokenName:   "admin_user_initial_token",
			},
		},
		"normal": {
			{
				roleName:    "normal-example",
				path:        "example",
				tokenType:   token.TypeGroup,
				accessLevel: gitlab.AccessLevelGuestPermissions,
				scopes:      []string{token.ScopeApi.String(), token.ScopeSelfRotate.String()},
				tokenName:   "normal_user_initial_token",
			},
		},
	}

	for cfg, rds := range roles {
		for _, rd := range rds {
			var tokenName = rd.tokenName
			var data = map[string]any{
				"name":       fmt.Sprintf("%s-{{ .role_name }}-{{ .config_name }}-{{ .token_type }}", rd.path),
				"token_type": rd.tokenType.String(), "path": rd.path, "config_name": cfg, "ttl": gitlab.DefaultAccessTokenMinTTL,
			}

			switch rd.tokenType {
			case token.TypePersonal:
				data["access_level"] = rd.accessLevel.String()
				data["scopes"] = rd.scopes
			case token.TypeGroup:
				data["access_level"] = rd.accessLevel.String()
				data["scopes"] = rd.scopes
			case token.TypeProject:
				data["access_level"] = rd.accessLevel.String()
				data["scopes"] = rd.scopes
			}

			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%s", gitlab.PathRoleStorage, rd.roleName), Storage: l,
				Data: data,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.Empty(t, resp.Warnings)
			require.EqualValues(t, cfg, resp.Data["config_name"])

			ctxIssueToken, _ := ctxTestTime(ctx, t.Name(), tokenName)
			resp, err = b.HandleRequest(ctxIssueToken, &logical.Request{
				Operation: logical.ReadOperation, Storage: l,
				Path: fmt.Sprintf("%s/%s", gitlab.PathTokenRoleStorage, rd.roleName),
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
