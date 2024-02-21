package gitlab

import (
	"errors"
	"fmt"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/slices"
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
)

func (i AccessLevel) String() string {
	return string(i)
}

func (i AccessLevel) Value() int {
	switch i {
	case AccessLevelNoPermissions:
		return int(gitlab.NoPermissions)
	case AccessLevelMinimalAccessPermissions:
		return int(gitlab.MinimalAccessPermissions)
	case AccessLevelGuestPermissions:
		return int(gitlab.GuestPermissions)
	case AccessLevelReporterPermissions:
		return int(gitlab.ReporterPermissions)
	case AccessLevelDeveloperPermissions:
		return int(gitlab.DeveloperPermissions)
	case AccessLevelMaintainerPermissions:
		return int(gitlab.MaintainerPermissions)
	case AccessLevelOwnerPermissions:
		return int(gitlab.OwnerPermissions)
	}

	return -1
}

func AccessLevelParse(value string) (AccessLevel, error) {
	if slices.Contains(ValidAccessLevels, value) {
		return AccessLevel(value), nil
	}
	return AccessLevelUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownAccessLevel)
}
