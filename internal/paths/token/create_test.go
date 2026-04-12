package token_test

import (
	"errors"
	"testing"
	"time"

	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	modelRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	pathtoken "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	tk "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

var (
	testNow       = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	testExpiresAt = testNow.Add(time.Hour)
	errTest       = errors.New("test error")
)

func callCreate(t *testing.T, mb *mockTokenBackend, raw map[string]any) (*logical.Response, error) {
	t.Helper()
	s := &framework.Secret{Type: "access_tokens"}
	p := pathtoken.New(mb, s).Paths()[0]
	fd := &framework.FieldData{Raw: raw, Schema: p.Fields}
	ctx := utils.WithStaticTime(t.Context(), testNow)
	return p.Operations[logical.ReadOperation].Handler()(ctx, &logical.Request{}, fd)
}

func role(tokenType tk.Type, path string) *modelRole.Role {
	return &modelRole.Role{RoleName: "r", TTL: time.Hour, Path: path, Name: "n", Scopes: []string{"api"}, TokenType: tokenType, AccessLevel: tk.AccessLevelDeveloperPermissions}
}

func TestPathTokenRoleCreate_Errors(t *testing.T) {
	badName := role(tk.TypeProject, "p")
	badName.Name = "{{"

	tests := []struct {
		name    string
		backend *mockTokenBackend
		errMsg  string
	}{
		{"role error", &mockTokenBackend{roleErr: errTest}, "error getting role"},
		{"role not found", &mockTokenBackend{}, "not found"},
		{"client error", &mockTokenBackend{role: role(tk.TypeProject, "p"), clientErr: errTest}, "test error"},
		{"unknown type", &mockTokenBackend{role: role(tk.Type("invalid"), "p"), client: &mockGitlabClient{}}, "unknown token type"},
		{"invalid name", &mockTokenBackend{role: badName, client: &mockGitlabClient{}}, "error generating token name"},
		{"create error", &mockTokenBackend{role: role(tk.TypeProject, "p"), client: &mockGitlabClient{createErr: errTest}}, "test error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := callCreate(t, tt.backend, map[string]any{"role_name": "r"})
			require.ErrorContains(t, err, tt.errMsg)
		})
	}
}

func TestPathTokenRoleCreate_LookupError(t *testing.T) {
	for _, tt := range []struct {
		name      string
		tokenType tk.Type
		path      string
	}{
		{"personal", tk.TypePersonal, "user"},
		{"user-service-account", tk.TypeUserServiceAccount, "user"},
		{"group-service-account", tk.TypeGroupServiceAccount, "group/sa"},
		{"project-deploy", tk.TypeProjectDeploy, "g/p"},
		{"group-deploy", tk.TypeGroupDeploy, "g"},
		{"pipeline-project-trigger", tk.TypePipelineProjectTrigger, "g/p"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mb := &mockTokenBackend{
				role:   role(tt.tokenType, tt.path),
				client: &mockGitlabClient{lookupErr: errTest},
			}
			_, err := callCreate(t, mb, map[string]any{"role_name": "r"})
			require.ErrorContains(t, err, "test error")
		})
	}
}

func TestPathTokenRoleCreate_Success(t *testing.T) {
	for _, tt := range []struct {
		name      string
		tokenType tk.Type
		path      string
	}{
		{"project", tk.TypeProject, "g/p"},
		{"group", tk.TypeGroup, "g"},
		{"personal", tk.TypePersonal, "user"},
		{"user-service-account", tk.TypeUserServiceAccount, "user"},
		{"group-service-account", tk.TypeGroupServiceAccount, "group/sa"},
		{"project-deploy", tk.TypeProjectDeploy, "g/p"},
		{"group-deploy", tk.TypeGroupDeploy, "g"},
		{"pipeline-project-trigger", tk.TypePipelineProjectTrigger, "g/p"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var sentEventType event.EventType
			var sentMetadata map[string]string

			mb := &mockTokenBackend{
				role:   role(tt.tokenType, tt.path),
				client: &mockGitlabClient{token: newToken(tt.tokenType, testNow, testExpiresAt)},
				sendEvent: func(_ context.Context, eventType event.EventType, metadata map[string]string) error {
					sentEventType = eventType
					sentMetadata = metadata
					return nil
				},
			}
			resp, err := callCreate(t, mb, map[string]any{"role_name": "r"})
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, "glpat-test", resp.Data["token"])
			assert.Equal(t, time.Hour, resp.Secret.TTL)
			assert.Equal(t, time.Hour, resp.Secret.MaxTTL)
			assert.Equal(t, testNow, resp.Secret.IssueTime)

			assert.Equal(t, "token-write", sentEventType.String())
			assert.Equal(t, "roles/r", sentMetadata["path"])
			assert.Equal(t, tt.tokenType.String(), sentMetadata["token_type"])
			assert.Equal(t, "r", sentMetadata["role_name"])
		})
	}
}

func TestPathTokenRoleCreate_RevocationMode(t *testing.T) {
	// Use a token expiry different from role TTL (1h) to distinguish code paths.
	tokenExpiresAt := testNow.Add(30 * time.Minute)

	t.Run("gitlab revokes", func(t *testing.T) {
		r := role(tk.TypeProject, "p")
		r.GitlabRevokesTokens = true
		mb := &mockTokenBackend{
			role:   r,
			client: &mockGitlabClient{token: newToken(tk.TypeProject, testNow, tokenExpiresAt)},
		}
		resp, err := callCreate(t, mb, map[string]any{"role_name": "r"})
		require.NoError(t, err)
		assert.Equal(t, 30*time.Minute, resp.Secret.TTL, "TTL should come from token, not role")
		assert.Equal(t, time.Hour, resp.Secret.MaxTTL)
	})

	t.Run("vault revokes", func(t *testing.T) {
		r := role(tk.TypeProject, "p")
		r.GitlabRevokesTokens = false
		mb := &mockTokenBackend{
			role:   r,
			client: &mockGitlabClient{token: newToken(tk.TypeProject, testNow, tokenExpiresAt)},
		}
		resp, err := callCreate(t, mb, map[string]any{"role_name": "r"})
		require.NoError(t, err)
		assert.Equal(t, time.Hour, resp.Secret.TTL, "TTL should be role TTL")
		assert.Equal(t, testNow.Add(time.Hour), resp.Data["expires_at"].(*time.Time).UTC(), "ExpiresAt should be startTime + role.TTL")
	})
}

func TestPathTokenRoleCreate_DynamicPath(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		r := role(tk.TypeProject, `^allowed/.*$`)
		r.DynamicPath = true
		mb := &mockTokenBackend{role: r}
		_, err := callCreate(t, mb, map[string]any{"role_name": "r", "path": ""})
		require.ErrorContains(t, err, "not valid")
	})

	t.Run("regexp mismatch", func(t *testing.T) {
		r := role(tk.TypeProject, `^allowed/.*$`)
		r.DynamicPath = true
		mb := &mockTokenBackend{role: r}
		_, err := callCreate(t, mb, map[string]any{"role_name": "r", "path": "other/project"})
		require.ErrorContains(t, err, "regexp")
	})

	t.Run("success", func(t *testing.T) {
		r := role(tk.TypeProject, `^allowed/.*$`)
		r.DynamicPath = true
		mb := &mockTokenBackend{
			role:   r,
			client: &mockGitlabClient{token: newToken(tk.TypeProject, testNow, testExpiresAt)},
		}
		resp, err := callCreate(t, mb, map[string]any{"role_name": "r", "path": "allowed/my-project"})
		require.NoError(t, err)
		assert.Equal(t, "glpat-test", resp.Data["token"])
	})
}
