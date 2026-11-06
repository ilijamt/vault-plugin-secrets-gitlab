package config_test

import (
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestPeriodicRefresh(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	run := func(t *testing.T, cfg *modelConfig.EntryConfig, client *mockGitlabClient) (*mockConfigBackend, error) {
		t.Helper()
		storage := &logical.InmemStorage{}
		entry, err := logical.StorageEntryJSON("config/default", cfg)
		require.NoError(t, err)
		require.NoError(t, storage.Put(t.Context(), entry))

		mb := &mockConfigBackend{config: cfg, client: client}
		ctx := utils.WithStaticTime(gitlab.ClientNewContext(t.Context(), client), now)
		return mb, pathConfig.New(mb).PeriodicFunc(ctx, &logical.Request{Storage: storage})
	}

	t.Run("stale data is refreshed and saved", func(t *testing.T) {
		cfg := testConfig()
		cfg.GitlabRefreshedAt = now.Add(-2 * time.Hour)
		mb, err := run(t, cfg, &mockGitlabClient{metadata: testMetadata()})
		require.NoError(t, err)
		require.NotNil(t, mb.savedConfig)
		assert.Equal(t, now, mb.savedConfig.GitlabRefreshedAt)
		assert.Equal(t, testMetadata().Version, mb.savedConfig.GitlabVersion)
		assert.Equal(t, int64(30), mb.savedConfig.GitlabInstance.MaxTokenLifetimeDays)
	})

	t.Run("recent data is left alone", func(t *testing.T) {
		cfg := testConfig()
		cfg.GitlabRefreshedAt = now.Add(-30 * time.Minute)
		mb, err := run(t, cfg, &mockGitlabClient{metadata: testMetadata()})
		require.NoError(t, err)
		assert.Nil(t, mb.savedConfig)
	})

	t.Run("failed refresh keeps the data and is not an error", func(t *testing.T) {
		cfg := testConfig()
		cfg.GitlabRefreshedAt = now.Add(-2 * time.Hour)
		cfg.GitlabVersion = "19.0.0"
		mb, err := run(t, cfg, &mockGitlabClient{metadata: testMetadata(), instanceErr: errors.New("down")})
		require.NoError(t, err)
		assert.Nil(t, mb.savedConfig)
	})
}
