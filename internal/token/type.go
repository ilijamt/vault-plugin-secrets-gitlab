package token

import (
	"errors"
	"fmt"
	"slices"
)

type TokenType string

const (
	TokenTypePersonal               = TokenType("personal")
	TokenTypeProject                = TokenType("project")
	TokenTypeGroup                  = TokenType("group")
	TokenTypeUserServiceAccount     = TokenType("user-service-account")
	TokenTypeGroupServiceAccount    = TokenType("group-service-account")
	TokenTypePipelineProjectTrigger = TokenType("pipeline-project-trigger")
	TokenTypeProjectDeploy          = TokenType("project-deploy")
	TokenTypeGroupDeploy            = TokenType("group-deploy")

	TokenTypeUnknown = TokenType("")
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

func (i TokenType) String() string {
	return string(i)
}

func (i TokenType) Value() string {
	return i.String()
}

func TokenTypeParse(value string) (TokenType, error) {
	if slices.Contains(ValidTokenTypes, value) {
		return TokenType(value), nil
	}
	return TokenTypeUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownTokenType)
}
