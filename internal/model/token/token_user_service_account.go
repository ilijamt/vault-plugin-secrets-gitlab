package token

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"

type TokenUserServiceAccount struct {
	TokenWithScopes `json:",inline"`
}

var _ token.Token = (*TokenUserServiceAccount)(nil)
