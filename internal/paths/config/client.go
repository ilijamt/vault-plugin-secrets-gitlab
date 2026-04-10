package config

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
	g "gitlab.com/gitlab-org/api/client-go"
)

func (p *Provider) updateConfigClientInfo(ctx context.Context, config *modelConfig.EntryConfig) (et *token.TokenConfig, err error) {
	var httpClient *http.Client
	var client gitlab.Client
	httpClient, _ = utils.HttpClientFromContext(ctx)
	if client, _ = gitlab.ClientFromContext(ctx); client == nil {
		if client, err = gitlab.NewGitlabClient(config, httpClient, p.b.Logger()); err == nil {
			p.b.SetClient(client, config.Name)
		} else {
			return nil, err
		}
	}

	et, err = client.CurrentTokenInfo(ctx)
	if err != nil {
		return et, fmt.Errorf("token cannot be validated: %s", errs.ErrInvalidValue)
	}

	config.TokenCreatedAt = *et.CreatedAt
	config.TokenExpiresAt = *et.ExpiresAt
	config.TokenId = et.TokenID
	config.Scopes = et.Scopes

	var metadata *g.Metadata
	if metadata, err = client.Metadata(ctx); err == nil {
		config.GitlabVersion = metadata.Version
		config.GitlabRevision = metadata.Revision
		config.GitlabIsEnterprise = metadata.Enterprise
	}

	return et, nil
}
