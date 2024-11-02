package gitlab_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestEntryRoleMerge(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		e := new(gitlab.EntryRole)
		warnings, changes, err := e.Merge(nil, gitlab.TypeUnknown)
		require.Empty(t, warnings)
		require.Empty(t, changes)
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})

	t.Run("unconvertible data type", func(t *testing.T) {
		e := new(gitlab.EntryRole)
		warnings, changes, err := e.Merge(&framework.FieldData{
			Raw:    map[string]interface{}{"token": struct{}{}},
			Schema: gitlab.FieldSchemaConfig,
		}, gitlab.TypeUnknown)
		require.Empty(t, warnings)
		require.Empty(t, changes)
		require.ErrorContains(t, err, "got unconvertible type")
	})

	// var tests = []struct {
	// 	name           string
	// 	originalConfig *gitlab.EntryConfig
	// 	expectedConfig *gitlab.EntryConfig
	// 	raw            map[string]interface{}
	// 	warnings       []string
	// 	changes        map[string]string
	// 	err            bool
	// 	errMap         map[string]int
	// }{}

	// for _, test := range tests {
	// 	t.Run(test.name, func(t *testing.T) {
	// 		warnings, changes, err := test.originalConfig.Merge(&framework.FieldData{
	// 			Raw:    test.raw,
	// 			Schema: gitlab.FieldSchemaRoles,
	// 		})
	// 		assert.EqualValues(t, test.warnings, warnings)
	// 		if test.changes == nil {
	// 			test.changes = make(map[string]string)
	// 		}
	// 		assert.EqualValues(t, test.changes, changes)
	// 		assert.EqualValues(t, test.expectedConfig, test.originalConfig)
	// 		if test.err {
	// 			assert.Error(t, err)
	// 			if len(test.errMap) > 0 {
	// 				assert.EqualValues(t, countErrByName(err.(*multierror.Error)), test.errMap)
	// 			}
	// 		} else {
	// 			assert.NoError(t, err)
	// 		}
	// 	})
	// }
}
