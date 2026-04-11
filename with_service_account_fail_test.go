//go:build selfhosted

package gitlab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	glab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithServiceAccountUserFail(t *testing.T) {
	for _, typ := range []gitlabTypes.Type{
		gitlabTypes.TypeSaaS,
		gitlabTypes.TypeDedicated,
	} {
		t.Run(typ.String(), func(t *testing.T) {
			httpClient, _ := getClient(t, "selfhosted")
			ctx := utils.HttpClientNewContext(t.Context(), httpClient)

			b, l, events, err := getBackendWithEvents(ctx)
			require.NoError(t, err)

			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.UpdateOperation,
				Path:      fmt.Sprintf("%s/%s", backend.PathConfigStorage, backend.DefaultConfigName), Storage: l,
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

			require.Nil(t, b.GetClient(backend.DefaultConfigName))
			var client glab.Client
			client, err = b.GetClientByName(ctx, l, backend.DefaultConfigName)
			require.NoError(t, err)
			require.NotNil(t, client)
			var gClient = client.GitlabClient(ctx)
			require.NotNil(t, gClient)

			usr, _, err := gClient.Users.CreateServiceAccountUser(&g.CreateServiceAccountUserOptions{})
			require.NoError(t, err)
			require.NotNil(t, usr)

			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/user-service-account", backend.PathRoleStorage), Storage: l,
				Data: map[string]any{
					"path":                 usr.Username,
					"name":                 fmt.Sprintf(`user-service-account-%s`, usr.Username),
					"token_type":           token.TypeUserServiceAccount.String(),
					"ttl":                  backend.DefaultAccessTokenMinTTL,
					"scopes":               token.ValidUserServiceAccountTokenScopes,
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
