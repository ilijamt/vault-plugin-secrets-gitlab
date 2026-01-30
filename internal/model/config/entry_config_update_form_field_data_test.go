package config_test

import (
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestEntryConfigUpdateFromFieldData(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		e := new(config.EntryConfig)
		_, err := e.UpdateFromFieldData(nil)
		require.ErrorIs(t, err, errs.ErrNilValue)
	})

	var tests = []struct {
		name           string
		raw            map[string]interface{}
		expectedConfig *config.EntryConfig
		warnings       []string
		err            bool
		errMap         map[string]int
	}{
		{
			name:     "no data should fail",
			raw:      map[string]interface{}{},
			err:      true,
			warnings: []string{"auto_rotate_token not specified setting to 24h0m0s"},
			errMap: map[string]int{
				errs.ErrFieldRequired.Error(): 3,
			},
		},
		{
			name: "empty token and invalid type",
			raw: map[string]interface{}{
				"base_url": "https://gitlab.com",
				"type":     "type",
			},
			expectedConfig: &config.EntryConfig{AutoRotateBefore: config.DefaultAutoRotateBeforeMinTTL, BaseURL: "https://gitlab.com"},
			warnings:       []string{"auto_rotate_token not specified setting to 24h0m0s"},
			err:            true,
			errMap: map[string]int{
				errs.ErrFieldRequired.Error():  1,
				gitlab2.ErrUnknownType.Error(): 1,
			},
		},
		{
			name:           "unconvertible data type",
			expectedConfig: &config.EntryConfig{},
			raw: map[string]interface{}{
				"token": struct{}{},
			},
			err:    true,
			errMap: map[string]int{},
		},
		{
			name: "valid config",
			expectedConfig: &config.EntryConfig{
				Token:            "token",
				Type:             gitlab2.TypeSelfManaged,
				AutoRotateToken:  false,
				AutoRotateBefore: config.DefaultAutoRotateBeforeMinTTL,
				BaseURL:          "https://gitlab.com",
			},
			warnings: []string{"auto_rotate_token not specified setting to 24h0m0s"},
			raw: map[string]interface{}{
				"token":    "token",
				"type":     gitlab2.TypeSelfManaged.String(),
				"base_url": "https://gitlab.com",
			},
		},
		{
			name: "auto_rotate_before specified (valid) should not warn and should set duration",
			expectedConfig: &config.EntryConfig{
				Token:            "token",
				Type:             gitlab2.TypeSelfManaged,
				AutoRotateToken:  false,
				AutoRotateBefore: config.DefaultAutoRotateBeforeMinTTL,
				BaseURL:          "https://gitlab.com",
			},
			warnings: nil,
			raw: map[string]interface{}{
				"token":              "token",
				"type":               gitlab2.TypeSelfManaged.String(),
				"base_url":           "https://gitlab.com",
				"auto_rotate_before": int(config.DefaultAutoRotateBeforeMinTTL.Seconds()),
			},
		},
		{
			name: "auto_rotate_before specified (too small) should error",
			expectedConfig: &config.EntryConfig{
				Token:            "token",
				Type:             gitlab2.TypeSelfManaged,
				AutoRotateToken:  false,
				AutoRotateBefore: 0,
				BaseURL:          "https://gitlab.com",
			},
			warnings: nil,
			raw: map[string]interface{}{
				"token":              "token",
				"type":               gitlab2.TypeSelfManaged.String(),
				"base_url":           "https://gitlab.com",
				"auto_rotate_before": int(config.DefaultAutoRotateBeforeMinTTL.Seconds()) - 1,
			},
			err: true,
			errMap: map[string]int{
				errs.ErrInvalidValue.Error(): 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := new(config.EntryConfig)
			assert.Empty(t, e)
			warnings, err := e.UpdateFromFieldData(&framework.FieldData{Raw: test.raw, Schema: gitlab.FieldSchemaConfig})
			assert.Equal(t, test.warnings, warnings)
			if test.expectedConfig == nil {
				test.expectedConfig = &config.EntryConfig{AutoRotateBefore: config.DefaultAutoRotateBeforeMinTTL}
			}
			assert.EqualValues(t, test.expectedConfig, e)
			if test.err {
				assert.Error(t, err)
				if len(test.errMap) > 0 {
					assert.Equal(t, utils.CountErrByName(err.(*multierror.Error)), test.errMap)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
