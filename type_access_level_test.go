package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestAccessLevel(t *testing.T) {
	var tests = []struct {
		expected gitlab.AccessLevel
		input    string
		err      bool
	}{
		{
			expected: gitlab.AccessLevelOwnerPermissions,
			input:    gitlab.AccessLevelOwnerPermissions.String(),
		},
		{
			expected: gitlab.AccessLevelReporterPermissions,
			input:    gitlab.AccessLevelReporterPermissions.String(),
		},
		{
			expected: gitlab.AccessLevelMaintainerPermissions,
			input:    gitlab.AccessLevelMaintainerPermissions.String(),
		},
		{
			expected: gitlab.AccessLevelDeveloperPermissions,
			input:    gitlab.AccessLevelDeveloperPermissions.String(),
		},
		{
			expected: gitlab.AccessLevelGuestPermissions,
			input:    gitlab.AccessLevelGuestPermissions.String(),
		},
		{
			expected: gitlab.AccessLevelNoPermissions,
			input:    gitlab.AccessLevelNoPermissions.String(),
		},
		{
			expected: gitlab.AccessLevelMinimalAccessPermissions,
			input:    gitlab.AccessLevelMinimalAccessPermissions.String(),
		},
		{
			expected: gitlab.AccessLevelUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Logf("assert parse(%s) = %s (err: %v)", test.input, test.expected, test.err)
		val, err := gitlab.AccessLevelParse(test.input)
		assert.EqualValues(t, test.expected, val)
		if test.err {
			assert.ErrorIs(t, err, gitlab.ErrUnknownAccessLevel)
			assert.Less(t, val.Value(), 0)
		} else {
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, val.Value(), 0)
		}
	}
}
