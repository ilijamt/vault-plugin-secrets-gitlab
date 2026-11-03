package utils

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

const maxTokenDescriptionLength = 255

// TokenDescriptionData is a role that can provide a token description template.
type TokenDescriptionData interface {
	TokenNameData
	GetDescription() string
}

// ValidateTokenDescription validates the template syntax of a token description.
func ValidateTokenDescription(role TokenDescriptionData) error {
	_, err := template.New("description").Funcs(tplFuncMap).Parse(role.GetDescription())
	return err
}

// TokenDescription renders the role's description, falling back to the default
// so an issued token always shows where it came from. GitLab rejects
// descriptions over 255 characters.
func TokenDescription(role TokenDescriptionData, mount string) (string, error) {
	tpl, err := template.New("description").Funcs(tplFuncMap).Parse(role.GetDescription())
	if err != nil {
		return "", err
	}

	data := role.LogicalResponseData()
	data["mount"] = mount
	data["unix_timestamp_utc"] = time.Now().UTC().Unix()
	delete(data, "description")

	buf := new(strings.Builder)
	if err = tpl.Execute(buf, data); err != nil {
		return "", err
	}

	description := buf.String()
	if strings.TrimSpace(description) == "" {
		description = fmt.Sprintf("Issued by Vault: mount=%s role=%s config=%s", mount, data["role_name"], data["config_name"])
	}

	runes := []rune(description)
	return string(runes[:min(len(runes), maxTokenDescriptionLength)]), nil
}
