package gitlab_test

import (
	"context"
	"fmt"
	"maps"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathRolesTTL(t *testing.T) {
	var defaultConfig = map[string]any{"token": "random-token"}

	t.Run("general ttl limits", func(t *testing.T) {
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)

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

		t.Run(fmt.Sprintf("maxTTL > DefaultAccessTokenMaxPossibleTTL [%s]", gitlab.DefaultAccessTokenMaxPossibleTTL), func(t *testing.T) {
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl":     "4h",
				"max_ttl": (gitlab.DefaultAccessTokenMaxPossibleTTL + time.Hour).String(),
			})
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.Error(t, err)
			require.ErrorIs(t, err, gitlab.ErrInvalidValue)
			require.NotNil(t, resp)
			require.True(t, resp.IsError())
			require.ErrorContains(t, resp.Error(), "max_ttl='8761h0m0s' [max_ttl <= 8760h0m0s]")
		})
		t.Run("role.TTL > role.MaxTTL", func(t *testing.T) {
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl":     "49h",
				"max_ttl": "48h",
			})
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.Error(t, err)
			require.ErrorIs(t, err, gitlab.ErrInvalidValue)
			require.NotNil(t, resp)
			require.True(t, resp.IsError())
			require.ErrorContains(t, resp.Error(), "ttl = 49h0m0s [ttl <= max_ttl = 48h0m0s")
		})

		t.Run("ttl = maxTTL", func(t *testing.T) {
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl":     "64h",
				"max_ttl": "64h",
			})
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Empty(t, resp.Warnings)

			// read a role
			resp, err = b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.ReadOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.EqualValues(t, int64((64 * time.Hour).Seconds()), resp.Data["ttl"])
		})
	})

	t.Run("vault revokes the token", func(t *testing.T) {
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)

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

		t.Run("ttl >= 1h && ttl <= maxTTL", func(t *testing.T) {
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl":     "1h",
				"max_ttl": "64h",
			})
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Empty(t, resp.Warnings)

			// read a role
			resp, err = b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.ReadOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NoError(t, resp.Error())
			require.EqualValues(t, int64((1 * time.Hour).Seconds()), resp.Data["ttl"])
		})
		t.Run("ttl < 1h", func(t *testing.T) {
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl":     "59m59s",
				"max_ttl": "24h",
			})
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.Error(t, err)
			require.ErrorIs(t, err, gitlab.ErrInvalidValue)
			require.NotNil(t, resp)
			require.True(t, resp.IsError())
			require.ErrorContains(t, resp.Error(), "ttl = 59m59s [ttl >= 1h]")
		})
	})

	t.Run("gitlab revokes the tokens", func(t *testing.T) {
		var b, l, err = getBackendWithConfig(defaultConfig)
		require.NoError(t, err)

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
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl":     "23h59m59s",
				"max_ttl": "64h",
			})
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.Error(t, err)
			require.ErrorIs(t, err, gitlab.ErrInvalidValue)
			require.NotNil(t, resp)
			require.True(t, resp.IsError())
			require.ErrorContains(t, resp.Error(), "ttl = 23h59m59s [ttl >= 24h0m0s and ttl <= 64h0m0s]")
		})

		t.Run("ttl >= 24h && ttl <= maxTTL", func(t *testing.T) {
			var role = maps.Clone(generalRole)
			maps.Copy(role, map[string]any{
				"ttl":     "24h",
				"max_ttl": "64h",
			})
			resp, err := b.HandleRequest(context.Background(), &logical.Request{
				Operation: logical.CreateOperation,
				Path:      fmt.Sprintf("%s/test", gitlab.PathRoleStorage), Storage: l,
				Data: role,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Empty(t, resp.Warnings)

			// read a role
			resp, err = b.HandleRequest(context.Background(), &logical.Request{
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
