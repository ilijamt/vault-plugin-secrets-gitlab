package config_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	pathConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/config"
)

func TestPathConfigRead(t *testing.T) {
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
		readOp := p.Paths()[0].Operations[logical.ReadOperation].Handler()

		resp, err := readOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsError())
		assert.Contains(t, resp.Error().Error(), errs.ErrBackendNotConfigured.Error())
	})

	t.Run("GetConfig error", func(t *testing.T) {
		mb := &mockConfigBackend{configErr: errors.New("storage failure")}
		p := pathConfig.New(mb)
		readOp := p.Paths()[0].Operations[logical.ReadOperation].Handler()

		resp, err := readOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "storage failure")
	})

	t.Run("config found without ShowConfigToken", func(t *testing.T) {
		cfg := testConfig()
		mb := &mockConfigBackend{config: cfg}
		p := pathConfig.New(mb)
		readOp := p.Paths()[0].Operations[logical.ReadOperation].Handler()

		resp, err := readOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())
		assert.NotEmpty(t, resp.Data["token_sha1_hash"])
		assert.NotContains(t, resp.Data, "token")
		assert.Equal(t, cfg.BaseURL, resp.Data["base_url"])
		assert.Equal(t, cfg.Type.String(), resp.Data["type"])
		assert.Equal(t, cfg.Name, resp.Data["name"])
	})

	t.Run("config found with ShowConfigToken", func(t *testing.T) {
		cfg := testConfig()
		mb := &mockConfigBackend{
			config:   cfg,
			flagsVal: flags.Flags{ShowConfigToken: true},
		}
		p := pathConfig.New(mb)
		readOp := p.Paths()[0].Operations[logical.ReadOperation].Handler()

		resp, err := readOp(t.Context(), &logical.Request{Storage: &logical.InmemStorage{}}, newFieldData())
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsError())
		assert.Equal(t, cfg.Token, resp.Data["token"])
		assert.NotEmpty(t, resp.Data["token_sha1_hash"])
	})
}
