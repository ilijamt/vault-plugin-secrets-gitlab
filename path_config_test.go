//go:build unit

package gitlab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestPathConfig(t *testing.T) {
	t.Run("initial config should be empty fail with backend not configured", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, resp.Error(), errs.ErrBackendNotConfigured)
	})

	t.Run("deleting uninitialized config should fail with backend not configured", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		b, l, err := getBackend(ctx)
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.True(t, resp.IsError())
		require.EqualValues(t, resp.Error(), errs.ErrBackendNotConfigured)
	})

	t.Run("write, read, delete and read config", func(t *testing.T) {
		httpClient, url := getClient(t, "unit")
		ctx := utils.HttpClientNewContext(t.Context(), httpClient)

		b, l, events, err := getBackendWithEvents(ctx)
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":    getGitlabToken("admin_user_root").Token,
				"base_url": url,
				"type":     gitlab2.TypeSelfManaged.String(),
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		assert.NotEmpty(t, resp.Data["token_sha1_hash"])
		assert.NotEmpty(t, resp.Data["base_url"])
		require.Len(t, events.eventsProcessed, 1)
		require.Empty(t, resp.Data["token"])

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
		require.Len(t, events.eventsProcessed, 2)

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
			{eventType: "gitlab/config-delete"},
		})
	})

	t.Run("write, read, delete and read config with show config token", func(t *testing.T) {
		httpClient, url := getClient(t, "unit")
		ctx := utils.HttpClientNewContext(t.Context(), httpClient)

		b, l, events, err := getBackendWithFlagsWithEvents(ctx, flags.Flags{ShowConfigToken: true})
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":    getGitlabToken("admin_user_root").Token,
				"base_url": url,
				"type":     gitlab2.TypeSelfManaged.String(),
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		assert.NotEmpty(t, resp.Data["token_sha1_hash"])
		assert.NotEmpty(t, resp.Data["base_url"])
		require.Len(t, events.eventsProcessed, 1)
		require.NotEmpty(t, resp.Data["token"])

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
		require.Len(t, events.eventsProcessed, 2)

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
			{eventType: "gitlab/config-delete"},
		})
	})
	t.Run("invalid token", func(t *testing.T) {
		httpClient, url := getClient(t, "unit")
		ctx := utils.HttpClientNewContext(t.Context(), httpClient)

		b, l, events, err := getBackendWithEvents(ctx)
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":    "invalid-token",
				"base_url": url,
				"type":     gitlab2.TypeSelfManaged.String(),
			},
		})

		require.Error(t, err)
		require.Nil(t, resp)

		events.expectEvents(t, []expectedEvent{})
	})

	t.Run("missing token from the request", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		b, l, err := getBackend(ctx)
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{},
		})

		require.Error(t, err)
		require.Nil(t, resp)

		var errorMap = utils.CountErrByName(err.(*multierror.Error))
		assert.EqualValues(t, 3, errorMap[errs.ErrFieldRequired.Error()])
		require.Len(t, errorMap, 1)
	})

	t.Run("patch a config with no storage", func(t *testing.T) {
		httpClient, url := getClient(t, "unit")
		ctx := utils.HttpClientNewContext(t.Context(), httpClient)

		b, _, err := getBackend(ctx)
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.PatchOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: nil,
			Data: map[string]any{
				"token":    getGitlabToken("admin_user_root").Token,
				"base_url": url,
				"type":     gitlab2.TypeSelfManaged.String(),
			},
		})

		require.ErrorIs(t, err, errs.ErrNilValue)
		require.Nil(t, resp)
	})

	t.Run("patch a config no backend", func(t *testing.T) {
		httpClient, url := getClient(t, "unit")
		ctx := utils.HttpClientNewContext(t.Context(), httpClient)

		b, l, err := getBackend(ctx)
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.PatchOperation,
			Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
			Data: map[string]any{
				"token":    getGitlabToken("admin_user_root").Token,
				"base_url": url,
				"type":     gitlab2.TypeSelfManaged.String(),
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.EqualValues(t, resp.Error(), errs.ErrBackendNotConfigured)
	})

	t.Run("patch a config", func(t *testing.T) {
		httpClient, url := getClient(t, "unit")
		ctx := utils.HttpClientNewContext(t.Context(), httpClient)
		var path = fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName)

		b, l, events, err := getBackendWithEvents(ctx)
		require.NoError(t, err)

		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      path, Storage: l,
			Data: map[string]any{
				"token":    getGitlabToken("admin_user_root").Token,
				"base_url": url,
				"type":     gitlab2.TypeSelfManaged.String(),
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      path, Storage: l,
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		tokenOriginalSha1Hash := resp.Data["token_sha1_hash"].(string)
		require.NotEmpty(t, tokenOriginalSha1Hash)
		require.Equal(t, gitlab2.TypeSelfManaged.String(), resp.Data["type"])
		require.NotNil(t, b.GetClient(gitlab.DefaultConfigName).GitlabClient(ctx))

		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.PatchOperation,
			Path:      path, Storage: l,
			Data: map[string]interface{}{
				"type":  gitlab2.TypeSaaS.String(),
				"token": getGitlabToken("admin_user_initial_token").Token,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		tokenNewSha1Hash := resp.Data["token_sha1_hash"].(string)
		require.NotEmpty(t, tokenNewSha1Hash)
		require.NotEqual(t, tokenOriginalSha1Hash, tokenNewSha1Hash)

		require.Equal(t, gitlab2.TypeSaaS.String(), resp.Data["type"])
		require.NotNil(t, b.GetClient(gitlab.DefaultConfigName).GitlabClient(ctx))

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
			{eventType: "gitlab/config-patch"},
		})

	})

}
