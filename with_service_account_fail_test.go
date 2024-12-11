package gitlab_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestWithServiceAccountUserFail(t *testing.T) {
	for _, typ := range []gitlab.Type{
		gitlab.TypeSaaS,
		gitlab.TypeDedicated,
	} {
		t.Run(typ.String(), func(t *testing.T) {
			httpClient, _ := getClient(t)
			ctx := gitlab.HttpClientNewContext(context.Background(), httpClient)

			b, l, events, err := getBackendWithEvents(ctx)
			require.NoError(t, err)

			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.UpdateOperation,
				Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
				Data: map[string]any{
					"token":              gitlabServiceAccountToken,
					"base_url":           gitlabServiceAccountUrl,
					"auto_rotate_token":  true,
					"auto_rotate_before": "24h",
					"type":               typ.String(),
				},
			})

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.NotEmpty(t, events)

			require.NotNil(t, b.GetClient(gitlab.DefaultConfigName))
			var gClient = b.GetClient(gitlab.DefaultConfigName).GitlabClient(ctx)
			require.NotNil(t, gClient)

			usr, _, err := gClient.Users.CreateServiceAccountUser(&g.CreateServiceAccountUserOptions{})
			require.NoError(t, err)
			require.NotNil(t, usr)

			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/user-service-account", gitlab.PathRoleStorage), Storage: l,
				Data: map[string]any{
					"path":                 usr.Username,
					"name":                 fmt.Sprintf(`user-service-account-%s`, usr.Username),
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
		})
	}

}
