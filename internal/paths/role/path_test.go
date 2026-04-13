package role_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pathRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/role"
)

func TestProvider_Name(t *testing.T) {
	p := pathRole.New(&mockRoleBackend{})
	assert.Equal(t, "role", p.Name())
}

func TestProvider_Paths(t *testing.T) {
	p := pathRole.New(&mockRoleBackend{})
	paths := p.Paths()
	require.Len(t, paths, 2)

	t.Run("list path has list operation", func(t *testing.T) {
		listPath := paths[0]
		assert.NotNil(t, listPath.Operations[logical.ListOperation])
	})

	t.Run("role CRUD path has expected operations", func(t *testing.T) {
		rolePath := paths[1]
		assert.NotNil(t, rolePath.Operations[logical.CreateOperation])
		assert.NotNil(t, rolePath.Operations[logical.UpdateOperation])
		assert.NotNil(t, rolePath.Operations[logical.ReadOperation])
		assert.NotNil(t, rolePath.Operations[logical.DeleteOperation])
	})
}
