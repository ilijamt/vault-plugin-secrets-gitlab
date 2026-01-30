package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func TestEntryConfig(t *testing.T) {
	cfg := config.EntryConfig{
		Name:           "test",
		TokenCreatedAt: time.Now(),
		TokenExpiresAt: time.Now(),
	}

	require.EqualValues(t, "test", cfg.GetName())
	require.Contains(t, cfg.LogicalResponseData(true), "token")
	require.NotContains(t, cfg.LogicalResponseData(false), "token")
}
