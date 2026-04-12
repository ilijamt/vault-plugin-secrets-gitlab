package config_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestProvider_Name(t *testing.T) {
	p := pathConfig.New(&mockConfigBackend{})
	assert.Equal(t, "config", p.Name())
}

func TestProvider_Paths(t *testing.T) {
	p := pathConfig.New(&mockConfigBackend{})
	paths := p.Paths()
	require.Len(t, paths, 3)

	t.Run("config CRUD path has expected operations", func(t *testing.T) {
		configPath := paths[0]
		assert.NotNil(t, configPath.Operations[logical.ReadOperation])
		assert.NotNil(t, configPath.Operations[logical.UpdateOperation])
		assert.NotNil(t, configPath.Operations[logical.PatchOperation])
		assert.NotNil(t, configPath.Operations[logical.DeleteOperation])
	})

	t.Run("list path has list operation", func(t *testing.T) {
		listPath := paths[1]
		assert.NotNil(t, listPath.Operations[logical.ListOperation])
	})

	t.Run("rotate path has update operation", func(t *testing.T) {
		rotatePath := paths[2]
		assert.NotNil(t, rotatePath.Operations[logical.UpdateOperation])
	})
}
