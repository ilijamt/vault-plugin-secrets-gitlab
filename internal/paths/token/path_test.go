package token_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pathtoken "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
)

func TestProvider_Name(t *testing.T) {
	p := pathtoken.New(&mockTokenBackend{}, &framework.Secret{})
	assert.Equal(t, "token", p.Name())
}

func TestProvider_Paths(t *testing.T) {
	p := pathtoken.New(&mockTokenBackend{}, &framework.Secret{})
	paths := p.Paths()
	require.Len(t, paths, 1)

	path := paths[0]
	assert.NotNil(t, path.Operations[logical.ReadOperation])
	assert.NotNil(t, path.Operations[logical.UpdateOperation])
	assert.Contains(t, path.Pattern, pathtoken.PathTokenRoleStorage)
}
