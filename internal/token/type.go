package token

import (
	"fmt"
	"slices"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

type Type string

const (
	TypePersonal               = Type("personal")
	TypeProject                = Type("project")
	TypeGroup                  = Type("group")
	TypeUserServiceAccount     = Type("user-service-account")
	TypeGroupServiceAccount    = Type("group-service-account")
	TypeProjectServiceAccount  = Type("project-service-account")
	TypePipelineProjectTrigger = Type("pipeline-project-trigger")
	TypeProjectDeploy          = Type("project-deploy")
	TypeGroupDeploy            = Type("group-deploy")

	TypeUnknown = Type("")
)

var (
	ValidTokenTypes = []string{
		TypePersonal.String(),
		TypeProject.String(),
		TypeGroup.String(),
		TypeUserServiceAccount.String(),
		TypeGroupServiceAccount.String(),
		TypeProjectServiceAccount.String(),
		TypePipelineProjectTrigger.String(),
		TypeProjectDeploy.String(),
		TypeGroupDeploy.String(),
	}
)

var typeMinVersionByTokenType = map[Type]string{
	TypePersonal:               "0.0",
	TypeProject:                "13.10",
	TypeGroup:                  "14.7",
	TypeUserServiceAccount:     "16.1",
	TypeGroupServiceAccount:    "16.1",
	TypeProjectServiceAccount:  "18.11",
	TypePipelineProjectTrigger: "0.0",
	TypeProjectDeploy:          "12.9",
	TypeGroupDeploy:            "12.9",
}

// MinVersionForType returns the minimum GitLab MAJOR.MINOR version supporting
// tokenType. ok is false for an unknown token type.
func MinVersionForType(tokenType Type) (minVersion string, ok bool) {
	minVersion, ok = typeMinVersionByTokenType[tokenType]
	return minVersion, ok
}

// IsTokenTypeAllowed reports whether tokenType is available on gitlabVersion.
// An empty or unparseable version is allowed, as with scopes and access levels.
func IsTokenTypeAllowed(tokenType Type, gitlabVersion string) bool {
	minVersion, ok := typeMinVersionByTokenType[tokenType]
	return ok && atLeast(gitlabVersion, minVersion)
}

func (i Type) String() string {
	return string(i)
}

func ParseType(value string) (Type, error) {
	if slices.Contains(ValidTokenTypes, value) {
		return Type(value), nil
	}
	return TypeUnknown, fmt.Errorf("failed to parse '%s': %w", value, errs.ErrUnknownTokenType)
}
