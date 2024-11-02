package gitlab_test

import (
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/stretchr/testify/require"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestEntryRoleUpdateFromFieldData(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		e := new(gitlab.EntryRole)
		_, err := e.UpdateFromFieldData(nil, "")
		require.ErrorIs(t, err, gitlab.ErrNilValue)
	})

	var tests = []struct {
		name     string
		raw      map[string]interface{}
		expected *gitlab.EntryRole
		warnings []string
		typ      gitlab.Type
		err      bool
		errMap   map[string]int
	}{
		{
			name:     "no data should fail",
			raw:      map[string]interface{}{},
			err:      true,
			expected: &gitlab.EntryRole{ConfigName: gitlab.DefaultConfigName},
			errMap: map[string]int{
				gitlab.ErrFieldRequired.Error(): 3,
			},
		},
		{
			name:     "unconvertible data type",
			expected: &gitlab.EntryRole{},
			raw: map[string]interface{}{
				"name": struct{}{},
			},
			err:    true,
			errMap: map[string]int{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := new(gitlab.EntryRole)
			require.Empty(t, e)
			warnings, err := e.UpdateFromFieldData(&framework.FieldData{Raw: test.raw, Schema: gitlab.FieldSchemaRoles}, test.typ)
			require.Equal(t, test.warnings, warnings)
			if test.expected == nil {
				test.expected = &gitlab.EntryRole{}
			}
			require.EqualValues(t, test.expected, e)
			if test.err {
				require.Error(t, err)
				if len(test.errMap) > 0 {
					require.Equal(t, countErrByName(err.(*multierror.Error)), test.errMap)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
