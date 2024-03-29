package gitlab_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestPathConfigRotate(t *testing.T) {
	t.Run("initial config should be empty", func(t *testing.T) {
		b, l, err := getBackend()
		require.NoError(t, err)
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("%s/rotate", gitlab.PathConfigStorage), Storage: l,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Error(t, resp.Error())
		require.EqualValues(t, resp.Error(), gitlab.ErrBackendNotConfigured)
	})
}
