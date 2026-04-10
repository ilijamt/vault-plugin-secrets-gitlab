package flags_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	pathflags "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/flags"
)

func TestPathFlagsRead(t *testing.T) {
	mb := newMockFlagsBackend(t)
	mb.MockFlagsProvider.EXPECT().Flags().Return(flags.Flags{})

	p := pathflags.New(mb)
	paths := p.Paths()
	require.Len(t, paths, 1)

	readOp := paths[0].Operations[logical.ReadOperation]
	require.NotNil(t, readOp)

	resp, err := readOp.Handler()(t.Context(), &logical.Request{}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, false, resp.Data["show_config_token"])
	assert.Equal(t, false, resp.Data["allow_runtime_flags_change"])
}

func TestPathFlagsRead_WithCustomFlags(t *testing.T) {
	mb := newMockFlagsBackend(t)
	mb.MockFlagsProvider.EXPECT().Flags().Return(flags.Flags{
		ShowConfigToken:         true,
		AllowRuntimeFlagsChange: true,
	})

	p := pathflags.New(mb)
	paths := p.Paths()

	readOp := paths[0].Operations[logical.ReadOperation]
	require.NotNil(t, readOp)

	resp, err := readOp.Handler()(t.Context(), &logical.Request{}, &framework.FieldData{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, true, resp.Data["show_config_token"])
	assert.Equal(t, true, resp.Data["allow_runtime_flags_change"])
}
