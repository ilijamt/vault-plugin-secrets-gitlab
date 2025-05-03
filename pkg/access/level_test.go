//go:build unit

package access_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/pkg/access"
)

func TestAccessLevel(t *testing.T) {
	var tests = []struct {
		expected access.AccessLevel
		input    string
		err      bool
	}{
		{
			expected: access.AccessLevelOwnerPermissions,
			input:    access.AccessLevelOwnerPermissions.String(),
		},
		{
			expected: access.AccessLevelReporterPermissions,
			input:    access.AccessLevelReporterPermissions.String(),
		},
		{
			expected: access.AccessLevelMaintainerPermissions,
			input:    access.AccessLevelMaintainerPermissions.String(),
		},
		{
			expected: access.AccessLevelDeveloperPermissions,
			input:    access.AccessLevelDeveloperPermissions.String(),
		},
		{
			expected: access.AccessLevelGuestPermissions,
			input:    access.AccessLevelGuestPermissions.String(),
		},
		{
			expected: access.AccessLevelNoPermissions,
			input:    access.AccessLevelNoPermissions.String(),
		},
		{
			expected: access.AccessLevelMinimalAccessPermissions,
			input:    access.AccessLevelMinimalAccessPermissions.String(),
		},
		{
			expected: access.AccessLevelUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := access.AccessLevelParse(test.input)
		assert.EqualValues(t, test.expected, val)
		if test.err {
			assert.ErrorIs(t, err, access.ErrUnknownAccessLevel)
			assert.Less(t, val.Value(), 0)
		} else {
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, val.Value(), 0)
		}
	}
}
