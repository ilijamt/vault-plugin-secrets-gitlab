//go:build e2e

package integration_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithNormalUser_PersonalAT_Fails(t *testing.T) {
	httpClient, url := getClient(t, "e2e")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)
	var tokenName = "normal_user_initial_token"

	b, l, events, err := getBackendWithEventsAndConfig(ctx, map[string]any{
		"token":              getGitlabToken(tokenName).Token,
		"base_url":           url,
		"auto_rotate_token":  true,
		"auto_rotate_before": "24h",
		"type":               gitlabTypes.TypeSelfManaged.String(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, events)

	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/normal-user", backend.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "normal-user",
				"name":                 token.TypePersonal.String(),
				"token_type":           token.TypePersonal.String(),
				"ttl":                  time.Hour * 120,
				"gitlab_revokes_token": strconv.FormatBool(true),
				"scopes": strings.Join(
					[]string{
						token.ScopeReadApi.String(),
					},
					","),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
	}

	// issue a personal access token
	{
		ctxIssueToken, _ := ctxTestTime(ctx, t, tokenName)
		resp, err := b.HandleRequest(ctxIssueToken, &logical.Request{
			Operation: logical.ReadOperation, Storage: l,
			Path: fmt.Sprintf("%s/normal-user", tokenPaths.PathTokenRoleStorage),
		})

		require.Nil(t, resp)
		require.Error(t, err)
		require.ErrorContains(t, err, "403 Forbidden")
	}

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
	})
}
