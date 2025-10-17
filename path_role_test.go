//go:build unit

package gitlab_test

import (
	"cmp"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestPathRolesList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
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
		"token":    getGitlabToken("admin_user_root").Token,
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		"type":     gitlab2.TypeSelfManaged.String(),
	}

	t.Run("delete non existing role", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
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
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, err = getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, errs.ErrBackendNotConfigured, resp.Error())
	})

	t.Run("access level", func(t *testing.T) {
		t.Run(token.TypePersonal.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t, "unit")
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 token.TypePersonal.String(),
						"token_type":           token.TypePersonal.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               token.ValidPersonalTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NoError(t, resp.Error())
				require.Empty(t, resp.Warnings)
			})
			t.Run("with access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t, "unit")
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 token.TypePersonal.String(),
						"access_level":         token.AccessLevelOwnerPermissions.String(),
						"token_type":           token.TypePersonal.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               token.ValidPersonalTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
		})

		t.Run(token.TypeProject.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t, "unit")
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 token.TypeProject.String(),
						"token_type":           token.TypeProject.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               token.ValidProjectTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
			t.Run("with access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t, "unit")
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 token.TypeProject.String(),
						"access_level":         token.AccessLevelOwnerPermissions.String(),
						"token_type":           token.TypeProject.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               token.ValidProjectTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NoError(t, resp.Error())
				require.Empty(t, resp.Warnings)
			})
		})

		t.Run(token.TypeGroup.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t, "unit")
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 token.TypeGroup.String(),
						"token_type":           token.TypeGroup.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               token.ValidGroupTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
			t.Run("with access level defined", func(t *testing.T) {
				ctx := getCtxGitlabClient(t, "unit")
				var b, l, err = getBackendWithConfig(ctx, defaultConfig)
				require.NoError(t, err)
				resp, err := b.HandleRequest(ctx, &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":                 "user",
						"name":                 token.TypeGroup.String(),
						"access_level":         token.AccessLevelOwnerPermissions.String(),
						"token_type":           token.TypeGroup.String(),
						"ttl":                  gitlab.DefaultAccessTokenMinTTL,
						"scopes":               token.ValidGroupTokenScopes,
						"gitlab_revokes_token": false,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NoError(t, resp.Error())
				require.Empty(t, resp.Warnings)
				require.EqualValues(t, resp.Data["config_name"], gitlab.TypeConfigDefault)
			})
		})

	})

	t.Run("create with missing parameters", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
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
		assert.EqualValues(t, 4, errorMap[errs.ErrFieldRequired.Error()])
		assert.EqualValues(t, 2, errorMap[errs.ErrFieldInvalidValue.Error()])
	})

	t.Run("invalid name template", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, err = getBackendWithConfig(ctx, defaultConfig)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: map[string]any{
				"path":                 "user",
				"name":                 "{{ . } invalid template",
				"token_type":           token.TypePersonal.String(),
				"ttl":                  gitlab.DefaultAccessTokenMinTTL,
				"scopes":               token.ValidPersonalTokenScopes,
				"gitlab_revokes_token": false,
			},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.ErrorContains(t, resp.Error(), "invalid template")
	})

	t.Run("Project token scopes", func(t *testing.T) {
		t.Run("valid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"access_level": token.AccessLevelOwnerPermissions.String(),
					"ttl":          "48h",
					"token_type":   token.TypeProject.String(),
					"scopes":       token.ValidProjectTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.EqualValues(t, resp.Data["config_name"], gitlab.TypeConfigDefault)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":                 "user",
					"name":                 "Example project personal token",
					"access_level":         token.AccessLevelOwnerPermissions.String(),
					"token_type":           token.TypeProject.String(),
					"ttl":                  "48h",
					"scopes":               token.ValidPersonalTokenScopes,
					"gitlab_revokes_token": false,
				},
			})
			require.Error(t, err)
			require.NotNil(t, resp)
			var errorMap = countErrByName(err.(*multierror.Error))
			assert.EqualValues(t, 1, errorMap[errs.ErrFieldInvalidValue.Error()])
		})
	})

	t.Run("Personal token scopes", func(t *testing.T) {
		t.Run("valid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":       "user",
					"name":       "Example user personal token",
					"ttl":        "48h",
					"token_type": token.TypePersonal.String(),
					"scopes":     token.ValidPersonalTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.EqualValues(t, resp.Data["config_name"], gitlab.TypeConfigDefault)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":       "user",
					"name":       "Example user personal token",
					"token_type": token.TypePersonal.String(),
					"scopes": strings.Join([]string{
						"invalid_scope",
					}, ", "),
				},
			})
			require.Error(t, err)
			require.NotNil(t, resp)
			var errorMap = countErrByName(err.(*multierror.Error))
			assert.EqualValues(t, 1, errorMap[errs.ErrFieldInvalidValue.Error()])
		})
	})

	t.Run("Group token scopes", func(t *testing.T) {
		t.Run("valid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"ttl":          "48h",
					"access_level": token.AccessLevelOwnerPermissions.String(),
					"token_type":   token.TypeGroup.String(),
					"scopes":       token.ValidProjectTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.EqualValues(t, resp.Data["config_name"], gitlab.TypeConfigDefault)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"access_level": token.AccessLevelOwnerPermissions.String(),
					"token_type":   token.TypeGroup.String(),
					"scopes":       token.ValidPersonalTokenScopes,
				},
			})
			require.Error(t, err)
			require.NotNil(t, resp)
			var errorMap = countErrByName(err.(*multierror.Error))
			assert.EqualValues(t, 1, errorMap[errs.ErrFieldInvalidValue.Error()])
		})
	})

	t.Run("update handler existence check", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
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
		ctx := getCtxGitlabClient(t, "unit")
		var b, l, events, err = getBackendWithEvents(ctx)
		require.NoError(t, err)

		var defaultConfig = map[string]any{
			"token":    getGitlabToken("admin_user_root").Token,
			"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
			"type":     gitlab2.TypeSelfManaged.String(),
		}

		// create a configuration with max ttl set to 10 days
		func() {
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.UpdateOperation,
				Path:      fmt.Sprintf("%s/%s", gitlab.PathConfigStorage, gitlab.DefaultConfigName), Storage: l,
				Data: defaultConfig,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		}()

		var roleData = map[string]any{
			"path":                 "user",
			"name":                 "Example user personal token",
			"token_type":           token.TypePersonal.String(),
			"ttl":                  int64((5 * 24 * time.Hour).Seconds()),
			"gitlab_revokes_token": false,
			"scopes": strings.Join([]string{
				token.ScopeApi.String(),
				token.ScopeReadRegistry.String(),
			}, ", "),
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
		require.EqualValues(t, resp.Data["config_name"], gitlab.TypeConfigDefault)

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
