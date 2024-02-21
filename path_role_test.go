package gitlab_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/logical"
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPathRolesList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		var b, l, err = getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
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
	var defaultConfig = map[string]any{"token": "random-token"}
	t.Run("delete non existing role", func(t *testing.T) {
		var b, l, err = getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("we get error if Backend is not set up during role write", func(t *testing.T) {
		var b, l, err = getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, gitlab.ErrBackendNotConfigured, resp.Error())
	})

	t.Run("access level", func(t *testing.T) {
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)

		t.Run(gitlab.TokenTypePersonal.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				resp, err := b.HandleRequest(context.Background(), &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":       "user",
						"name":       gitlab.TokenTypePersonal.String(),
						"token_type": gitlab.TokenTypePersonal.String(),
						"token_ttl":  gitlab.DefaultAccessTokenMinTTL,
						"scopes":     gitlab.ValidPersonalTokenScopes,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NoError(t, resp.Error())
				require.Empty(t, resp.Warnings)
			})
			t.Run("with access level defined", func(t *testing.T) {
				resp, err := b.HandleRequest(context.Background(), &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":         "user",
						"name":         gitlab.TokenTypePersonal.String(),
						"access_level": gitlab.AccessLevelOwnerPermissions.String(),
						"token_type":   gitlab.TokenTypePersonal.String(),
						"token_ttl":    gitlab.DefaultAccessTokenMinTTL,
						"scopes":       gitlab.ValidPersonalTokenScopes,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
		})

		t.Run(gitlab.TokenTypeProject.String(), func(t *testing.T) {
			t.Run("no access level defined", func(t *testing.T) {
				resp, err := b.HandleRequest(context.Background(), &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":       "user",
						"name":       gitlab.TokenTypeProject.String(),
						"token_type": gitlab.TokenTypeProject.String(),
						"token_ttl":  gitlab.DefaultAccessTokenMinTTL,
						"scopes":     gitlab.ValidProjectTokenScopes,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
			t.Run("with access level defined", func(t *testing.T) {
				resp, err := b.HandleRequest(context.Background(), &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":         "user",
						"name":         gitlab.TokenTypeProject.String(),
						"access_level": gitlab.AccessLevelOwnerPermissions.String(),
						"token_type":   gitlab.TokenTypeProject.String(),
						"token_ttl":    gitlab.DefaultAccessTokenMinTTL,
						"scopes":       gitlab.ValidProjectTokenScopes,
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
				resp, err := b.HandleRequest(context.Background(), &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":       "user",
						"name":       gitlab.TokenTypeGroup.String(),
						"token_type": gitlab.TokenTypeGroup.String(),
						"token_ttl":  gitlab.DefaultAccessTokenMinTTL,
						"scopes":     gitlab.ValidGroupTokenScopes,
					},
				})
				require.Error(t, err)
				require.NotNil(t, resp)
				require.Error(t, resp.Error())
			})
			t.Run("with access level defined", func(t *testing.T) {
				resp, err := b.HandleRequest(context.Background(), &logical.Request{
					Operation: logical.CreateOperation,
					Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
					Data: map[string]any{
						"path":         "user",
						"name":         gitlab.TokenTypeGroup.String(),
						"access_level": gitlab.AccessLevelOwnerPermissions.String(),
						"token_type":   gitlab.TokenTypeGroup.String(),
						"token_ttl":    gitlab.DefaultAccessTokenMinTTL,
						"scopes":       gitlab.ValidGroupTokenScopes,
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
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: map[string]any{},
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		var errorMap = countErrByName(err.(*multierror.Error))
		assert.EqualValues(t, 3, errorMap[gitlab.ErrFieldRequired.Error()])
		assert.EqualValues(t, 2, errorMap[gitlab.ErrFieldInvalidValue.Error()])
	})

	t.Run("Project token scopes", func(t *testing.T) {
		t.Run("valid scopes", func(t *testing.T) {
			var b, l, err = getBackendWithConfig(defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"access_level": gitlab.AccessLevelOwnerPermissions.String(),
					"token_type":   gitlab.TokenTypeProject.String(),
					"scopes":       gitlab.ValidProjectTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			var b, l, err = getBackendWithConfig(defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example project personal token",
					"access_level": gitlab.AccessLevelOwnerPermissions.String(),
					"token_type":   gitlab.TokenTypeProject.String(),
					"scopes":       gitlab.ValidPersonalTokenScopes,
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
			var b, l, err = getBackendWithConfig(defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":       "user",
					"name":       "Example user personal token",
					"token_type": gitlab.TokenTypePersonal.String(),
					"scopes":     gitlab.ValidPersonalTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			var b, l, err = getBackendWithConfig(defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
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
			var b, l, err = getBackendWithConfig(defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/%d", gitlab.PathRoleStorage, time.Now().UnixNano()), Storage: l,
				Data: map[string]any{
					"path":         "user",
					"name":         "Example user personal token",
					"access_level": gitlab.AccessLevelOwnerPermissions.String(),
					"token_type":   gitlab.TokenTypeGroup.String(),
					"scopes":       gitlab.ValidProjectTokenScopes,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})

		t.Run("invalid scopes", func(t *testing.T) {
			var b, l, err = getBackendWithConfig(defaultConfig)
			require.NoError(t, err)
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
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

	t.Run("24h > TokenTTL > MaxTTL (10 days)", func(t *testing.T) {
		var b, l, err = getBackend()
		require.NoError(t, err)

		// create a configuration with max ttl set to 10 days
		func() {
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.UpdateOperation,
				Path:      gitlab.PathConfigStorage, Storage: l,
				Data: map[string]any{
					"max_ttl": (10 * 24 * time.Hour).Seconds(),
					"token":   "token",
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		}()

		var roleData = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"token_ttl":  int64((12 * 24 * time.Hour).Seconds()),
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
		}

		// create a role
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)

		// read a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.EqualValues(t, (10 * 24 * time.Hour).Seconds(), resp.Data["token_ttl"])
	})

	t.Run("0 > TokenTTL > 24h", func(t *testing.T) {
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)

		var roleData = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"token_ttl":  (12 * time.Hour).Seconds(),
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
		}

		// create a role
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)

		// read a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.EqualValues(t, (24 * time.Hour).Seconds(), resp.Data["token_ttl"])
	})

	t.Run("not set token_ttl should default to 24h", func(t *testing.T) {
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)

		var roleData = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
		}

		// create a role
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)

		// read a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.EqualValues(t, int64((24 * time.Hour).Seconds()), resp.Data["token_ttl"])
	})

	t.Run("token_ttl set to 0 should default to config max_ttl", func(t *testing.T) {
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)

		var roleData = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"token_ttl":  0,
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
		}

		// create a role
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Warnings)

		// read a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.EqualValues(t, int64((365 * 24 * time.Hour).Seconds()), resp.Data["token_ttl"])
	})

	t.Run("update handler existence check", func(t *testing.T) {
		var b, l, err = getBackend()
		require.NoError(t, err)
		hasExistenceCheck, exists, err := b.HandleExistenceCheck(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})

		require.True(t, hasExistenceCheck)
		require.False(t, exists)
		require.NoError(t, err)
	})

	t.Run("full flow check roles", func(t *testing.T) {
		var b, l, events, err = getBackendWithEvents()
		require.NoError(t, err)

		// create a configuration with max ttl set to 10 days
		func() {
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.UpdateOperation,
				Path:      gitlab.PathConfigStorage, Storage: l,
				Data: map[string]any{
					"max_ttl": (10 * 24 * time.Hour).Seconds(),
					"token":   "token",
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
		}()

		var roleData = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"token_ttl":  int64((5 * 24 * time.Hour).Seconds()),
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
		}

		// create a role
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.Empty(t, resp.Warnings)

		// read a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.EqualValues(t, "test", resp.Data["role_name"])
		require.Equal(t, int64((5 * 24 * time.Hour).Seconds()), resp.Data["token_ttl"])
		require.Subset(t, resp.Data, roleData)

		// update a role
		roleData["name"] = "Example user personal token - updated"
		roleData["path"] = "user2"
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			Data: roleData,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.Subset(t, resp.Data, roleData)
		require.EqualValues(t, "test", resp.Data["role_name"])

		// read a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NoError(t, resp.Error())
		require.Subset(t, resp.Data, roleData)
		require.EqualValues(t, "test", resp.Data["role_name"])
		require.EqualValues(t, "user2", resp.Data["path"])
		require.EqualValues(t, "Example user personal token - updated", resp.Data["name"])

		// delete a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
		})
		require.NoError(t, err)
		require.Nil(t, resp)

		// read a role
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
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
