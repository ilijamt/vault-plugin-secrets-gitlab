package gitlab

import (
	"errors"
	"fmt"
	"slices"
)

type Type string

const (
	TypeSaaS        Type = "saas"
	TypeDedicated   Type = "dedicated"
	TypeSelfManaged Type = "self-managed"
	TypeUnknown          = Type("")
)

var (
	ErrUnknownType = errors.New("unknown gitlab type")

	validGitlabTypes = []string{
		TypeSaaS.String(),
		TypeSelfManaged.String(),
		TypeDedicated.String(),
	}
)

func (i Type) String() string {
	return string(i)
}

func (i Type) Value() string {
	return i.String()
}

func TypeParse(value string) (Type, error) {
	if slices.Contains(validGitlabTypes, value) {
		return Type(value), nil
	}
	return TypeUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownType)
}
