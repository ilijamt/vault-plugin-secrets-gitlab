//go:build unit

package gitlab_test

import (
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestEntryConfigMerge(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		e := new(gitlab.EntryConfig)
		warnings, changes, err := e.Merge(nil)
		require.Empty(t, warnings)
		require.Empty(t, changes)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})

	t.Run("unconvertible data type", func(t *testing.T) {
		e := new(gitlab.EntryConfig)
		warnings, changes, err := e.Merge(&framework.FieldData{
			Raw:    map[string]interface{}{"token": struct{}{}},
			Schema: gitlab.FieldSchemaConfig,
		})
		require.Empty(t, warnings)
		require.Empty(t, changes)
		require.ErrorContains(t, err, "got unconvertible type")
	})

	var tests = []struct {
		name           string
		originalConfig *gitlab.EntryConfig
		expectedConfig *gitlab.EntryConfig
		raw            map[string]interface{}
		warnings       []string
		changes        map[string]string
		err            bool
		errMap         map[string]int
	}{
		{
			name:           "update type only",
			originalConfig: &gitlab.EntryConfig{Type: gitlab.TypeSelfManaged},
			expectedConfig: &gitlab.EntryConfig{Type: gitlab.TypeSaaS},
			raw:            map[string]interface{}{"type": gitlab.TypeSaaS},
			changes:        map[string]string{"type": gitlab.TypeSaaS.String()},
		},
		{
			name:           "auto rotate token set to false",
			originalConfig: &gitlab.EntryConfig{},
			expectedConfig: &gitlab.EntryConfig{},
			raw:            map[string]interface{}{"auto_rotate_token": false},
			changes:        map[string]string{"auto_rotate_token": "false"},
		},
		{
			name:           "auto rotate token set to true",
			originalConfig: &gitlab.EntryConfig{AutoRotateToken: false},
			expectedConfig: &gitlab.EntryConfig{AutoRotateToken: true},
			raw:            map[string]interface{}{"auto_rotate_token": true},
			changes:        map[string]string{"auto_rotate_token": "true"},
		},
		{
			name:           "update type with invalid type",
			originalConfig: &gitlab.EntryConfig{Type: gitlab.TypeSelfManaged},
			expectedConfig: &gitlab.EntryConfig{Type: gitlab.TypeSelfManaged},
			raw:            map[string]interface{}{"type": "test"},
			err:            true,
			errMap: map[string]int{
				gitlab.ErrUnknownType.Error(): 1,
			},
		},
		{
			name:           "set base url to a non empty value",
			originalConfig: &gitlab.EntryConfig{},
			expectedConfig: &gitlab.EntryConfig{BaseURL: "https://gitlab.com/"},
			raw:            map[string]interface{}{"base_url": "https://gitlab.com/"},
			changes:        map[string]string{"base_url": "https://gitlab.com/"},
		},
		{
			name:           "set base url to an empty value should fail",
			originalConfig: &gitlab.EntryConfig{BaseURL: "https://gitlab.com/"},
			expectedConfig: &gitlab.EntryConfig{BaseURL: "https://gitlab.com/"},
			raw:            map[string]interface{}{"base_url": ""},
		},

		{
			name:           "auto rotate before invalid value lower than min",
			originalConfig: &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL + time.Hour},
			expectedConfig: &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL + time.Hour},
			raw:            map[string]interface{}{"auto_rotate_before": "1h"},
			err:            true,
			errMap:         map[string]int{gitlab.ErrInvalidValue.Error(): 1},
		},
		{
			name:           "auto rotate before invalid value higher than min",
			originalConfig: &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL + time.Hour},
			expectedConfig: &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL + time.Hour},
			raw:            map[string]interface{}{"auto_rotate_before": (gitlab.DefaultAutoRotateBeforeMaxTTL + time.Hour).String()},
			err:            true,
			errMap:         map[string]int{gitlab.ErrInvalidValue.Error(): 1},
		},
		{
			name:           "auto rotate with a valid value",
			originalConfig: &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL + time.Hour},
			expectedConfig: &gitlab.EntryConfig{AutoRotateBefore: gitlab.DefaultAutoRotateBeforeMinTTL + time.Hour*2},
			raw:            map[string]interface{}{"auto_rotate_before": (gitlab.DefaultAutoRotateBeforeMinTTL + time.Hour*2).String()},
			err:            false,
			changes:        map[string]string{"auto_rotate_before": "26h0m0s"},
		},
		{
			name:           "token a valid value",
			originalConfig: &gitlab.EntryConfig{Token: "token1"},
			expectedConfig: &gitlab.EntryConfig{Token: "token"},
			raw:            map[string]interface{}{"token": "token"},
			err:            false,
			changes:        map[string]string{"token": "*****"},
		},
		{
			name:           "token an empty value",
			originalConfig: &gitlab.EntryConfig{Token: "token"},
			expectedConfig: &gitlab.EntryConfig{Token: "token"},
			raw:            map[string]interface{}{"token": ""},
			err:            false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			warnings, changes, err := test.originalConfig.Merge(&framework.FieldData{
				Raw:    test.raw,
				Schema: gitlab.FieldSchemaConfig,
			})
			assert.EqualValues(t, test.warnings, warnings)
			if test.changes == nil {
				test.changes = make(map[string]string)
			}
			assert.EqualValues(t, test.changes, changes)
			assert.EqualValues(t, test.expectedConfig, test.originalConfig)
			if test.err {
				assert.Error(t, err)
				if len(test.errMap) > 0 {
					assert.EqualValues(t, countErrByName(err.(*multierror.Error)), test.errMap)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
