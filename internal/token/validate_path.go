package token

import (
	"regexp"
	"strings"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

var (
	allowedSegment      = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	invalidPathPrefixes = []string{"-", "_", "."}
	invalidPathSuffixes = []string{"-", "_", ".", ".git", ".atom"}
	invalidSegmentEdges = []string{"-", "_", "."}
)

/*
IsValidPath validates a path string for a specified tokenType.

Validation rules:
  - Each segment can contain only ASCII letters, digits, '_', '-', '.'.
  - Path must not start with '-', '_', or '.'.
  - Path must not end with '-', '_', '.', '.git' or '.atom'.
  - Segment count rules per token type:
    -- TypePersonal, TypeUserServiceAccount: exactly 1 segment.
    -- TypeGroupServiceAccount: exactly 2 segments.
    -- TypeProject, TypeGroup, TypeProjectDeploy, TypeGroupDeploy, TypePipelineProjectTrigger: 1 or more segments.

Returns true if valid, else false.
*/
func IsValidPath(path string, tokenType Type) (valid bool) {
	if strings.TrimSpace(path) == "" {
		return false
	}

	if utils.HasAny(path, invalidPathPrefixes, strings.HasPrefix) ||
		utils.HasAny(path, invalidPathSuffixes, strings.HasSuffix) {
		return false
	}

	segments := strings.Split(path, "/")
	for _, s := range segments {
		if s == "" {
			return false
		}

		if !allowedSegment.MatchString(s) ||
			utils.HasAny(s, invalidSegmentEdges, strings.HasPrefix) ||
			utils.HasAny(s, invalidSegmentEdges, strings.HasSuffix) {
			return false
		}
	}

	switch tokenType {
	case TypePersonal, TypeUserServiceAccount:
		/*
			Format of the paths:
				- {username}
		*/
		return len(segments) == 1

	case TypeGroupServiceAccount:
		/*
			Format of the paths:
				- {groupId}/{serviceAccountName}
		*/
		return len(segments) == 2

	case TypeProject, TypeGroup, TypeProjectDeploy, TypeGroupDeploy, TypePipelineProjectTrigger:
		/*
			Format of the paths:
				- group/project or group/subgroup/project
				- group or group/subgroup
		*/
		return len(segments) >= 1
	}

	return false
}
