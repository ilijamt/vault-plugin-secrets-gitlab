package gitlab_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestWithNormalUser_PersonalAT_Fails(t *testing.T) {
	httpClient, url := getClient(t)
	ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

	b, l, events, err := getBackendWithEvents(ctx)
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigStorage, Storage: l,
		Data: map[string]any{
			"token":              "glpat-secret-normal-token",
			"base_url":           url,
			"auto_rotate_token":  true,
			"auto_rotate_before": "24h",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.NotEmpty(t, events)

	{
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/normal-user", gitlab.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "normal-user",
				"name":                 gitlab.TokenTypePersonal.String(),
				"token_type":           gitlab.TokenTypePersonal.String(),
				"ttl":                  time.Hour * 120,
				"gitlab_revokes_token": strconv.FormatBool(true),
				"scopes": strings.Join(
					[]string{
						gitlab.TokenScopeReadApi.String(),
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
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation, Storage: l,
			Path: fmt.Sprintf("%s/normal-user", gitlab.PathTokenRoleStorage),
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
