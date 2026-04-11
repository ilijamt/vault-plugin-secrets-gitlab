package config

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths"
)

const pathConfigRotateHelpSynopsis = `Rotate the gitlab token for this configuration.`

const pathConfigRotateHelpDescription = `
This endpoint allows you to rotate the GitLab token associated with your current configuration. When you invoke this 
operation, Vault securely generates a new token and replaces the existing one revealing the new token to you. It
will only reveal it once, after that you will be unable to retrieve it. The newly generated token is securely 
stored within Vault's internal storage, ensuring that only Vault has access to it for future use when interacting 
with the GitLab API.'`

func (p *Provider) pathConfigTokenRotate() *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathConfigRotateHelpSynopsis),
		HelpDescription: strings.TrimSpace(pathConfigRotateHelpDescription),
		Pattern:         fmt.Sprintf("%s/%s/rotate$", backend.PathConfigStorage, framework.GenericNameRegex("config_name")),
		Fields:          FieldSchemaConfig,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: paths.OperationPrefixGitlabAccessTokens,
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback:     p.pathConfigTokenRotateHandler,
				DisplayAttrs: &framework.DisplayAttributes{OperationVerb: "configure"},
				Summary:      "Rotate the main Gitlab Access Token.",
			},
		},
	}
}

func (p *Provider) checkAndRotateConfigToken(ctx context.Context, request *logical.Request, config *modelConfig.EntryConfig) error {
	var err error
	p.b.Logger().Debug("Running check and rotate config token")
	if time.Until(config.TokenExpiresAt) <= config.AutoRotateBefore {
		_, err = p.pathConfigTokenRotateHandler(ctx, request, &framework.FieldData{
			Raw: map[string]interface{}{
				"config_name": cmp.Or(config.Name, backend.DefaultConfigName),
			},
			Schema: FieldSchemaConfig,
		})
	}
	return err
}

func (p *Provider) pathConfigTokenRotateHandler(ctx context.Context, request *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	var name = data.Get("config_name").(string)
	p.b.Logger().Debug("Running pathConfigTokenRotate")
	var config *modelConfig.EntryConfig
	var client gitlab.Client

	p.b.ClientRLock()
	if config, err = p.b.GetConfig(ctx, request.Storage, name); err != nil {
		p.b.ClientRUnlock()
		p.b.Logger().Error("Failed to fetch configuration", "error", err.Error())
		return nil, err
	}
	p.b.ClientRUnlock()

	if config == nil {
		// no configuration yet so we don't need to rotate anything
		return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
	}

	if client, err = p.b.GetClientByName(ctx, request.Storage, name); err != nil {
		return nil, err
	}

	var entryToken *token.TokenConfig
	entryToken, _, err = client.RotateCurrentToken(ctx)
	if err != nil {
		p.b.Logger().Error("Failed to rotate main token", "err", err)
		return nil, err
	}

	config.Token = entryToken.Token.Token
	config.TokenId = entryToken.TokenID
	config.Scopes = entryToken.Scopes
	if entryToken.CreatedAt != nil {
		config.TokenCreatedAt = *entryToken.CreatedAt
	}
	if entryToken.ExpiresAt != nil {
		config.TokenExpiresAt = *entryToken.ExpiresAt
	}
	p.b.ClientLock()
	defer p.b.ClientUnlock()
	err = p.b.SaveConfig(ctx, request.Storage, config)
	if err != nil {
		p.b.Logger().Error("failed to store configuration for revocation", "err", err)
		return nil, err
	}

	lResp = &logical.Response{Data: config.LogicalResponseData(p.b.Flags().ShowConfigToken)}
	lResp.Data["token"] = config.Token
	_ = p.b.SendEvent(ctx, eventTokenRotate, map[string]string{
		"path":        fmt.Sprintf("%s/%s", backend.PathConfigStorage, name),
		"expires_at":  entryToken.ExpiresAt.Format(time.RFC3339),
		"created_at":  entryToken.CreatedAt.Format(time.RFC3339),
		"scopes":      strings.Join(entryToken.Scopes, ", "),
		"token_id":    strconv.FormatInt(entryToken.TokenID, 10),
		"name":        entryToken.Name,
		"config_name": entryToken.ConfigName,
	})

	p.b.SetClient(nil, name)
	return lResp, err
}

// PeriodicFunc implements backend.PeriodicHandler.
// It checks all configs for auto-rotation needs.
func (p *Provider) PeriodicFunc(ctx context.Context, req *logical.Request) (err error) {
	var configs []string
	configs, err = req.Storage.List(ctx, fmt.Sprintf("%s/", backend.PathConfigStorage))
	if err != nil {
		return err
	}

	for _, name := range configs {
		config, cfgErr := p.b.GetConfig(ctx, req.Storage, name)
		if cfgErr != nil {
			err = errors.Join(err, cfgErr)
			continue
		}
		if config != nil && config.AutoRotateToken {
			p.b.Logger().Debug("Trying to rotate the config", "name", name)
			err = errors.Join(err, p.checkAndRotateConfigToken(ctx, req, config))
		}
	}

	return err
}

// Invalidate implements backend.InvalidateHandler.
// It clears the cached client when a config key changes.
func (p *Provider) Invalidate(ctx context.Context, key string) {
	if strings.HasPrefix(key, backend.PathConfigStorage) {
		parts := strings.SplitN(key, "/", 2)
		var name = parts[1]
		p.b.Logger().Warn(fmt.Sprintf("Gitlab config for %s changed, reinitializing the gitlab client", name))
		p.b.ClientLock()
		defer p.b.ClientUnlock()
		p.b.DeleteClient(name)
	}
}
