//go:build unit

package integration_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

func TestPathConfigRotate(t *testing.T) {
	t.Run("initial config should be empty fail with backend not configured", func(t *testing.T) {
		ctx := getCtxGitlabClient(t, "unit")
		b, l, err := getBackend(ctx)
		require.NoError(t, err)
		resp, err := b.HandleRequest(ctx, &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      fmt.Sprintf("%s/%s/rotate", backend.PathConfigStorage, backend.DefaultConfigName), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, resp.Error(), errs.ErrBackendNotConfigured)
	})
}
