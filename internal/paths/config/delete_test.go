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
	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestPathConfigDelete(t *testing.T) {
	configPath := pathConfig.New(&mockConfigBackend{}).Paths()[0]

	newFieldData := func() *framework.FieldData {
		return &framework.FieldData{
			Raw:    map[string]interface{}{"config_name": "default"},
			Schema: configPath.Fields,
		}
	}

	t.Run("config not found", func(t *testing.T) {
		mb := &mockConfigBackend{config: nil}
		p := pathConfig.New(mb)
		deleteOp := p.Paths()[0].Operations[logical.DeleteOperation].Handler()

		resp, err := deleteOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsError())
		assert.Contains(t, resp.Error().Error(), errs.ErrBackendNotConfigured.Error())
	})

	t.Run("GetConfig error", func(t *testing.T) {
		mb := &mockConfigBackend{configErr: errors.New("storage failure")}
		p := pathConfig.New(mb)
		deleteOp := p.Paths()[0].Operations[logical.DeleteOperation].Handler()

		resp, err := deleteOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Empty(t, mb.deleteClientName)
	})

	t.Run("happy path", func(t *testing.T) {
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
		deleteOp := p.Paths()[0].Operations[logical.DeleteOperation].Handler()

		s := &logical.InmemStorage{}
		require.NoError(t, s.Put(t.Context(), &logical.StorageEntry{
			Key:   "config/default",
			Value: []byte("{}"),
		}))

		resp, err := deleteOp(t.Context(), &logical.Request{Storage: s}, newFieldData())
		require.NoError(t, err)
		assert.Nil(t, resp)

		assert.Equal(t, "config-delete", sentEventType.String())
		assert.Equal(t, "config/default", sentMetadata["path"])
		assert.Equal(t, "default", mb.deleteClientName)

		entry, err := s.Get(t.Context(), "config/default")
		require.NoError(t, err)
		assert.Nil(t, entry)
	})
}
