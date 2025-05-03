package access

import (
	"errors"
	"fmt"
	"slices"

	g "gitlab.com/gitlab-org/api/client-go"
)

type AccessLevel string

const (
	AccessLevelNoPermissions            = AccessLevel("no_permissions")
	AccessLevelMinimalAccessPermissions = AccessLevel("minimal_access")
	AccessLevelGuestPermissions         = AccessLevel("guest")
	AccessLevelReporterPermissions      = AccessLevel("reporter")
	AccessLevelDeveloperPermissions     = AccessLevel("developer")
	AccessLevelMaintainerPermissions    = AccessLevel("maintainer")
	AccessLevelOwnerPermissions         = AccessLevel("owner")

	AccessLevelUnknown = AccessLevel("")
)

var (
	ErrUnknownAccessLevel = errors.New("unknown access level")

	ValidAccessLevels = []string{
		AccessLevelNoPermissions.String(),
		AccessLevelMinimalAccessPermissions.String(),
		AccessLevelGuestPermissions.String(),
		AccessLevelReporterPermissions.String(),
		AccessLevelDeveloperPermissions.String(),
		AccessLevelMaintainerPermissions.String(),
		AccessLevelOwnerPermissions.String(),
	}
	ValidPersonalAccessLevels = []string{
		AccessLevelUnknown.String(),
	}
	ValidUserServiceAccountAccessLevels = []string{
		AccessLevelUnknown.String(),
	}
	ValidGroupServiceAccountAccessLevels = []string{
		AccessLevelUnknown.String(),
	}
	ValidProjectAccessLevels = []string{
		AccessLevelGuestPermissions.String(),
		AccessLevelReporterPermissions.String(),
		AccessLevelDeveloperPermissions.String(),
		AccessLevelMaintainerPermissions.String(),
		AccessLevelOwnerPermissions.String(),
	}
	ValidGroupAccessLevels = []string{
		AccessLevelGuestPermissions.String(),
		AccessLevelReporterPermissions.String(),
		AccessLevelDeveloperPermissions.String(),
		AccessLevelMaintainerPermissions.String(),
		AccessLevelOwnerPermissions.String(),
	}

	ValidPipelineProjectTriggerAccessLevels = []string{AccessLevelUnknown.String()}
	ValidProjectDeployAccessLevels          = []string{AccessLevelUnknown.String()}
	ValidGroupDeployAccessLevels            = []string{AccessLevelUnknown.String()}
)

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
	case AccessLevelReporterPermissions:
		return int(g.ReporterPermissions)
	case AccessLevelDeveloperPermissions:
		return int(g.DeveloperPermissions)
	case AccessLevelMaintainerPermissions:
		return int(g.MaintainerPermissions)
	case AccessLevelOwnerPermissions:
		return int(g.OwnerPermissions)
	}

	return -1
}

func AccessLevelParse(value string) (AccessLevel, error) {
	if slices.Contains(ValidAccessLevels, value) {
		return AccessLevel(value), nil
	}
	return AccessLevelUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownAccessLevel)
}
