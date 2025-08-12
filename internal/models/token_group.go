package models

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"

type TokenGroup struct {
	TokenWithScopesAndAccessLevel `json:",inline"`
}

var _ token.Token = (*TokenGroup)(nil)
