package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/logging"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

const gitlabInfoRefreshInterval = time.Hour

func (p *Provider) updateConfigClientInfo(ctx context.Context, config *modelConfig.EntryConfig) (err error) {
	var httpClient *http.Client
	var client gitlab.Client
	var et *token.TokenConfig
	httpClient, _ = utils.HttpClientFromContext(ctx)
	if client, _ = gitlab.ClientFromContext(ctx); client == nil {
		if client, err = gitlab.NewGitlabClient(config, httpClient, p.b.Logger()); err == nil {
			p.b.SetClient(client, config.Name)
		} else {
			return err
		}
	}

	et, err = client.CurrentTokenInfo(ctx)
	if err != nil {
		return fmt.Errorf("token cannot be validated: %w", errs.ErrInvalidValue)
	}

	config.TokenCreatedAt = *et.CreatedAt
	config.TokenExpiresAt = *et.ExpiresAt
	config.TokenId = et.TokenID
	config.Scopes = et.Scopes

	if err = refreshGitlabInfo(ctx, client, config); err != nil {
		logging.New(p.b.Logger()).Debug("Could not read all GitLab information", logging.Fields{ConfigName: config.Name, Err: err})
	}

	return nil
}

func refreshGitlabInfo(ctx context.Context, client gitlab.Client, config *modelConfig.EntryConfig) error {
	metadata, metadataErr := client.Metadata(ctx)
	if metadataErr == nil && metadata != nil {
		config.GitlabVersion = metadata.Version
		config.GitlabRevision = metadata.Revision
		config.GitlabIsEnterprise = metadata.Enterprise
	}

	info, infoErr := client.InstanceInfo(ctx)
	if infoErr == nil && info != nil {
		config.GitlabInstance = *info
	}

	if metadataErr == nil && infoErr == nil {
		config.GitlabRefreshedAt = utils.TimeFromContext(ctx)
	}
	return errors.Join(metadataErr, infoErr)
}

func (p *Provider) refreshConfig(ctx context.Context, req *logical.Request, name string) error {
	l := p.b.LockForKey("config", name)
	l.Lock()
	defer l.Unlock()

	config, err := p.b.GetConfig(ctx, req.Storage, name)
	if err != nil || config == nil {
		return err
	}
	if utils.TimeFromContext(ctx).Sub(config.GitlabRefreshedAt) < gitlabInfoRefreshInterval {
		return nil
	}

	client, err := p.b.GetClientByName(ctx, req.Storage, name)
	if err != nil || client == nil {
		return err
	}
	if err = refreshGitlabInfo(ctx, client, config); err != nil {
		return err
	}
	return p.b.SaveConfig(ctx, req.Storage, config)
}
