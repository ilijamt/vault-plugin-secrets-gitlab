//go:build unit

package gitlab_test

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

func TestPathRolesTTL(t *testing.T) {
	var defaultConfig = map[string]any{
		"token":    getGitlabToken("admin_user_root").Token,
		"base_url": cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/"),
		"type":     gitlab2.TypeSelfManaged.String(),
	}

	t.Run("general ttl limits", func(t *testing.T) {
		var generalRole = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
			"gitlab_revokes_token": false,
		}

		t.Run("role.TTL > DefaultAccessTokenMaxPossibleTTL", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl": (gitlab.DefaultAccessTokenMaxPossibleTTL + time.Hour).String(),
			})
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.Error(t, err)
			require.ErrorIs(t, err, errs.ErrInvalidValue)
			require.NotNil(t, resp)
			require.True(t, resp.IsError())
			require.ErrorContains(t, resp.Error(), "ttl = 8761h0m0s [ttl <= max_ttl = 8760h0m0s]")
		})

		t.Run("ttl = maxTTL", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl": (gitlab.DefaultAccessTokenMaxPossibleTTL).String(),
			})
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Empty(t, resp.Warnings)

			// read a role
			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.ReadOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.EqualValues(t, int64(gitlab.DefaultAccessTokenMaxPossibleTTL.Seconds()), resp.Data["ttl"])
		})
	})

	t.Run("vault revokes the token", func(t *testing.T) {
		var generalRole = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
			"gitlab_revokes_token": false,
		}

		t.Run("ttl >= 1h && ttl <= DefaultAccessTokenMaxPossibleTTL", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl": "1h",
			})

			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Empty(t, resp.Warnings)

			// read a role
			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.ReadOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.EqualValues(t, int64((1 * time.Hour).Seconds()), resp.Data["ttl"])
		})

		t.Run("ttl < 1h", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl": "59m59s",
			})
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.Error(t, err)
			require.ErrorIs(t, err, errs.ErrInvalidValue)
			require.NotNil(t, resp)
			require.True(t, resp.IsError())
			require.ErrorContains(t, resp.Error(), "ttl = 59m59s [ttl >= 1h]")
		})
	})

	t.Run("gitlab revokes the tokens", func(t *testing.T) {
		var generalRole = map[string]any{
			"path":       "user",
			"name":       "Example user personal token",
			"token_type": gitlab.TokenTypePersonal.String(),
			"scopes": []string{
				gitlab.TokenScopeApi.String(),
				gitlab.TokenScopeReadRegistry.String(),
			},
			"gitlab_revokes_token": true,
		}

		t.Run("ttl < 24h", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl": "23h59m59s",
			})
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.Error(t, err)
			require.ErrorIs(t, err, errs.ErrInvalidValue)
			require.NotNil(t, resp)
			require.True(t, resp.IsError())
			require.ErrorContains(t, resp.Error(), "ttl = 23h59m59s [24h0m0s <= ttl <= 8760h0m0s]")
		})

		t.Run("ttl >= 24h && ttl <= DefaultAccessTokenMaxPossibleTTL", func(t *testing.T) {
			ctx := getCtxGitlabClient(t, "unit")
			var b, l, err = getBackendWithConfig(ctx, defaultConfig)
			require.NoError(t, err)
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl": "24h",
			})
			resp, err := b.HandleRequest(ctx, &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Empty(t, resp.Warnings)

			// read a role
			resp, err = b.HandleRequest(ctx, &logical.Request{
				Operation: logical.ReadOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.EqualValues(t, int64((24 * time.Hour).Seconds()), resp.Data["ttl"])
		})
	})

}
