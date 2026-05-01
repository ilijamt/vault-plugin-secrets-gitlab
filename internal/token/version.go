package token

import (
	"strings"

	"golang.org/x/mod/semver"
)

// atLeast reports whether version >= minVersion using golang.org/x/mod/semver.
// Both inputs are normalized to a canonical "vMAJOR.MINOR.PATCH" before
// comparison so GitLab build flavours ("17.7", "v17.7.0", "18.0.0-ee",
// "17.9.1-pre") all parse correctly.
//
// Lenient (returns true) when:
//   - version is empty (config.GitlabVersion not yet populated on first-write),
//   - minVersion is empty (no lower bound on this value),
//   - version is not a parseable semver (e.g. an in-memory test stub returning
//     the literal "version"). Refusing every value because the upstream
//     reported a string we can't parse would silently break existing roles;
//     GitLab itself rejects unsupported values at token-create time.
func atLeast(version, minVersion string) bool {
	if version == "" || minVersion == "" {
		return true
	}
	a, b := canonicalize(version), canonicalize(minVersion)
	if !semver.IsValid(a) {
		return true
	}
	if !semver.IsValid(b) {
		return false
	}
	return semver.Compare(a, b) >= 0
}

// canonicalize coerces a GitLab version string ("17.7", "17.7.0-ee",
// "v18.0.1-pre") into a form semver.Compare accepts: a leading "v" and at
// least MAJOR.MINOR.PATCH. Pre-release and build suffixes pass through
// unchanged so semver's precedence rules apply.
func canonicalize(v string) string {
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	// Find the boundary between the dotted numeric core and any
	// pre-release / build suffix so we can pad missing components.
	core, suffix := v, ""
	for i := 1; i < len(v); i++ {
		c := v[i]
		if (c >= '0' && c <= '9') || c == '.' {
			continue
		}
		core, suffix = v[:i], v[i:]
		break
	}
	dots := strings.Count(core, ".")
	for ; dots < 2; dots++ {
		core += ".0"
	}
	return core + suffix
}
