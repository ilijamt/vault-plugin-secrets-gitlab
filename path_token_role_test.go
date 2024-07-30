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
)

func TestPathTokenRoles(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    "glpat-secret-random-token",
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
	}

	t.Run("role not found", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
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

	var generalTokenCreation = func(t *testing.T, tokenType gitlab.TokenType, level gitlab.AccessLevel, gitlabRevokesToken bool) {
		t.Logf("token creation, token type: %s, level: %s, gitlab revokes token: %t", tokenType, level, gitlabRevokesToken)
		ctx := getCtxGitlabClient(t)
		client := newInMemoryClient(true)
		ctx = gitlab.GitlabClientNewContext(ctx, client)
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

		var path string
		switch tokenType {
		case gitlab.TokenTypeProject:
			path = "example/example"
		case gitlab.TokenTypePersonal:
			path = "admin-user"
		case gitlab.TokenTypeGroup:
			path = "example"
		case gitlab.TokenTypeServiceAccount:
			path = "example/example-service-account"
		}

		// create a role
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 path,
				"name":                 tokenType.String(),
				"token_type":           tokenType.String(),
				"access_level":         level,
				"ttl":                  ttl,
				"gitlab_revokes_token": strconv.FormatBool(gitlabRevokesToken),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		// read an access token
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathTokenRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Secret)
		require.NoError(t, resp.Error())

		var tokenId = resp.Secret.InternalData["token_id"].(int)
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
			case gitlab.TokenTypeProject:
				client.projectAccessTokenRevokeError = true
			case gitlab.TokenTypePersonal:
				client.personalAccessTokenRevokeError = true
			case gitlab.TokenTypeGroup:
				client.groupAccessTokenRevokeError = true
			case gitlab.TokenTypeServiceAccount:
				client.serviceAccountPersonalAccessTokenRevokeError = true
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
		generalTokenCreation(t, gitlab.TokenTypePersonal, gitlab.AccessLevelUnknown, false)
		generalTokenCreation(t, gitlab.TokenTypePersonal, gitlab.AccessLevelUnknown, true)
	})

	t.Run("project access token", func(t *testing.T) {
		generalTokenCreation(t, gitlab.TokenTypeProject, gitlab.AccessLevelGuestPermissions, false)
		generalTokenCreation(t, gitlab.TokenTypeProject, gitlab.AccessLevelGuestPermissions, true)
	})

	t.Run("group access token", func(t *testing.T) {
		generalTokenCreation(t, gitlab.TokenTypeGroup, gitlab.AccessLevelGuestPermissions, false)
		generalTokenCreation(t, gitlab.TokenTypeGroup, gitlab.AccessLevelGuestPermissions, true)
	})

	t.Run("service account personal access token", func(t *testing.T) {
		generalTokenCreation(t, gitlab.TokenTypeServiceAccount, gitlab.AccessLevelUnknown, false)
		generalTokenCreation(t, gitlab.TokenTypeServiceAccount, gitlab.AccessLevelUnknown, true)
	})
}
