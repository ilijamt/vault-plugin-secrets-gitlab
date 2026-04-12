package config_test

import (
	"slices"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestPathConfigList(t *testing.T) {
	listPath := pathConfig.New(&mockConfigBackend{}).Paths()[1]

	newFieldData := func() *framework.FieldData {
		return &framework.FieldData{
			Raw:    map[string]interface{}{},
			Schema: listPath.Fields,
		}
	}

	t.Run("empty storage", func(t *testing.T) {
		p := pathConfig.New(&mockConfigBackend{})
		listOp := p.Paths()[1].Operations[logical.ListOperation].Handler()

		resp, err := listOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Data["keys"])
	})

	t.Run("multiple configs", func(t *testing.T) {
		p := pathConfig.New(&mockConfigBackend{})
		listOp := p.Paths()[1].Operations[logical.ListOperation].Handler()

		s := &logical.InmemStorage{}
		for _, name := range []string{"default", "admin", "normal"} {
			require.NoError(t, s.Put(t.Context(), &logical.StorageEntry{
				Key:   "config/" + name,
				Value: []byte("{}"),
			}))
		}

		resp, err := listOp(t.Context(), &logical.Request{Storage: s}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		keys := resp.Data["keys"].([]string)
		slices.Sort(keys)
		assert.Equal(t, []string{"admin", "default", "normal"}, keys)
	})
}
