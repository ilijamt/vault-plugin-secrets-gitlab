package token

import (
	"fmt"
	"slices"
	"sort"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"

	g "gitlab.com/gitlab-org/api/client-go/v2"
)

type AccessLevel string

const (
	AccessLevelNoPermissions = AccessLevel("no_permissions")
	// AccessLevelMinimalAccessPermissions allows view limited group information without access to projects. For more information, see Users with Minimal Access.
	AccessLevelMinimalAccessPermissions = AccessLevel("minimal_access")
	// AccessLevelGuestPermissions allows view and comment on issues and epics. Cannot push code or access repository. This role applies to private and internal projects only.
	AccessLevelGuestPermissions = AccessLevel("guest")
	// AccessLevelPlannerPermissions allows create and manage issues, epics, milestones, and iterations. Focused on project planning and tracking with the ability to view and collaborate on code changes.
	AccessLevelPlannerPermissions = AccessLevel("planner")
	// AccessLevelReporterPermissions allows view code, create issues, and generate reports. Cannot push code or manage protected branches.
	AccessLevelReporterPermissions = AccessLevel("reporter")
	// AccessLevelSecurityManagerPermissions allows view and manage security vulnerabilities, compliance configurations, and audit events. Focused on security operations without code push access.
	AccessLevelSecurityManagerPermissions = AccessLevel("security_manager")
	// AccessLevelDeveloperPermissions allows push code to non-protected branches, create merge requests, and run CI/CD pipelines. Cannot manage project settings.
	AccessLevelDeveloperPermissions = AccessLevel("developer")
	// AccessLevelMaintainerPermissions allows manage branches, merge requests, CI/CD settings, and project members. Cannot delete the project.
	AccessLevelMaintainerPermissions = AccessLevel("maintainer")
	// AccessLevelOwnerPermissions allows full control over the project or group, including deletion and visibility settings.
	AccessLevelOwnerPermissions = AccessLevel("owner")

	AccessLevelUnknown = AccessLevel("")
)

// ValidAccessLevels is the union of every AccessLevel string the parser
// accepts. It is the parser whitelist, not a per-token-type validator —
// per-token-type and per-version gating live in accessLevelMinVersionByTokenType.
var ValidAccessLevels = []string{
	AccessLevelNoPermissions.String(),
	AccessLevelMinimalAccessPermissions.String(),
	AccessLevelGuestPermissions.String(),
	AccessLevelPlannerPermissions.String(),
	AccessLevelReporterPermissions.String(),
	AccessLevelDeveloperPermissions.String(),
	AccessLevelSecurityManagerPermissions.String(),
	AccessLevelMaintainerPermissions.String(),
	AccessLevelOwnerPermissions.String(),
}

// accessLevelMinVersionByTokenType maps a token type to the access_levels it
// supports and the GitLab MAJOR.MINOR version each became available.
//
// A token type whose value is nil indicates "access_level is not applicable
// for this token type" (callers should reject any non-empty access_level).
// "0.0" means always allowed within the supported window.
var accessLevelMinVersionByTokenType = map[Type]map[AccessLevel]string{
	TypeProject: {
		AccessLevelGuestPermissions:           "0.0",
		AccessLevelReporterPermissions:        "0.0",
		AccessLevelDeveloperPermissions:       "0.0",
		AccessLevelMaintainerPermissions:      "0.0",
		AccessLevelOwnerPermissions:           "0.0",
		AccessLevelPlannerPermissions:         "17.7",
		AccessLevelSecurityManagerPermissions: "18.11",
	},
	TypeGroup: {
		AccessLevelGuestPermissions:           "0.0",
		AccessLevelReporterPermissions:        "0.0",
		AccessLevelDeveloperPermissions:       "0.0",
		AccessLevelMaintainerPermissions:      "0.0",
		AccessLevelOwnerPermissions:           "0.0",
		AccessLevelPlannerPermissions:         "17.7",
		AccessLevelSecurityManagerPermissions: "18.11",
	},
	TypePersonal:               nil,
	TypeUserServiceAccount:     nil,
	TypeGroupServiceAccount:    nil,
	TypePipelineProjectTrigger: nil,
	TypeProjectDeploy:          nil,
	TypeGroupDeploy:            nil,
}

// ValidAccessLevelsFor returns the access_levels allowed for tokenType on the
// given GitLab version, sorted by AccessLevel.Value(). applicable is false if
// tokenType does not take an access_level field at all (e.g. personal,
// pipeline trigger, deploy tokens). When version is empty the gate is lenient
// — every level the token type accepts is returned.
func ValidAccessLevelsFor(tokenType Type, gitlabVersion string) (levels []AccessLevel, applicable bool) {
	inner, present := accessLevelMinVersionByTokenType[tokenType]
	if !present || inner == nil {
		return nil, false
	}
	for level, minV := range inner {
		if atLeast(gitlabVersion, minV) {
			levels = append(levels, level)
		}
	}
	sort.Slice(levels, func(i, j int) bool { return levels[i].Value() < levels[j].Value() })
	return levels, true
}

// IsAccessLevelAllowed reports whether level is a valid access_level for
// tokenType on gitlabVersion. Returns false if tokenType does not take an
// access_level field.
func IsAccessLevelAllowed(tokenType Type, level AccessLevel, gitlabVersion string) bool {
	inner, present := accessLevelMinVersionByTokenType[tokenType]
	if !present || inner == nil {
		return false
	}
	minV, ok := inner[level]
	if !ok {
		return false
	}
	return atLeast(gitlabVersion, minV)
}

// AllValidAccessLevels returns the union of access_levels accepted by any
// token type at any version — used to populate the OpenAPI schema's
// AllowedValues at backend startup, before a GitLab version is known.
func AllValidAccessLevels() []string {
	seen := map[AccessLevel]struct{}{}
	for _, inner := range accessLevelMinVersionByTokenType {
		for level := range inner {
			seen[level] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for level := range seen {
		out = append(out, level.String())
	}
	sort.Strings(out)
	return out
}

func (i AccessLevel) String() string {
	return string(i)
}

func (i AccessLevel) Value() int {
	switch i {
	case AccessLevelNoPermissions:
		return int(g.NoPermissions)
	case AccessLevelMinimalAccessPermissions:
		return int(g.MinimalAccessPermissions)
	case AccessLevelGuestPermissions:
		return int(g.GuestPermissions)
	case AccessLevelPlannerPermissions:
		return int(g.PlannerPermissions)
	case AccessLevelReporterPermissions:
		return int(g.ReporterPermissions)
	case AccessLevelDeveloperPermissions:
		return int(g.DeveloperPermissions)
	case AccessLevelSecurityManagerPermissions:
		// Beta in GitLab 18.11; not yet exposed as a constant in
		// gitlab.com/gitlab-org/api/client-go v1.46.0, so the integer is
		// hardcoded from the GitLab REST API documentation.
		return 25
	case AccessLevelMaintainerPermissions:
		return int(g.MaintainerPermissions)
	case AccessLevelOwnerPermissions:
		return int(g.OwnerPermissions)
	}

	return -1
}

func ParseAccessLevel(value string) (AccessLevel, error) {
	if slices.Contains(ValidAccessLevels, value) {
		return AccessLevel(value), nil
	}
	return AccessLevelUnknown, fmt.Errorf("failed to parse '%s': %w", value, errs.ErrUnknownAccessLevel)
}
