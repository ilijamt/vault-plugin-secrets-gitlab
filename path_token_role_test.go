//go:build unit

package gitlab_test

import (
	"cmp"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	g "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestPathTokenRoles(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    getGitlabToken("admin_user_root").Token,
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		"type":     g.TypeSelfManaged.String(),
	}

	t.Run("role not found", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathTokenRoleStorage), Storage: l,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.ErrorIs(t, err, gitlab.ErrRoleNotFound)
	})

	var generalTokenCreation = func(t *testing.T, tokenType token.Type, level token.AccessLevel, gitlabRevokesToken bool, path string, dynamicPath bool) {
		t.Logf("token creation, token type: %s, level: %s, gitlab revokes token: %t, path: %s", tokenType, level, gitlabRevokesToken, path)
		ctx := getCtxGitlabClient(t, "unit")
		client := newInMemoryClient(true)
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
			Path:      fmt.Sprintf("%s/%s", gitlab.PathRoleStorage, roleName), Storage: l,
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
		reqPath := fmt.Sprintf("%s/%s", gitlab.PathTokenRoleStorage, roleName)
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
			Path:      fmt.Sprintf("%s/%s", gitlab.PathTokenRoleStorage, leaseId), Storage: l,
			Secret: secret,
		})
		require.NoError(t, err)
		require.Nil(t, resp)

		if gitlabRevokesToken {
			require.Contains(t, client.accessTokens, fmt.Sprintf("%s_%v", tokenType.String(), tokenId))
		} else {
			require.NotContains(t, client.accessTokens, fmt.Sprintf("%s_%v", tokenType.String(), tokenId))
		}

		// calling revoke with nil secret
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.RevokeOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathTokenRoleStorage, leaseId), Storage: l,
		})
		require.Error(t, err)
		require.Nil(t, resp)

		if !gitlabRevokesToken {
			// calling revoke again would return a token not found in internal error
			switch tokenType {
			case token.TypeProject:
				client.projectAccessTokenRevokeError = true
			case token.TypePersonal:
				client.personalAccessTokenRevokeError = true
			case token.TypeGroup:
				client.groupAccessTokenRevokeError = true
			}
			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.RevokeOperation,
				Path:      fmt.Sprintf("%s/%s", gitlab.PathTokenRoleStorage, leaseId), Storage: l,
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
		generalTokenCreation(t, token.TypePersonal, token.AccessLevelUnknown, false, "admin-user", false)
		generalTokenCreation(t, token.TypePersonal, token.AccessLevelUnknown, true, "admin-user", false)
		// generalTokenCreation(t, token.TypeProject, token.AccessLevelGuestPermissions, false, "*", "some-user")
	})

	t.Run("project access token", func(t *testing.T) {
		generalTokenCreation(t, token.TypeProject, token.AccessLevelGuestPermissions, false, "example/example", false)
		generalTokenCreation(t, token.TypeProject, token.AccessLevelGuestPermissions, true, "example/example", false)
	})

	t.Run("group access token", func(t *testing.T) {
		generalTokenCreation(t, token.TypeGroup, token.AccessLevelGuestPermissions, false, "example", false)
		generalTokenCreation(t, token.TypeGroup, token.AccessLevelGuestPermissions, true, "example", false)
	})
}
