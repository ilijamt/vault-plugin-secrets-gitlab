//go:build paths

package integration_test

import (
	"cmp"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	g "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestPathTokenRoles(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    getGitlabToken("admin_user_root").Token,
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		"type":     gitlabTypes.TypeSelfManaged.String(),
	}

	t.Run("role not found", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "paths")
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", tokenPaths.PathTokenRoleStorage), Storage: l,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, errs.ErrNotFound)
	})

	var generalTokenCreation = func(t *testing.T, tokenType token.Type, level token.AccessLevel, gitlabRevokesToken bool, path string, dynamicPath bool, pathExtra string) {
		t.Logf("token creation, token type: %s, level: %s, gitlab revokes token: %t, path: %s", tokenType, level, gitlabRevokesToken, path)
		ctx := getCtxGitlabClient(t, "paths")
		client := newInMemoryClient(t, true)
		ctx = g.ClientNewContext(ctx, client)
		var b, l, events, err = getBackendWithEvents(ctx)
		require.NoError(t, err)
		require.NoError(t, writeBackendConfig(ctx, b, l, defaultConfig))
		require.NoError(t, err)

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
		})

		var ttl = "1h"
		if gitlabRevokesToken {
			ttl = "48h"
		}

		roleName := "test"

		// create a role
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/%s", backend.PathRoleStorage, roleName), Storage: l,
			Data: map[string]any{
				"path":                 path,
				"name":                 tokenType.String(),
				"token_type":           tokenType.String(),
				"access_level":         level,
				"ttl":                  ttl,
				"gitlab_revokes_token": strconv.FormatBool(gitlabRevokesToken),
				"dynamic_path":         dynamicPath,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		// read an access token
		reqPath := fmt.Sprintf("%s/%s", tokenPaths.PathTokenRoleStorage, roleName)
		if dynamicPath && pathExtra != "" {
			reqPath = fmt.Sprintf("%s/%s", reqPath, pathExtra)
		}
		req := &logical.Request{
			Operation: logical.ReadOperation,
			Path:      reqPath,
			Storage:   l,
		}

		resp, err = b.HandleRequest(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Secret)
		require.NoError(t, resp.Error())

		var tokenId = resp.Secret.InternalData["token_id"].(int64)
		var leaseId = resp.Secret.LeaseID
		var secret = resp.Secret

		require.Contains(t, client.accessTokens, fmt.Sprintf("%s_%v", tokenType.String(), tokenId))

		// revoke the access token
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.RevokeOperation,
			Path:      fmt.Sprintf("%s/%s", tokenPaths.PathTokenRoleStorage, leaseId), Storage: l,
			Secret: secret,
		})
		require.NoError(t, err)
		require.Nil(t, resp)

		key := fmt.Sprintf("%s_%v", tokenType.String(), tokenId)
		if gitlabRevokesToken {
			require.Contains(t, client.accessTokens, key)
			// GitLab is the side that revokes; mirror that on the fake so
			// newInMemoryClient's clean-state assertion holds at exit.
			client.ForgetToken(key)
		} else {
			require.NotContains(t, client.accessTokens, key)
		}

		// calling revoke with nil secret
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.RevokeOperation,
			Path:      fmt.Sprintf("%s/%s", tokenPaths.PathTokenRoleStorage, leaseId), Storage: l,
		})
		require.Error(t, err)
		require.Nil(t, resp)

		if !gitlabRevokesToken {
			// calling revoke again would return a token not found in internal error
			switch tokenType {
			case token.TypeProject:
				client.InjectError("RevokeProjectAccessToken")
			case token.TypePersonal:
				client.InjectError("RevokePersonalAccessToken")
			case token.TypeGroup:
				client.InjectError("RevokeGroupAccessToken")
			}
			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.RevokeOperation,
				Path:      fmt.Sprintf("%s/%s", tokenPaths.PathTokenRoleStorage, leaseId), Storage: l,
				Secret: secret,
			})
			require.Error(t, err)
			require.Error(t, resp.Error())
		}

		var expectedEvents = []expectedEvent{
			{eventType: "gitlab/config-write"},
			{eventType: "gitlab/role-write"},
			{eventType: "gitlab/token-write"},
			{eventType: "gitlab/token-revoke"},
		}

		events.expectEvents(t, expectedEvents)
	}

	t.Run("personal access token", func(t *testing.T) {
		generalTokenCreation(t, token.TypePersonal, token.AccessLevelUnknown, false, "admin-user", false, "")
		generalTokenCreation(t, token.TypePersonal, token.AccessLevelUnknown, true, "admin-user", false, "")
	})

	t.Run("personal access token - dynamic path", func(t *testing.T) {
		generalTokenCreation(t, token.TypePersonal, token.AccessLevelUnknown, false, "admin-user", true, "admin-user")
		generalTokenCreation(t, token.TypePersonal, token.AccessLevelUnknown, true, "admin-user", true, "admin-user")
	})

	t.Run("project access token", func(t *testing.T) {
		generalTokenCreation(t, token.TypeProject, token.AccessLevelGuestPermissions, false, "example/example", false, "")
		generalTokenCreation(t, token.TypeProject, token.AccessLevelGuestPermissions, true, "example/example", false, "")
	})

	t.Run("project access token - dynamic path", func(t *testing.T) {
		generalTokenCreation(t, token.TypeProject, token.AccessLevelGuestPermissions, false, "example/.*", true, "example/test")
		generalTokenCreation(t, token.TypeProject, token.AccessLevelGuestPermissions, true, "example/exple", true, "example/exple")
	})

	t.Run("group access token", func(t *testing.T) {
		generalTokenCreation(t, token.TypeGroup, token.AccessLevelGuestPermissions, false, "example", false, "")
		generalTokenCreation(t, token.TypeGroup, token.AccessLevelGuestPermissions, true, "example", false, "")
	})

	t.Run("group access token - dynamic path", func(t *testing.T) {
		generalTokenCreation(t, token.TypeGroup, token.AccessLevelGuestPermissions, false, "ex.*mple", true, "extammmple")
		generalTokenCreation(t, token.TypeGroup, token.AccessLevelGuestPermissions, true, "example", true, "example")
	})

	t.Run("edge cases with dynamic path", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "paths")
		client := newInMemoryClient(t, true)
		t.Cleanup(client.ForgetAllTokens) // this test isn't exercising revocation
		ctx = g.ClientNewContext(ctx, client)
		var b, l, _, err = getBackendWithEvents(ctx)
		require.NoError(t, err)
		require.NoError(t, writeBackendConfig(ctx, b, l, defaultConfig))
		require.NoError(t, err)

		path := "v-.*-end$"
		roleName := "test"
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/%s", backend.PathRoleStorage, roleName), Storage: l,
			Data: map[string]any{
				"path":                 path,
				"name":                 token.TypePersonal.String(),
				"token_type":           token.TypePersonal.String(),
				"access_level":         token.AccessLevelUnknown,
				"ttl":                  "1h",
				"gitlab_revokes_token": false,
				"dynamic_path":         true,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		// valid path
		reqPath := fmt.Sprintf("%s/%s/v-test-end", tokenPaths.PathTokenRoleStorage, roleName)
		req := &logical.Request{
			Operation: logical.ReadOperation,
			Path:      reqPath,
			Storage:   l,
		}
		resp, err = b.HandleRequest(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Secret)
		require.NoError(t, resp.Error())

		// path doesn't match regex
		reqPath = fmt.Sprintf("%s/%s/a-test-end", tokenPaths.PathTokenRoleStorage, roleName)
		req = &logical.Request{
			Operation: logical.ReadOperation,
			Path:      reqPath,
			Storage:   l,
		}
		resp, err = b.HandleRequest(ctx, req)
		require.ErrorIs(t, err, errs.ErrInvalidValue)
		require.NotNil(t, resp)
		require.Nil(t, resp.Secret)
		require.ErrorContains(t, resp.Error(), "path doesn't match regex")

		// invalid path
		reqPath = fmt.Sprintf("%s/%s/test/end", tokenPaths.PathTokenRoleStorage, roleName)
		req = &logical.Request{
			Operation: logical.ReadOperation,
			Path:      reqPath,
			Storage:   l,
		}
		resp, err = b.HandleRequest(ctx, req)
		require.ErrorIs(t, err, errs.ErrInvalidValue)
		require.NotNil(t, resp)
		require.Nil(t, resp.Secret)
		require.ErrorContains(t, resp.Error(), "invalid path")
	})
}
