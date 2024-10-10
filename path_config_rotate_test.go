//go:build !integration

package gitlab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathConfigRotate(t *testing.T) {
	t.Run("initial config should be empty fail with backend not configured", func(t *testing.T) {
		ctx := getCtxGitlabClient(t)
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/rotate", gitlab.PathConfigStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, resp.Error(), gitlab.ErrBackendNotConfigured)
	})
}
