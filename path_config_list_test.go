//go:build unit

package gitlab_test

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

func TestPathConfigList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ListOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		assert.Empty(t, resp.Data)
	})

	t.Run("multiple configs", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, events, err = getBackendWithEventsAndConfigName(ctx,
			map[string]any{
				"token":    getGitlabToken("admin_user_root").Token,
				"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
				"type":     gitlab2.TypeSaaS.String(),
			},
			gitlab.DefaultConfigName,
		)
		require.NoError(t, err)
		require.NotNil(t, events)
		require.NotNil(t, b)
		require.NotNil(t, l)

		require.NoError(t,
			writeBackendConfigWithName(ctx, b, l,
				map[string]any{
					"token":    getGitlabToken("admin_user_initial_token").Token,
					"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
					"type":     gitlab2.TypeSelfManaged.String(),
				},
				"admin",
			),
		)

		require.NoError(t,
			writeBackendConfigWithName(ctx, b, l,
				map[string]any{
					"token":    getGitlabToken("normal_user_initial_token").Token,
					"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
					"type":     gitlab2.TypeDedicated.String(),
				},
				"normal",
			),
		)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ListOperation,
			Path:      gitlab.PathConfigStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.NotNil(t, resp.Data["keys"])
		keysResponse := resp.Data["keys"].([]string)
		slices.Sort(keysResponse)
		keysExpected := []string{gitlab.DefaultConfigName, "admin", "normal"}
		slices.Sort(keysExpected)
		require.EqualValues(t, keysExpected, keysResponse)
		require.Len(t, keysResponse, 3)

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
			{eventType: "gitlab/config-write"},
			{eventType: "gitlab/config-write"},
		})

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Data)
		require.EqualValues(t, gitlab2.TypeSaaS.String(), resp.Data["type"])

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/normal", gitlab.PathConfigStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Data)
		require.EqualValues(t, gitlab2.TypeDedicated.String(), resp.Data["type"])

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/admin", gitlab.PathConfigStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Data)
		require.EqualValues(t, gitlab2.TypeSelfManaged.String(), resp.Data["type"])
	})
}
