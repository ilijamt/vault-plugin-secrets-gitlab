package role_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	modelRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
)

func TestPathRolesDelete(t *testing.T) {
	fd := newFieldData(map[string]interface{}{"role_name": "test-role"})

	t.Run("happy path", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		s := &logical.InmemStorage{}
		require.NoError(t, s.Put(t.Context(), &logical.StorageEntry{
			Key:   "roles/test-role",
			Value: []byte("{}"),
		}))

		resp, err := deleteHandler(&mockRoleBackend{
			role: &modelRole.Role{RoleName: "test-role"},
			sendEvent: func(_ context.Context, et event.EventType, md map[string]string) error {
				sentEventType = et
				sentMetadata = md
				return nil
			},
		})(t.Context(), &logical.Request{Storage: s}, fd)
		require.NoError(t, err)
		assert.Nil(t, resp)

		assert.Equal(t, "role-delete", sentEventType.String())
		assert.Equal(t, "test-role", sentMetadata["role_name"])

		entry, err := s.Get(t.Context(), "roles/test-role")
		require.NoError(t, err)
		assert.Nil(t, entry)
	})

	t.Run("GetRole error", func(t *testing.T) {
		resp, err := deleteHandler(&mockRoleBackend{roleErr: errors.New("storage failure")})(
			t.Context(), newRequest(), fd)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}
