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

func TestPathFlagsUpdate(t *testing.T) {
	mb := &mockFlagsBackend{
		flags: flags.Flags{AllowRuntimeFlagsChange: true},
	}

	p := pathflags.New(mb)
	paths := p.Paths()

	updateOp := paths[0].Operations[logical.UpdateOperation]
	require.NotNil(t, updateOp)

	fd := &framework.FieldData{
		Raw:    map[string]interface{}{"show_config_token": true},
		Schema: paths[0].Fields,
	}

	resp, err := updateOp.Handler()(t.Context(), &logical.Request{}, fd)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, true, resp.Data["show_config_token"])
	assert.Equal(t, true, resp.Data["allow_runtime_flags_change"])
	assert.True(t, mb.flags.ShowConfigToken)
}
