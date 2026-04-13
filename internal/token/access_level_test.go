package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

func TestAccessLevel(t *testing.T) {
	var tests = []struct {
		name     string
		expected gitlab.AccessLevel
		input    string
		err      bool
	}{
		{
			name:     "owner",
			expected: gitlab.AccessLevelOwnerPermissions,
			input:    gitlab.AccessLevelOwnerPermissions.String(),
		},
		{
			name:     "reporter",
			expected: gitlab.AccessLevelReporterPermissions,
			input:    gitlab.AccessLevelReporterPermissions.String(),
		},
		{
			name:     "maintainer",
			expected: gitlab.AccessLevelMaintainerPermissions,
			input:    gitlab.AccessLevelMaintainerPermissions.String(),
		},
		{
			name:     "developer",
			expected: gitlab.AccessLevelDeveloperPermissions,
			input:    gitlab.AccessLevelDeveloperPermissions.String(),
		},
		{
			name:     "guest",
			expected: gitlab.AccessLevelGuestPermissions,
			input:    gitlab.AccessLevelGuestPermissions.String(),
		},
		{
			name:     "no_permissions",
			expected: gitlab.AccessLevelNoPermissions,
			input:    gitlab.AccessLevelNoPermissions.String(),
		},
		{
			name:     "minimal_access",
			expected: gitlab.AccessLevelMinimalAccessPermissions,
			input:    gitlab.AccessLevelMinimalAccessPermissions.String(),
		},
		{
			name:     "unknown",
			expected: gitlab.AccessLevelUnknown,
			input:    "unknown",
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := gitlab.ParseAccessLevel(test.input)
			assert.EqualValues(t, test.expected, val)
			if test.err {
				assert.ErrorIs(t, err, errs.ErrUnknownAccessLevel)
				assert.Less(t, val.Value(), 0)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, val.Value(), 0)
			}
		})
	}
}
