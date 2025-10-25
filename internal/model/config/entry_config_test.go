package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func TestEntryConfig(t *testing.T) {
	cfg := config.EntryConfig{
		TokenCreatedAt: time.Now(),
		TokenExpiresAt: time.Now(),
	}

	require.Contains(t, cfg.LogicalResponseData(true), "token")
	require.NotContains(t, cfg.LogicalResponseData(false), "token")
}
