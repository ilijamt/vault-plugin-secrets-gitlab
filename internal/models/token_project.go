package models

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"

type TokenProject struct {
	TokenWithScopesAndAccessLevel `json:",inline"`
}

var _ token.Token = (*TokenProject)(nil)
