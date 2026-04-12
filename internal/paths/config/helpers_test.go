package config_test

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/logical"
	g "gitlab.com/gitlab-org/api/client-go"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	modelToken "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
)

// mockConfigBackend is a hand-written mock satisfying the configBackend interface.
type mockConfigBackend struct {
	// FlagsProvider
	flagsVal flags.Flags

	// ClientReader
	client    gitlab.Client
	clientErr error

	// ClientSetter -- track calls
	setClientName string

	// ClientDeleter -- track calls
	deleteClientName string

	// ConfigStore
	config      *modelConfig.EntryConfig
	configErr   error
	getConfig   func(ctx context.Context, s logical.Storage, name string) (*modelConfig.EntryConfig, error)
	saveErr     error
	savedConfig *modelConfig.EntryConfig

	// EventSender
	sendEvent func(ctx context.Context, eventType event.EventType, metadata map[string]string) error
}

func (m *mockConfigBackend) Logger() hclog.Logger { return hclog.NewNullLogger() }

func (m *mockConfigBackend) Flags() flags.Flags { return m.flagsVal }

func (m *mockConfigBackend) UpdateFlags(fn func(*flags.Flags)) { fn(&m.flagsVal) }

func (m *mockConfigBackend) GetClientByName(_ context.Context, _ logical.Storage, _ string) (gitlab.Client, error) {
	return m.client, m.clientErr
}

func (m *mockConfigBackend) SetClient(_ gitlab.Client, name string) {
	m.setClientName = name
}

func (m *mockConfigBackend) DeleteClient(name string) {
	m.deleteClientName = name
}

func (m *mockConfigBackend) GetConfig(ctx context.Context, s logical.Storage, name string) (*modelConfig.EntryConfig, error) {
	if m.getConfig != nil {
		return m.getConfig(ctx, s, name)
	}
	return m.config, m.configErr
}

func (m *mockConfigBackend) SaveConfig(_ context.Context, _ logical.Storage, cfg *modelConfig.EntryConfig) error {
	m.savedConfig = cfg
	return m.saveErr
}

func (m *mockConfigBackend) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	if m.sendEvent != nil {
		return m.sendEvent(ctx, eventType, metadata)
	}
	return nil
}

// mockGitlabClient is a minimal mock satisfying the gitlab.Client interface
// for updateConfigClientInfo and rotate operations.
type mockGitlabClient struct {
	gitlab.Client

	tokenInfo    *modelToken.TokenConfig
	tokenInfoErr error
	metadata     *g.Metadata
	metadataErr  error
	rotatedToken *modelToken.TokenConfig
	rotatedOld   *modelToken.TokenConfig
	rotateErr    error
}

func (m *mockGitlabClient) CurrentTokenInfo(_ context.Context) (*modelToken.TokenConfig, error) {
	return m.tokenInfo, m.tokenInfoErr
}

func (m *mockGitlabClient) Metadata(_ context.Context) (*g.Metadata, error) {
	return m.metadata, m.metadataErr
}

func (m *mockGitlabClient) RotateCurrentToken(_ context.Context) (*modelToken.TokenConfig, *modelToken.TokenConfig, error) {
	return m.rotatedToken, m.rotatedOld, m.rotateErr
}

func (m *mockGitlabClient) Valid(_ context.Context) bool { return true }

// testConfig returns a realistic EntryConfig for test use.
func testConfig() *modelConfig.EntryConfig {
	now := time.Now()
	expires := now.Add(30 * 24 * time.Hour)
	return &modelConfig.EntryConfig{
		TokenId:          42,
		BaseURL:          "https://gitlab.example.com",
		Token:            "glpat-test-token-value",
		AutoRotateToken:  false,
		AutoRotateBefore: modelConfig.DefaultAutoRotateBeforeMinTTL,
		TokenCreatedAt:   now,
		TokenExpiresAt:   expires,
		Scopes:           []string{"api", "read_user"},
		Type:             gitlabTypes.TypeSelfManaged,
		Name:             "default",
	}
}

// testTokenInfo returns a TokenConfig with sensible defaults.
func testTokenInfo() *modelToken.TokenConfig {
	now := time.Now()
	expires := now.Add(30 * 24 * time.Hour)
	return &modelToken.TokenConfig{
		TokenWithScopes: modelToken.TokenWithScopes{
			Token: modelToken.Token{
				TokenID:   42,
				Token:     "glpat-test-token-value",
				Name:      "test-token",
				CreatedAt: &now,
				ExpiresAt: &expires,
			},
			Scopes: []string{"api", "read_user"},
		},
	}
}

// testMetadata returns a Metadata with sensible defaults.
func testMetadata() *g.Metadata {
	return &g.Metadata{
		Version:    "17.0.0",
		Revision:   "abc123",
		Enterprise: true,
	}
}
