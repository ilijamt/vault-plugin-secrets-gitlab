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
		TypePipelineProjectTrigger.String(),
		TypeProjectDeploy.String(),
		TypeGroupDeploy.String(),
	}
)

func (i Type) String() string {
	return string(i)
}

func (i Type) Value() string {
	return i.String()
}

func ParseType(value string) (Type, error) {
	if slices.Contains(ValidTokenTypes, value) {
		return Type(value), nil
	}
	return TypeUnknown, fmt.Errorf("failed to parse '%s': %w", value, errs.ErrUnknownTokenType)
}
