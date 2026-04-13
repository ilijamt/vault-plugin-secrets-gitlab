package role_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modelRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestPathRolesRead(t *testing.T) {
	fd := newFieldData(map[string]interface{}{"role_name": "test-role"})

	t.Run("role found", func(t *testing.T) {
		resp, err := readHandler(&mockRoleBackend{
			role: &modelRole.Role{
				RoleName:  "test-role",
				Path:      "testuser",
				Name:      "my-token",
				TokenType: token.TypePersonal,
				TTL:       time.Hour,
				Scopes:    []string{"api"},
			},
		})(t.Context(), newRequest(), fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "test-role", resp.Data["role_name"])
		assert.Equal(t, token.TypePersonal.String(), resp.Data["token_type"])
	})

	t.Run("role not found", func(t *testing.T) {
		resp, err := readHandler(&mockRoleBackend{})(t.Context(), newRequest(), fd)
		require.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("GetRole error", func(t *testing.T) {
		resp, err := readHandler(&mockRoleBackend{roleErr: errors.New("storage failure")})(
			t.Context(), newRequest(), fd)
		require.Error(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.IsError())
	})
}
