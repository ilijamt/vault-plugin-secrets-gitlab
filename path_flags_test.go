//go:build unit

package gitlab_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

func TestPathFlags(t *testing.T) {
	var ctx = t.Context()
	b, l, events, err := getBackendWithFlagsWithEvents(ctx, flags.Flags{AllowRuntimeFlagsChange: true})
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.ReadOperation,
		Path:      gitlab.PathConfigFlags, Storage: l,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.True(t, resp.Data["allow_runtime_flags_change"].(bool))
	require.False(t, resp.Data["show_config_token"].(bool))

	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigFlags, Storage: l,
		Data: map[string]interface{}{
			"show_config_token":          "true",
			"allow_runtime_flags_change": "false",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.True(t, resp.Data["show_config_token"].(bool))
	require.True(t, resp.Data["allow_runtime_flags_change"].(bool))

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/flags-write"},
	})
}
