package gitlab

import (
	"errors"
	"fmt"

	"golang.org/x/exp/slices"
)

type TokenType string

const (
	TokenTypePersonal = TokenType("personal")
	TokenTypeProject  = TokenType("project")
	TokenTypeGroup    = TokenType("group")

	TokenTypeUnknown = TokenType("")
)

var (
	ErrUnknownTokenType = errors.New("unknown token type")

	validTokenTypes = []string{
		TokenTypePersonal.String(),
		TokenTypeProject.String(),
		TokenTypeGroup.String(),
	}
)

func (i TokenType) String() string {
	return string(i)
}

func (i TokenType) Value() string {
	return i.String()
}

func TokenTypeParse(value string) (TokenType, error) {
	if slices.Contains(validTokenTypes, value) {
		return TokenType(value), nil
	}
	return TokenTypeUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownTokenType)
}
