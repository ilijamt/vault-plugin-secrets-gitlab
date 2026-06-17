//go:build e2e

package integration_test

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	token2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithNormalUser_GAT(t *testing.T) {
	httpClient, url := getClient(t, "e2e")
	ctx := utils.HttpClientNewContext(t.Context(), httpClient)

	b, l, events, err := getBackendWithEventsAndConfig(ctx,
		standardConfig(gitlabTypes.TypeSelfManaged, url, getGitlabToken("normal_user_initial_token").Token))
	require.NoError(t, err)

	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/gat", backend.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "example",
				"name":                 `gat-token`,
				"token_type":           token2.TypeGroup.String(),
				"ttl":                  time.Hour * 120,
				"gitlab_revokes_token": strconv.FormatBool(false),
				"access_level":         token2.AccessLevelMaintainerPermissions.String(),
				"scopes":               strings.Join([]string{token2.ScopeReadApi.String()}, ","),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
	}

	token, secret := issueToken(ctx, t, b, l, "e2e", "gat")
	requireTokenStatus(t, httpClient, url, token, http.StatusOK)
	revokeSecret(ctx, t, b, l, secret)
	requireTokenStatus(t, httpClient, url, token, http.StatusUnauthorized)

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/config-write"},
		{eventType: "gitlab/role-write"},
		{eventType: "gitlab/token-write"},
		{eventType: "gitlab/token-revoke"},
	})
}
