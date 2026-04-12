package config_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestPathConfigTokenRotate(t *testing.T) {
	rotatePath := pathConfig.New(&mockConfigBackend{}).Paths()[2]

	newFieldData := func() *framework.FieldData {
		return &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default"},
			Schema: rotatePath.Fields,
		}
	}

	t.Run("config not found", func(t *testing.T) {
		mb := &mockConfigBackend{config: nil}
		p := pathConfig.New(mb)
		rotateOp := p.Paths()[2].Operations[logical.UpdateOperation].Handler()

		resp, err := rotateOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsError())
		assert.Contains(t, resp.Error().Error(), errs.ErrBackendNotConfigured.Error())
	})

	t.Run("error propagation", func(t *testing.T) {
		tests := map[string]struct {
			mb     *mockConfigBackend
			errMsg string
		}{
			"GetConfig error": {
				mb:     &mockConfigBackend{configErr: errors.New("storage failure")},
				errMsg: "storage failure",
			},
			"GetClientByName error": {
				mb: &mockConfigBackend{
					config:    testConfig(),
					clientErr: errors.New("client error"),
				},
				errMsg: "client error",
			},
			"RotateCurrentToken error": {
				mb: &mockConfigBackend{
					config: testConfig(),
					client: &mockGitlabClient{rotateErr: errors.New("rotate failed")},
				},
				errMsg: "rotate failed",
			},
			"SaveConfig error": {
				mb: &mockConfigBackend{
					config:  testConfig(),
					client:  &mockGitlabClient{rotatedToken: testTokenInfo(), rotatedOld: testTokenInfo()},
					saveErr: errors.New("save failed"),
				},
				errMsg: "save failed",
			},
		}

		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				p := pathConfig.New(tt.mb)
				rotateOp := p.Paths()[2].Operations[logical.UpdateOperation].Handler()

				resp, err := rotateOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tt.errMsg)
			})
		}
	})

	t.Run("happy path", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		rotatedInfo := testTokenInfo()
		rotatedInfo.Token.Token = "glpat-new-rotated-token"
		rotatedInfo.TokenID = 99

		mb := &mockConfigBackend{
			config: testConfig(),
			client: &mockGitlabClient{rotatedToken: rotatedInfo, rotatedOld: testTokenInfo()},
			sendEvent: func(_ context.Context, eventType event.EventType, metadata map[string]string) error {
				sentEventType = eventType
				sentMetadata = metadata
				return nil
			},
		}
		p := pathConfig.New(mb)
		rotateOp := p.Paths()[2].Operations[logical.UpdateOperation].Handler()

		resp, err := rotateOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())

		assert.Equal(t, "glpat-new-rotated-token", resp.Data["token"])
		assert.Equal(t, "config-token-rotate", sentEventType.String())
		assert.Equal(t, "config/default", sentMetadata["path"])
		assert.Equal(t, "99", sentMetadata["token_id"])
		assert.Equal(t, "default", mb.deleteClientName)

		require.NotNil(t, mb.savedConfig)
		assert.Equal(t, "glpat-new-rotated-token", mb.savedConfig.Token)
		assert.Equal(t, int64(99), mb.savedConfig.TokenId)
	})
}

func TestPeriodicFunc(t *testing.T) {
	t.Run("no configs in storage", func(t *testing.T) {
		p := pathConfig.New(&mockConfigBackend{})

		err := p.PeriodicFunc(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}})
		require.NoError(t, err)
	})

	t.Run("skips rotation when not needed", func(t *testing.T) {
		tests := map[string]struct {
			cfg *modelConfig.EntryConfig
		}{
			"auto_rotate disabled": {
				cfg: func() *modelConfig.EntryConfig {
					c := testConfig()
					c.AutoRotateToken = false
					return c
				}(),
			},
			"token not expiring soon": {
				cfg: func() *modelConfig.EntryConfig {
					c := testConfig()
					c.AutoRotateToken = true
					c.AutoRotateBefore = 24 * time.Hour
					c.TokenExpiresAt = time.Now().Add(100 * 24 * time.Hour)
					return c
				}(),
			},
		}

		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				s := &logical.InmemStorage{}
				require.NoError(t, s.Put(t.Context(), &logical.StorageEntry{
					Key: "config/default", Value: []byte("{}"),
				}))

				mb := &mockConfigBackend{
					getConfig: func(_ context.Context, _ logical.Storage, _ string) (*modelConfig.EntryConfig, error) {
						return tt.cfg, nil
					},
				}
				p := pathConfig.New(mb)

				require.NoError(t, p.PeriodicFunc(t.Context(), &logical.Request{Storage: s}))
				assert.Empty(t, mb.deleteClientName)
			})
		}
	})

	t.Run("token expiring soon triggers rotation", func(t *testing.T) {
		cfg := &modelConfig.EntryConfig{
			Name:             "default",
			Token:            "glpat-old-token",
			TokenId:          42,
			BaseURL:          "https://gitlab.example.com",
			Type:             gitlabTypes.TypeSelfManaged,
			AutoRotateToken:  true,
			AutoRotateBefore: 48 * time.Hour,
			TokenExpiresAt:   time.Now().Add(1 * time.Hour),
			Scopes:           []string{"api"},
		}

		s := &logical.InmemStorage{}
		require.NoError(t, s.Put(t.Context(), &logical.StorageEntry{
			Key: "config/default", Value: []byte("{}"),
		}))

		rotatedInfo := testTokenInfo()
		rotatedInfo.Token.Token = "glpat-rotated"
		rotatedInfo.TokenID = 99

		mb := &mockConfigBackend{
			client: &mockGitlabClient{rotatedToken: rotatedInfo, rotatedOld: testTokenInfo()},
			getConfig: func(_ context.Context, _ logical.Storage, _ string) (*modelConfig.EntryConfig, error) {
				return cfg, nil
			},
		}
		p := pathConfig.New(mb)

		require.NoError(t, p.PeriodicFunc(t.Context(), &logical.Request{Storage: s}))
		assert.Equal(t, "default", mb.deleteClientName)
		require.NotNil(t, mb.savedConfig)
		assert.Equal(t, "glpat-rotated", mb.savedConfig.Token)
	})

	t.Run("GetConfig error is joined", func(t *testing.T) {
		s := &logical.InmemStorage{}
		require.NoError(t, s.Put(t.Context(), &logical.StorageEntry{
			Key: "config/default", Value: []byte("{}"),
		}))

		mb := &mockConfigBackend{
			getConfig: func(_ context.Context, _ logical.Storage, _ string) (*modelConfig.EntryConfig, error) {
				return nil, errors.New("config error")
			},
		}
		p := pathConfig.New(mb)

		err := p.PeriodicFunc(t.Context(), &logical.Request{Storage: s})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config error")
	})
}

func TestInvalidate(t *testing.T) {
	tests := map[string]struct {
		key        string
		wantClient string
	}{
		"matching config key": {key: "config/myconfig", wantClient: "myconfig"},
		"non-matching key":    {key: "roles/myrole", wantClient: ""},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mb := &mockConfigBackend{}
			p := pathConfig.New(mb)
			p.Invalidate(t.Context(), tt.key)
			assert.Equal(t, tt.wantClient, mb.deleteClientName)
		})
	}
}
