//go:build paths

package integration_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestPathTokenRolesMultipleConfigs(t *testing.T) {
	httpClient, gitlabUrl := getClient(t, "paths")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)

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
		var gitlabType = gitlabTypes.TypeSelfManaged.String()
		t.Logf("writing backend config %q (base_url=%s, type=%s, token=%s)", name, gitlabUrl, gitlabType, token)
		require.NoError(t,
			writeBackendConfigWithName(ctx, b, l,
				map[string]any{
					"token":    token,
					"base_url": gitlabUrl,
					"type":     gitlabType,
				},
				name,
			),
		)
	}

	type roleData struct {
		roleName, path, tokenName string
		tokenType                 token.Type
		accessLevel               token.AccessLevel
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
				accessLevel: token.AccessLevelGuestPermissions,
				scopes:      []string{token.ScopeApi.String(), token.ScopeSelfRotate.String()},
				tokenName:   "admin_user_initial_token",
			},
		},
		"normal": {
			{
				roleName:    "normal-example",
				path:        "example",
				tokenType:   token.TypeGroup,
				accessLevel: token.AccessLevelGuestPermissions,
				scopes:      []string{token.ScopeApi.String(), token.ScopeSelfRotate.String()},
				tokenName:   "normal_user_initial_token",
			},
		},
	}

	for cfg, rds := range roles {
		for _, rd := range rds {
			t.Logf("=== config=%q role=%q path=%q token_type=%s ===", cfg, rd.roleName, rd.path, rd.tokenType.String())
			var tokenName = rd.tokenName
			var data = map[string]any{
				"name":       fmt.Sprintf("%s-{{ .role_name }}-{{ .config_name }}-{{ .token_type }}", rd.path),
				"token_type": rd.tokenType.String(), "path": rd.path, "config_name": cfg, "ttl": backend.DefaultAccessTokenMinTTL,
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

			t.Logf("creating role %q under config %q with data=%+v", rd.roleName, cfg, data)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%s", backend.PathRoleStorage, rd.roleName), Storage: l,
				Data: data,
			})
			require.NoError(t, err, "create role %q failed", rd.roleName)
			require.NotNil(t, resp, "create role %q returned nil response", rd.roleName)
			require.NoError(t, resp.Error(), "create role %q response error", rd.roleName)
			require.Empty(t, resp.Warnings, "create role %q produced warnings", rd.roleName)
			require.EqualValues(t, cfg, resp.Data["config_name"], "create role %q config_name mismatch", rd.roleName)

			t.Logf("issuing token for role %q (impersonating %q)", rd.roleName, tokenName)
			ctxIssueToken, _ := ctxTestTime(ctx, t.Name(), tokenName)
			resp, err = b.HandleRequest(ctxIssueToken, &logical.Request{
				Operation: logical.ReadOperation, Storage: l,
				Path: fmt.Sprintf("%s/%s", tokenPaths.PathTokenRoleStorage, rd.roleName),
			})
			require.NoError(t, err, "issue token for role %q failed", rd.roleName)
			require.NotNil(t, resp, "issue token for role %q returned nil response", rd.roleName)
			require.NotNil(t, resp.Secret, "issue token for role %q returned nil secret", rd.roleName)
			require.NoError(t, resp.Error(), "issue token for role %q response error", rd.roleName)

			var token = resp.Data["token"].(string)
			require.NotEmpty(t, token, "issue token for role %q produced empty token", rd.roleName)
			var secret = resp.Secret
			require.NotNil(t, secret)
			t.Logf("issued token for role %q (lease_id=%s, ttl=%s)", rd.roleName, secret.LeaseID, secret.LeaseOptions.TTL)

			// verify token that it works
			t.Logf("verifying issued token for role %q against gitlab", rd.roleName)
			var c *g.Client
			c, err = g.NewClient(token, g.WithHTTPClient(httpClient), g.WithBaseURL(gitlabUrl))
			require.NoError(t, err, "new gitlab client for role %q failed", rd.roleName)
			require.NotNil(t, c)

			pat, r, err := c.PersonalAccessTokens.GetSinglePersonalAccessToken()
			require.NoError(t, err, "GetSinglePersonalAccessToken for role %q failed", rd.roleName)
			require.NotNil(t, r, "GetSinglePersonalAccessToken for role %q returned nil http response", rd.roleName)
			require.NotNil(t, pat, "GetSinglePersonalAccessToken for role %q returned nil pat", rd.roleName)
			t.Logf("verified token for role %q (gitlab status=%s pat_id=%d)", rd.roleName, r.Status, pat.ID)

			t.Logf("revoking token for role %q (lease_id=%s)", rd.roleName, secret.LeaseID)
			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.RevokeOperation,
				Path:      "/",
				Storage:   l,
				Secret:    secret,
			})
			require.NoError(t, err, "revoke token for role %q failed", rd.roleName)
			require.Nil(t, resp, "revoke token for role %q expected nil response", rd.roleName)

		}
	}

	t.Logf("listing roles at %q", backend.PathRoleStorage)
	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.ListOperation,
		Path:      backend.PathRoleStorage, Storage: l,
	})
	require.NoError(t, err, "list roles failed")
	require.NotNil(t, resp, "list roles returned nil response")
	require.NoError(t, resp.Error(), "list roles response error")
	require.NotEmpty(t, resp.Data, "list roles returned empty data")
	t.Logf("list roles returned: %+v", resp.Data)

	t.Logf("verifying expected events sequence")
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
