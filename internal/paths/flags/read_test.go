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
	tests := map[string]struct {
		flags flags.Flags
		want  map[string]any
	}{
		"default flags": {
			flags: flags.Flags{},
			want: map[string]any{
				"show_config_token":          false,
				"allow_runtime_flags_change": false,
			},
		},
		"custom flags": {
			flags: flags.Flags{
				ShowConfigToken:         true,
				AllowRuntimeFlagsChange: true,
			},
			want: map[string]any{
				"show_config_token":          true,
				"allow_runtime_flags_change": true,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mb := &mockFlagsBackend{flags: tt.flags}
			p := pathflags.New(mb)
			paths := p.Paths()
			require.Len(t, paths, 1)

			readOp := paths[0].Operations[logical.ReadOperation]
			require.NotNil(t, readOp)

			resp, err := readOp.Handler()(t.Context(), &logical.Request{}, &framework.FieldData{})
			require.NoError(t, err)
			require.NotNil(t, resp)
			for k, v := range tt.want {
				assert.Equal(t, v, resp.Data[k], "key: %s", k)
			}
		})
	}
}
