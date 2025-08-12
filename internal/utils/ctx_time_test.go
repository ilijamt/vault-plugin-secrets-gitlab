package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestWithTime(t *testing.T) {
	t.Run("no time should default to time now", func(t *testing.T) {
		tm := utils.TimeFromContext(t.Context())
		require.False(t, tm.IsZero())
	})

	t.Run("with time", func(t *testing.T) {
		tm := time.Date(2009, 1, 1, 1, 0, 0, 0, time.UTC)
		ctx := utils.WithStaticTime(t.Context(), tm)
		tmCtx := utils.TimeFromContext(ctx)
		require.False(t, tmCtx.IsZero())
	})
}
