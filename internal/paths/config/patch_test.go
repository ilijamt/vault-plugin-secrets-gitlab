package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestPathConfigPatch(t *testing.T) {
	configPath := pathConfig.New(&mockConfigBackend{}).Paths()[0]

	t.Run("config not found", func(t *testing.T) {
		mb := &mockConfigBackend{config: nil}
		p := pathConfig.New(mb)
		patchOp := p.Paths()[0].Operations[logical.PatchOperation].Handler()

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default", "type": gitlabTypes.TypeSaaS.String()},
			Schema: configPath.Fields,
		}
		resp, err := patchOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsError())
		assert.Contains(t, resp.Error().Error(), errs.ErrBackendNotConfigured.Error())
	})

	t.Run("GetConfig error", func(t *testing.T) {
		mb := &mockConfigBackend{configErr: errors.New("storage failure")}
		p := pathConfig.New(mb)
		patchOp := p.Paths()[0].Operations[logical.PatchOperation].Handler()

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default", "type": gitlabTypes.TypeSaaS.String()},
			Schema: configPath.Fields,
		}
		resp, err := patchOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("patch type only without token change", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		cfg := testConfig()
		mb := &mockConfigBackend{
			config: cfg,
			sendEvent: func(_ context.Context, eventType event.EventType, metadata map[string]string) error {
				sentEventType = eventType
				sentMetadata = metadata
				return nil
			},
		}
		p := pathConfig.New(mb)
		patchOp := p.Paths()[0].Operations[logical.PatchOperation].Handler()

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default", "type": gitlabTypes.TypeSaaS.String()},
			Schema: configPath.Fields,
		}
		resp, err := patchOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())

		assert.Equal(t, gitlabTypes.TypeSaaS.String(), resp.Data["type"])
		assert.Equal(t, "config-patch", sentEventType.String())
		assert.Equal(t, gitlabTypes.TypeSaaS.String(), sentMetadata["type"])
		assert.Equal(t, "default", mb.deleteClientName)
		require.NotNil(t, mb.savedConfig)
		assert.Equal(t, gitlabTypes.TypeSaaS, mb.savedConfig.Type)
	})

	t.Run("patch with token change calls updateConfigClientInfo", func(t *testing.T) {
		var sentEventType event.EventType

		cfg := testConfig()
		mb := &mockConfigBackend{
			config: cfg,
			sendEvent: func(_ context.Context, eventType event.EventType, _ map[string]string) error {
				sentEventType = eventType
				return nil
			},
		}
		p := pathConfig.New(mb)
		patchOp := p.Paths()[0].Operations[logical.PatchOperation].Handler()

		mc := &mockGitlabClient{
			tokenInfo: testTokenInfo(),
			metadata:  testMetadata(),
		}
		ctx := gitlab.ClientNewContext(t.Context(), mc)

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default", "token": "glpat-new-token"},
			Schema: configPath.Fields,
		}
		resp, err := patchOp(ctx, &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())

		assert.Equal(t, "config-patch", sentEventType.String())
		assert.Equal(t, "default", mb.deleteClientName)
		require.NotNil(t, mb.savedConfig)
		assert.Equal(t, "glpat-new-token", mb.savedConfig.Token)
		assert.Equal(t, testTokenInfo().TokenID, mb.savedConfig.TokenId)
	})

	t.Run("patch with invalid token change fails", func(t *testing.T) {
		cfg := testConfig()
		mb := &mockConfigBackend{config: cfg}
		p := pathConfig.New(mb)
		patchOp := p.Paths()[0].Operations[logical.PatchOperation].Handler()

		mc := &mockGitlabClient{
			tokenInfoErr: errors.New("unauthorized"),
		}
		ctx := gitlab.ClientNewContext(t.Context(), mc)

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default", "token": "glpat-bad-token"},
			Schema: configPath.Fields,
		}
		resp, err := patchOp(ctx, &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}
