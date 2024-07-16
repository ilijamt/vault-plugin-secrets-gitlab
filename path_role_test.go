package gitlab_test

import (
	"cmp"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathRolesList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ListOperation,
			Path:      gitlab.PathRoleStorage, Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		assert.Empty(t, resp.Data)
	})
}

func TestPathRoles(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    "glpat-secret-random-token",
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
	}

	t.Run("delete non existing role", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("we get error if backend is not set up during role write", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, gitlab.ErrBackendNotConfigured, resp.Error())
	})

	t.Run("access level", func(t *testing.T) {
		t.Run(gitlab.TokenTypePersonal.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t)
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 gitlab.TokenTypePersonal.String(),
						"token_type":           gitlab.TokenTypePersonal.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               gitlab.ValidPersonalTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NoError(t, resp.Error())
				require.Empty(t, resp.Warnings)
			})
			t.Run("with access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t)
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 gitlab.TokenTypePersonal.String(),
						"access_level":         gitlab.AccessLevelOwnerPermissions.String(),
						"token_type":           gitlab.TokenTypePersonal.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               gitlab.ValidPersonalTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
		})

		t.Run(gitlab.TokenTypeProject.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t)
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 gitlab.TokenTypeProject.String(),
						"token_type":           gitlab.TokenTypeProject.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               gitlab.ValidProjectTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
			t.Run("with access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t)
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 gitlab.TokenTypeProject.String(),
						"access_level":         gitlab.AccessLevelOwnerPermissions.String(),
						"token_type":           gitlab.TokenTypeProject.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               gitlab.ValidProjectTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NoError(t, resp.Error())
				require.Empty(t, resp.Warnings)
			})
		})

		t.Run(gitlab.TokenTypeGroup.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t)
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 gitlab.TokenTypeGroup.String(),
						"token_type":           gitlab.TokenTypeGroup.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               gitlab.ValidGroupTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
			t.Run("with access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t)
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 gitlab.TokenTypeGroup.String(),
						"access_level":         gitlab.AccessLevelOwnerPermissions.String(),
						"token_type":           gitlab.TokenTypeGroup.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               gitlab.ValidGroupTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NoError(t, resp.Error())
				require.Empty(t, resp.Warnings)
			})
		})

	})

	t.Run("create with missing parameters", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		var b, l, err = getBackendWithConfig(ctx, defaultConfig)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: map[string]any{},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		var errorMap = countErrByName(err.(*multierror.Error))
		assert.EqualValues(t, 4, errorMap[gitlab.ErrFieldRequired.Error()])
		assert.EqualValues(t, 2, errorMap[gitlab.ErrFieldInvalidValue.Error()])
	})

	t.Run("Project token scopes", func(t *testing.T) {
		t.Run("valid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t)
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"access_level": gitlab.AccessLevelOwnerPermissions.String(),
					"ttl":          "48h",
					"token_type":   gitlab.TokenTypeProject.String(),
					"scopes":       gitlab.ValidProjectTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t)
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":                 "user",
					"name":                 "Example project personal token",
					"access_level":         gitlab.AccessLevelOwnerPermissions.String(),
					"token_type":           gitlab.TokenTypeProject.String(),
					"ttl":                  "48h",
					"scopes":               gitlab.ValidPersonalTokenScopes,
					"gitlab_revokes_token": false,
				},
			})
			require.Error(t, err)
			require.NotNil(t, resp)
			var errorMap = countErrByName(err.(*multierror.Error))
			assert.EqualValues(t, 1, errorMap[gitlab.ErrFieldInvalidValue.Error()])
		})
	})

	t.Run("Personal token scopes", func(t *testing.T) {
		t.Run("valid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t)
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":       "user",
					"name":       "Example user personal token",
					"ttl":        "48h",
					"token_type": gitlab.TokenTypePersonal.String(),
					"scopes":     gitlab.ValidPersonalTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t)
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":       "user",
					"name":       "Example user personal token",
					"token_type": gitlab.TokenTypePersonal.String(),
					"scopes": []string{
						"invalid_scope",
					},
				},
			})
			require.Error(t, err)
			require.NotNil(t, resp)
			var errorMap = countErrByName(err.(*multierror.Error))
			assert.EqualValues(t, 1, errorMap[gitlab.ErrFieldInvalidValue.Error()])
		})
	})

	t.Run("Group token scopes", func(t *testing.T) {
		t.Run("valid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t)
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"ttl":          "48h",
					"access_level": gitlab.AccessLevelOwnerPermissions.String(),
					"token_type":   gitlab.TokenTypeGroup.String(),
					"scopes":       gitlab.ValidProjectTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t)
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"access_level": gitlab.AccessLevelOwnerPermissions.String(),
					"token_type":   gitlab.TokenTypeGroup.String(),
					"scopes":       gitlab.ValidPersonalTokenScopes,
				},
			})
			require.Error(t, err)
			require.NotNil(t, resp)
			var errorMap = countErrByName(err.(*multierror.Error))
			assert.EqualValues(t, 1, errorMap[gitlab.ErrFieldInvalidValue.Error()])
		})
	})

	t.Run("update handler existence check", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		hasExistenceCheck, exists, err := b.HandleExistenceCheck(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})

		require.True(t, hasExistenceCheck)
		require.False(t, exists)
		require.NoError(t, err)
	})

	t.Run("full flow check roles", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		var b, l, events, err = getBackendWithEvents(ctx)
		require.NoError(t, err)

		var defaultConfig = map[string]any{
			"token":    "glpat-secret-random-token",
			"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		}

		// create a configuration with max ttl set to 10 days
		func() {
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.UpdateOperation,
				Path:      gitlab.PathConfigStorage, Storage: l,
				Data: defaultConfig,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		}()

		var roleData = map[string]any{
			"path":                 "user",
			"name":                 "Example user personal token",
			"token_type":           gitlab.TokenTypePersonal.String(),
			"ttl":                  int64((5 * 24 * time.Hour).Seconds()),
			"gitlab_revokes_token": false,
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
		}

		// create a role
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.Empty(t, resp.Warnings)

		// read a role
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.EqualValues(t, "test", resp.Data["role_name"])
		require.Equal(t, int64((5 * 24 * time.Hour).Seconds()), resp.Data["ttl"])
		assert.Subset(t, resp.Data, roleData)

		// update a role
		roleData["name"] = "Example user personal token - updated"
		roleData["path"] = "user2"
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		assert.Subset(t, resp.Data, roleData)
		require.EqualValues(t, "test", resp.Data["role_name"])

		// read a role
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		assert.Subset(t, resp.Data, roleData)
		require.EqualValues(t, "test", resp.Data["role_name"])
		require.EqualValues(t, "user2", resp.Data["path"])
		require.EqualValues(t, "Example user personal token - updated", resp.Data["name"])

		// delete a role
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)

		// read a role
		resp, err = b.HandleRequest(ctx, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)

		// check the events
		require.NotEmpty(t, events)

		events.expectEvents(t, []expectedEvent{
			{eventType: "gitlab/config-write"},
			{eventType: "gitlab/role-write"},
			{eventType: "gitlab/role-write"},
			{eventType: "gitlab/role-delete"},
		})

	})

}
