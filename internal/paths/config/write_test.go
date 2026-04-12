package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestPathConfigWrite(t *testing.T) {
	configPath := pathConfig.New(&mockConfigBackend{}).Paths()[0]

	validRaw := func() map[string]interface{} {
		return map[string]interface{}{
			"config_name": "default",
			"token":       "glpat-test-token",
			"base_url":    "https://gitlab.example.com",
			"type":        gitlabTypes.TypeSelfManaged.String(),
		}
	}

	t.Run("happy path", func(t *testing.T) {
		var sentEventType event.EventType
		var sentMetadata map[string]string

		mb := &mockConfigBackend{
			sendEvent: func(_ context.Context, eventType event.EventType, metadata map[string]string) error {
				sentEventType = eventType
				sentMetadata = metadata
				return nil
			},
		}
		p := pathConfig.New(mb)
		writeOp := p.Paths()[0].Operations[logical.UpdateOperation].Handler()

		mc := &mockGitlabClient{
			tokenInfo: testTokenInfo(),
			metadata:  testMetadata(),
		}
		ctx := gitlab.ClientNewContext(t.Context(), mc)

		fd := &framework.FieldData{Raw: validRaw(), Schema: configPath.Fields}
		resp, err := writeOp(ctx, &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())

		assert.Equal(t, "https://gitlab.example.com", resp.Data["base_url"])
		assert.Equal(t, gitlabTypes.TypeSelfManaged.String(), resp.Data["type"])
		assert.NotEmpty(t, resp.Data["token_sha1_hash"])

		assert.Equal(t, "config-write", sentEventType.String())
		assert.Equal(t, "config/default", sentMetadata["path"])
		assert.Equal(t, "default", sentMetadata["config_name"])
		assert.Equal(t, "https://gitlab.example.com", sentMetadata["base_url"])

		assert.Equal(t, "default", mb.deleteClientName)
		require.NotNil(t, mb.savedConfig)
		assert.Equal(t, "default", mb.savedConfig.Name)
		assert.Equal(t, "glpat-test-token", mb.savedConfig.Token)
	})

	t.Run("missing required fields", func(t *testing.T) {
		mb := &mockConfigBackend{}
		p := pathConfig.New(mb)
		writeOp := p.Paths()[0].Operations[logical.UpdateOperation].Handler()

		fd := &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default"},
			Schema: configPath.Fields,
		}
		resp, err := writeOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid token", func(t *testing.T) {
		mb := &mockConfigBackend{}
		p := pathConfig.New(mb)
		writeOp := p.Paths()[0].Operations[logical.UpdateOperation].Handler()

		mc := &mockGitlabClient{
			tokenInfoErr: errors.New("unauthorized"),
			metadata:     testMetadata(),
		}
		ctx := gitlab.ClientNewContext(t.Context(), mc)

		fd := &framework.FieldData{Raw: validRaw(), Schema: configPath.Fields}
		resp, err := writeOp(ctx, &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("SaveConfig fails", func(t *testing.T) {
		mb := &mockConfigBackend{
			saveErr: errors.New("save failed"),
		}
		p := pathConfig.New(mb)
		writeOp := p.Paths()[0].Operations[logical.UpdateOperation].Handler()

		mc := &mockGitlabClient{
			tokenInfo: testTokenInfo(),
			metadata:  testMetadata(),
		}
		ctx := gitlab.ClientNewContext(t.Context(), mc)

		fd := &framework.FieldData{Raw: validRaw(), Schema: configPath.Fields}
		resp, err := writeOp(ctx, &logical.Request{Storage: &logical.InmemStorage{}}, fd)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Empty(t, mb.deleteClientName)
	})
}
