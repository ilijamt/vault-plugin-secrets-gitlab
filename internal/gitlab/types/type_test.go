package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
)

func TestType(t *testing.T) {
	var tests = []struct {
		name     string
		expected types.Type
		input    string
		err      bool
	}{
		{
			name:     "saas",
			expected: types.TypeSaaS,
			input:    types.TypeSaaS.String(),
		},
		{
			name:     "self-managed",
			expected: types.TypeSelfManaged,
			input:    types.TypeSelfManaged.String(),
		},
		{
			name:     "dedicated",
			expected: types.TypeDedicated,
			input:    types.TypeDedicated.String(),
		},
		{
			name:     "unknown",
			expected: types.TypeUnknown,
			input:    types.TypeUnknown.String(),
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := types.TypeParse(test.input)
			if test.err {
				assert.ErrorIs(t, err, types.ErrUnknownType)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, test.expected, val)
				assert.EqualValues(t, test.expected.Value(), test.expected.String())
			}
		})
	}
}
