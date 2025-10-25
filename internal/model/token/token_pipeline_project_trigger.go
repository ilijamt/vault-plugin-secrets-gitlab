package token

import "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"

type TokenPipelineProjectTrigger struct {
	Token `json:",inline"`
}

var _ token.Token = (*TokenPipelineProjectTrigger)(nil)
