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
			name:     "planner",
			expected: gitlab.AccessLevelPlannerPermissions,
			input:    gitlab.AccessLevelPlannerPermissions.String(),
		},
		{
			name:     "security_manager",
			expected: gitlab.AccessLevelSecurityManagerPermissions,
			input:    gitlab.AccessLevelSecurityManagerPermissions.String(),
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

func TestValidAccessLevelsFor_Applicability(t *testing.T) {
	notApplicable := []gitlab.Type{
		gitlab.TypePersonal,
		gitlab.TypeUserServiceAccount,
		gitlab.TypeGroupServiceAccount,
		gitlab.TypePipelineProjectTrigger,
		gitlab.TypeProjectDeploy,
		gitlab.TypeGroupDeploy,
	}
	for _, tt := range notApplicable {
		t.Run(string(tt)+" not applicable", func(t *testing.T) {
			levels, applicable := gitlab.ValidAccessLevelsFor(tt, "18.0")
			assert.False(t, applicable)
			assert.Empty(t, levels)
		})
	}

	t.Run("project applicable", func(t *testing.T) {
		_, applicable := gitlab.ValidAccessLevelsFor(gitlab.TypeProject, "17.0")
		assert.True(t, applicable)
	})
}

func TestIsAccessLevelAllowed_VersionGating(t *testing.T) {
	tests := []struct {
		name      string
		tokenType gitlab.Type
		level     gitlab.AccessLevel
		version   string
		want      bool
	}{
		{"planner gated below 17.7", gitlab.TypeGroup, gitlab.AccessLevelPlannerPermissions, "17.0", false},
		{"planner allowed at 17.7", gitlab.TypeGroup, gitlab.AccessLevelPlannerPermissions, "17.7", true},
		{"security_manager gated below 18.11", gitlab.TypeProject, gitlab.AccessLevelSecurityManagerPermissions, "18.10", false},
		{"security_manager allowed at 18.11", gitlab.TypeProject, gitlab.AccessLevelSecurityManagerPermissions, "18.11", true},
		{"security_manager rejected on personal", gitlab.TypePersonal, gitlab.AccessLevelSecurityManagerPermissions, "18.11", false},
		{"planner allowed on project at 18.0", gitlab.TypeProject, gitlab.AccessLevelPlannerPermissions, "18.0", true},
		{"planner lenient on empty version", gitlab.TypeGroup, gitlab.AccessLevelPlannerPermissions, "", true},
		{"maintainer always on group", gitlab.TypeGroup, gitlab.AccessLevelMaintainerPermissions, "17.0", true},
		{"maintainer rejected on personal", gitlab.TypePersonal, gitlab.AccessLevelMaintainerPermissions, "18.0", false},
		{"unknown level rejected", gitlab.TypeGroup, gitlab.AccessLevelUnknown, "18.0", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, gitlab.IsAccessLevelAllowed(tc.tokenType, tc.level, tc.version))
		})
	}
}

func TestAllValidAccessLevels(t *testing.T) {
	all := gitlab.AllValidAccessLevels()
	assert.Contains(t, all, gitlab.AccessLevelMaintainerPermissions.String())
	assert.Contains(t, all, gitlab.AccessLevelPlannerPermissions.String())
	assert.NotContains(t, all, gitlab.AccessLevelUnknown.String())
}
