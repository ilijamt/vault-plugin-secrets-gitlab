package flags_test

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	pathflags "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/flags"
)

func TestPathFlagsUpdate(t *testing.T) {
	t.Run("sets show_config_token and sends event", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		mb := &mockFlagsBackend{
			flags: flags.Flags{AllowRuntimeFlagsChange: true},
			sendEvent: func(_ context.Context, eventType event.EventType, metadata map[string]string) error {
				sentEventType = eventType
				sentMetadata = metadata
				return nil
			},
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
		assert.True(t, mb.flags.ShowConfigToken)

		assert.Equal(t, "flags-write", sentEventType.String())
		assert.Equal(t, "true", sentMetadata["show_config_token"])
	})

	t.Run("allow_runtime_flags_change is not modifiable", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		mb := &mockFlagsBackend{
			flags: flags.Flags{
				AllowRuntimeFlagsChange: true,
				ShowConfigToken:         false,
			},
			sendEvent: func(_ context.Context, eventType event.EventType, metadata map[string]string) error {
				sentEventType = eventType
				sentMetadata = metadata
				return nil
			},
		}

		p := pathflags.New(mb)
		paths := p.Paths()

		updateOp := paths[0].Operations[logical.UpdateOperation]
		require.NotNil(t, updateOp)

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{"allow_runtime_flags_change": false},
			Schema: paths[0].Fields,
		}

		resp, err := updateOp.Handler()(t.Context(), &logical.Request{}, fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, true, resp.Data["allow_runtime_flags_change"])
		assert.True(t, mb.flags.AllowRuntimeFlagsChange)

		assert.Equal(t, "flags-write", sentEventType.String())
		assert.Empty(t, sentMetadata)
	})

	t.Run("no fields provided leaves flags unchanged", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		mb := &mockFlagsBackend{
			flags: flags.Flags{
				AllowRuntimeFlagsChange: true,
				ShowConfigToken:         true,
			},
			sendEvent: func(_ context.Context, eventType event.EventType, metadata map[string]string) error {
				sentEventType = eventType
				sentMetadata = metadata
				return nil
			},
		}

		p := pathflags.New(mb)
		paths := p.Paths()

		updateOp := paths[0].Operations[logical.UpdateOperation]
		require.NotNil(t, updateOp)

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{},
			Schema: paths[0].Fields,
		}

		resp, err := updateOp.Handler()(t.Context(), &logical.Request{}, fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, true, resp.Data["show_config_token"])
		assert.True(t, mb.flags.ShowConfigToken)

		assert.Equal(t, "flags-write", sentEventType.String())
		assert.Empty(t, sentMetadata)
	})
}
