//go:build !integration

package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestType(t *testing.T) {
	var tests = []struct {
		expected gitlab.Type
		input    string
		err      bool
	}{
		{
			expected: gitlab.TypeSaaS,
			input:    gitlab.TypeSaaS.String(),
			err:      false,
		},
		{
			expected: gitlab.TypeSelfManaged,
			input:    gitlab.TypeSelfManaged.String(),
			err:      false,
		},
		{
			expected: gitlab.TypeDedicated,
			input:    gitlab.TypeDedicated.String(),
			err:      true,
		},
		{
			expected: gitlab.TypeUnknown,
			input:    gitlab.TypeUnknown.String(),
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := gitlab.TypeParse(test.input)
		if test.err {
			assert.ErrorIs(t, err, gitlab.ErrUnknownType)
		} else {
			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, val)
			assert.EqualValues(t, test.expected.Value(), test.expected.String())
		}
	}
}
