package token

import (
	"errors"
	"fmt"
	"slices"
)

type Type string

const (
	TokenTypePersonal               = Type("personal")
	TokenTypeProject                = Type("project")
	TokenTypeGroup                  = Type("group")
	TokenTypeUserServiceAccount     = Type("user-service-account")
	TokenTypeGroupServiceAccount    = Type("group-service-account")
	TokenTypePipelineProjectTrigger = Type("pipeline-project-trigger")
	TokenTypeProjectDeploy          = Type("project-deploy")
	TokenTypeGroupDeploy            = Type("group-deploy")

	TokenTypeUnknown = Type("")
)

var (
	ErrUnknownTokenType = errors.New("unknown token type")

	ValidTokenTypes = []string{
		TokenTypePersonal.String(),
		TokenTypeProject.String(),
		TokenTypeGroup.String(),
		TokenTypeUserServiceAccount.String(),
		TokenTypeGroupServiceAccount.String(),
		TokenTypePipelineProjectTrigger.String(),
		TokenTypeProjectDeploy.String(),
		TokenTypeGroupDeploy.String(),
	}
)

func (i Type) String() string {
	return string(i)
}

func (i Type) Value() string {
	return i.String()
}

func TokenTypeParse(value string) (Type, error) {
	if slices.Contains(ValidTokenTypes, value) {
		return Type(value), nil
	}
	return TokenTypeUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownTokenType)
}
