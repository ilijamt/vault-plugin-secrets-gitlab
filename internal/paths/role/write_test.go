package role_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func personalRaw() map[string]interface{} {
	return map[string]interface{}{
		"role_name":  "test-role",
		"path":       "testuser",
		"name":       "test-token",
		"token_type": token.TypePersonal.String(),
		"scopes":     token.ScopeApi.String(),
		"ttl":        3600,
	}
}

func TestPathRolesWrite_HappyPath(t *testing.T) {
	t.Run("personal token", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		resp, err := writeHandler(&mockRoleBackend{
			config: testConfig(),
			sendEvent: func(_ context.Context, et event.EventType, md map[string]string) error {
				sentEventType = et
				sentMetadata = md
				return nil
			},
		})(t.Context(), newRequest(), newFieldData(personalRaw()))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())

		assert.Equal(t, "test-role", resp.Data["role_name"])
		assert.Equal(t, "default", resp.Data["config_name"])
		assert.Equal(t, "role-write", sentEventType.String())
		assert.Equal(t, "test-role", sentMetadata["role_name"])
	})

	t.Run("group token with access level", func(t *testing.T) {
		resp, err := writeHandler(&mockRoleBackend{config: testConfig()})(
			t.Context(), newRequest(), newFieldData(map[string]interface{}{
				"role_name":    "group-role",
				"path":         "my-group/sub",
				"name":         "group-token",
				"token_type":   token.TypeGroup.String(),
				"access_level": token.AccessLevelDeveloperPermissions.String(),
				"scopes":       token.ScopeApi.String(),
				"ttl":          86400,
			}))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())
	})

	t.Run("pipeline trigger without TTL", func(t *testing.T) {
		resp, err := writeHandler(&mockRoleBackend{config: testConfig()})(
			t.Context(), newRequest(), newFieldData(map[string]interface{}{
				"role_name":  "test-role",
				"path":       "my-group/my-project",
				"name":       "trigger",
				"token_type": token.TypePipelineProjectTrigger.String(),
			}))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())
		assert.Equal(t, int64(0), resp.Data["ttl"])
	})

	t.Run("dynamic path with valid regex", func(t *testing.T) {
		raw := personalRaw()
		raw["path"] = "test-.*123$"
		raw["dynamic_path"] = true
		resp, err := writeHandler(&mockRoleBackend{config: testConfig()})(
			t.Context(), newRequest(), newFieldData(raw))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())
		assert.Equal(t, true, resp.Data["dynamic_path"])
	})
}

func TestPathRolesWrite_ConfigErrors(t *testing.T) {
	t.Run("config not found", func(t *testing.T) {
		resp, err := writeHandler(&mockRoleBackend{config: nil})(
			t.Context(), newRequest(), newFieldData(personalRaw()))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsError())
	})

	t.Run("GetConfig error", func(t *testing.T) {
		resp, err := writeHandler(&mockRoleBackend{configErr: errors.New("storage failure")})(
			t.Context(), newRequest(), newFieldData(personalRaw()))
		require.Error(t, err)
		require.NotNil(t, resp)
	})
}

func TestPathRolesWrite_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		raw         map[string]interface{}
		config      func() *mockRoleBackend
		errContains string
	}{
		{
			name: "invalid token type",
			raw: func() map[string]interface{} {
				r := personalRaw()
				r["token_type"] = "invalid_type"
				return r
			}(),
			errContains: "token_type",
		},
		{
			name: "invalid scopes for project token",
			raw: map[string]interface{}{
				"role_name":    "test-role",
				"path":         "my-group/my-project",
				"name":         "proj-token",
				"token_type":   token.TypeProject.String(),
				"access_level": token.AccessLevelDeveloperPermissions.String(),
				"scopes":       token.ScopeSudo.String(),
				"ttl":          86400,
			},
			errContains: "scopes",
		},
		{
			name: "empty scopes for deploy token",
			raw: map[string]interface{}{
				"role_name":  "test-role",
				"path":       "my-group/my-project",
				"name":       "deploy-token",
				"token_type": token.TypeProjectDeploy.String(),
				"scopes":     "",
			},
		},
		{
			name: "TTL exceeds max",
			raw: func() map[string]interface{} {
				r := personalRaw()
				r["ttl"] = 365*24*3600 + 3600
				return r
			}(),
			errContains: "max_ttl",
		},
		{
			name: "TTL below 1h when vault revokes",
			raw: func() map[string]interface{} {
				r := personalRaw()
				r["ttl"] = 3599
				r["gitlab_revokes_token"] = false
				return r
			}(),
			errContains: "ttl",
		},
		{
			name: "TTL below 24h when gitlab revokes",
			raw: func() map[string]interface{} {
				r := personalRaw()
				r["ttl"] = 23*3600 + 3599
				r["gitlab_revokes_token"] = true
				return r
			}(),
			errContains: "ttl",
		},
		{
			name: "access level on personal token",
			raw: func() map[string]interface{} {
				r := personalRaw()
				r["access_level"] = token.AccessLevelOwnerPermissions.String()
				return r
			}(),
			errContains: "access_level",
		},
		{
			name: "missing access level on group token",
			raw: map[string]interface{}{
				"role_name":  "test-role",
				"path":       "my-group/sub",
				"name":       "group-token",
				"token_type": token.TypeGroup.String(),
				"scopes":     token.ScopeApi.String(),
				"ttl":        86400,
			},
			errContains: "access_level",
		},
		{
			name: "dynamic path with invalid regex",
			raw: func() map[string]interface{} {
				r := personalRaw()
				r["path"] = "[a-z"
				r["dynamic_path"] = true
				return r
			}(),
		},
		{
			name: "invalid name template",
			raw: func() map[string]interface{} {
				r := personalRaw()
				r["name"] = "{{ . } invalid template"
				return r
			}(),
			errContains: "invalid template",
		},
		{
			name: "user service account with SaaS config",
			raw: map[string]interface{}{
				"role_name":  "test-role",
				"path":       "testuser",
				"name":       "sa-token",
				"token_type": token.TypeUserServiceAccount.String(),
				"scopes":     token.ScopeApi.String(),
				"ttl":        3600,
			},
			config: func() *mockRoleBackend {
				cfg := testConfig()
				cfg.Type = gitlabTypes.TypeSaaS
				return &mockRoleBackend{config: cfg}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &mockRoleBackend{config: testConfig()}
			if tt.config != nil {
				mb = tt.config()
			}
			resp, err := writeHandler(mb)(t.Context(), newRequest(), newFieldData(tt.raw))
			require.Error(t, err)
			require.NotNil(t, resp)
			if tt.errContains != "" {
				assert.Contains(t, err.Error(), tt.errContains)
			}
		})
	}
}
