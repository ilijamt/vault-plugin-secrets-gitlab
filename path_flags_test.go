//go:build unit

package gitlab_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathFlags(t *testing.T) {
	var ctx = t.Context()
	b, l, events, err := getBackendWithFlagsWithEvents(ctx, gitlab.Flags{AllowRuntimeFlagsChange: true})
	require.NoError(t, err)

	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.ReadOperation,
		Path:      gitlab.PathConfigFlags, Storage: l,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.False(t, resp.Data["show_config_token"].(bool))

	resp, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      gitlab.PathConfigFlags, Storage: l,
		Data: map[string]interface{}{
			"show_config_token": "true",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NoError(t, resp.Error())
	require.True(t, resp.Data["show_config_token"].(bool))

	events.expectEvents(t, []expectedEvent{
		{eventType: "gitlab/flags-write"},
	})
}
