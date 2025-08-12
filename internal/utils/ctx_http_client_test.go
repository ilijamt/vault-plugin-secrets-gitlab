package utils_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestHttpClientFromContext(t *testing.T) {
	t.Run("no http client", func(t *testing.T) {
		c, ok := utils.HttpClientFromContext(t.Context())
		require.False(t, ok)
		require.Nil(t, c)
	})

	t.Run("with http client", func(t *testing.T) {
		ctx := utils.HttpClientNewContext(t.Context(), &http.Client{})
		c, ok := utils.HttpClientFromContext(ctx)
		require.True(t, ok)
		require.NotNil(t, c)
	})
}
