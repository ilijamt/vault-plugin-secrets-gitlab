package role_test

import (
	"slices"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathRolesList(t *testing.T) {
	fd := newFieldData(map[string]interface{}{})

	t.Run("empty storage", func(t *testing.T) {
		resp, err := listHandler(&mockRoleBackend{})(t.Context(), newRequest(), fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Data["keys"])
	})

	t.Run("multiple roles", func(t *testing.T) {
		s := &logical.InmemStorage{}
		for _, name := range []string{"admin-role", "dev-role", "reader-role"} {
			require.NoError(t, s.Put(t.Context(), &logical.StorageEntry{
				Key:   "roles/" + name,
				Value: []byte("{}"),
			}))
		}

		resp, err := listHandler(&mockRoleBackend{})(t.Context(), &logical.Request{Storage: s}, fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		keys := resp.Data["keys"].([]string)
		slices.Sort(keys)
		assert.Equal(t, []string{"admin-role", "dev-role", "reader-role"}, keys)
	})
}
