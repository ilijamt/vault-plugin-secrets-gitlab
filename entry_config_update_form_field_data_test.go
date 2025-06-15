//go:build unit

package gitlab_test

import (
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	gitlab2 "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

func TestEntryConfigUpdateFromFieldData(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		e := new(gitlab.EntryConfig)
		_, err := e.UpdateFromFieldData(nil)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})

	var tests = []struct {
		name           string
		raw            map[string]interface{}
		expectedConfig *gitlab.EntryConfig
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
				gitlab.ErrFieldRequired.Error(): 3,
			},
		},
		{
			name: "empty token and invalid type",
			raw: map[string]interface{}{
				"base_url": "https://gitlab.com",
				"type":     "type",
			},
			expectedConfig: &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL, BaseURL: "https://gitlab.com"},
			warnings:       []string{"auto_rotate_token not specified setting to 24h0m0s"},
			err:            true,
			errMap: map[string]int{
				gitlab.ErrFieldRequired.Error(): 1,
				gitlab2.ErrUnknownType.Error():  1,
			},
		},
		{
			name:           "unconvertible data type",
			expectedConfig: &gitlab.EntryConfig{},
			raw: map[string]interface{}{
				"token": struct{}{},
			},
			err:    true,
			errMap: map[string]int{},
		},
		{
			name: "valid config",
			expectedConfig: &gitlab.EntryConfig{
				Token:            "token",
				Type:             gitlab2.TypeSelfManaged,
				AutoRotateToken:  false,
				AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL,
				BaseURL:          "https://gitlab.com",
			},
			warnings: []string{"auto_rotate_token not specified setting to 24h0m0s"},
			raw: map[string]interface{}{
				"token":    "token",
				"type":     gitlab2.TypeSelfManaged.String(),
				"base_url": "https://gitlab.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := new(gitlab.EntryConfig)
			assert.Empty(t, e)
			warnings, err := e.UpdateFromFieldData(&framework.FieldData{Raw: test.raw, Schema: gitlab.FieldSchemaConfig})
			assert.Equal(t, test.warnings, warnings)
			if test.expectedConfig == nil {
				test.expectedConfig = &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL}
			}
			assert.EqualValues(t, test.expectedConfig, e)
			if test.err {
				assert.Error(t, err)
				if len(test.errMap) > 0 {
					assert.Equal(t, countErrByName(err.(*multierror.Error)), test.errMap)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
