//go:build unit

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
)

func TestType(t *testing.T) {
	var tests = []struct {
		expected types.Type
		input    string
		err      bool
	}{
		{
			expected: types.TypeSaaS,
			input:    types.TypeSaaS.String(),
			err:      false,
		},
		{
			expected: types.TypeSelfManaged,
			input:    types.TypeSelfManaged.String(),
			err:      false,
		},
		{
			expected: types.TypeDedicated,
			input:    types.TypeDedicated.String(),
			err:      false,
		},
		{
			expected: types.TypeUnknown,
			input:    types.TypeUnknown.String(),
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := types.TypeParse(test.input)
		if test.err {
			assert.ErrorIs(t, err, types.ErrUnknownType)
		} else {
			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, val)
			assert.EqualValues(t, test.expected.Value(), test.expected.String())
		}
	}
}
