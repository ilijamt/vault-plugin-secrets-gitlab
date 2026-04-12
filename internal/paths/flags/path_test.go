package flags_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	pathflags "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/flags"
)

func TestPathFlags_Name(t *testing.T) {
	mb := &mockFlagsBackend{}
	p := pathflags.New(mb)
	assert.Equal(t, "flags", p.Name())
}

func TestPathFlags_UpdateOperationAvailability(t *testing.T) {
	t.Run("runtime flags change disabled", func(t *testing.T) {
		mb := &mockFlagsBackend{flags: flags.Flags{AllowRuntimeFlagsChange: false}}

		p := pathflags.New(mb)
		paths := p.Paths()
		require.Len(t, paths, 1)

		assert.NotNil(t, paths[0].Operations[logical.ReadOperation])
		assert.Nil(t, paths[0].Operations[logical.UpdateOperation])
	})

	t.Run("runtime flags change enabled", func(t *testing.T) {
		mb := &mockFlagsBackend{flags: flags.Flags{AllowRuntimeFlagsChange: true}}

		p := pathflags.New(mb)
		paths := p.Paths()
		require.Len(t, paths, 1)

		assert.NotNil(t, paths[0].Operations[logical.ReadOperation])
		assert.NotNil(t, paths[0].Operations[logical.UpdateOperation])
	})
}
