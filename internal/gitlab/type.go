package gitlab

import (
	"fmt"
	"slices"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

// Type defines a string-based type to represent specific categories or modes, such as "saas" or "self-managed". This is the Gitlab Type
type Type string

const (
	// TypeSaaS represents the "saas" type, indicating the software-as-a-service mode for GitLab deployments.
	TypeSaaS Type = "saas"
	// TypeDedicated represents the "dedicated" type, indicating a dedicated mode for GitLab deployments.
	TypeDedicated Type = "dedicated"
	// TypeSelfManaged represents the "self-managed" type, indicating a self-hosted mode for GitLab deployments.
	TypeSelfManaged Type = "self-managed"

	// TypeUnknown represents an uninitialized or unknown GitLab deployment type, used as a default fallback value.
	TypeUnknown = Type("")
)

var (
	ErrUnknownType = fmt.Errorf("%s: gitlab type", errs.ErrInvalidValue)

	validGitlabTypes = []string{
		TypeSaaS.String(),
		TypeSelfManaged.String(),
		TypeDedicated.String(),
	}
)

// String converts the Type value to its underlying string representation.
func (i Type) String() string {
	return string(i)
}

// Value returns the string representation of the Type by invoking the String method.
func (i Type) Value() string {
	return i.String()
}

// TypeParse attempts to parse the given string into a valid GitLab Type.
// Returns the corresponding Type and nil error if successful, or TypeUnknown and an error if parsing fails.
func TypeParse(value string) (Type, error) {
	if slices.Contains(validGitlabTypes, value) {
		return Type(value), nil
	}
	return TypeUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownType)
}
