package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	PathConfigStorage = "config"
)

var (
	fieldSchemaConfig = map[string]*framework.FieldSchema{
		"token": {
			Type:        framework.TypeString,
			Description: "The token to access Gitlab API",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name:      "Token",
				Sensitive: true,
			},
		},
		"base_url": {
			Type:        framework.TypeString,
			Description: `The address to access Gitlab. Default is "https://gitlab.com".`,
			Default:     "https://gitlab.com",
		},
		"max_ttl": {
			Type:        framework.TypeDurationSecond,
			Description: `Maximum lifetime expected generated token will be valid for. If set to 0 it will be set for maximum 8670 hours`,
			Default:     DefaultConfigFieldAccessTokenMaxTTL,
		},
		"auto_rotate_token": {
			Type:        framework.TypeBool,
			Default:     false,
			Description: `Should we autorotate the token when it's close to expiry?`,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Auto rotate token",
			},
		},
		"auto_rotate_before": {
			Type:        framework.TypeDurationSecond,
			Description: `How much time should be remaining on the token validity before we should rotate it?`,
			Default:     DefaultConfigFieldAccessTokenRotate,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Auto rotate before",
			},
		},
	}
)

func (b *Backend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return logical.ErrorResponse(ErrBackendNotConfigured.Error()), nil
	}

	if err = req.Storage.Delete(ctx, PathConfigStorage); err != nil {
		return nil, err
	}

	event(ctx, b.Backend, "config-delete", map[string]string{
		"path": "config",
	})

	return nil, nil
}

func (b *Backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return logical.ErrorResponse(ErrBackendNotConfigured.Error()), nil
	}

	return &logical.Response{
		Data: config.LogicalResponseData(),
	}, nil
}

func (b *Backend) checkAndRotateConfigToken(ctx context.Context, request *logical.Request, config *entryConfig) error {
	var client Client
	var err error

	if client, err = b.getClient(ctx, request.Storage); err != nil {
		return err
	}

	if config.TokenExpiresAt.IsZero() {
		var entryToken *EntryToken
		// we need to fetch the token expiration information
		entryToken, err = client.MainTokenInfo()
		if err != nil {
			return err
		}
		// and update the information so we can do the checks
		config.TokenExpiresAt = *entryToken.ExpiresAt
		b.lockClientMutex.Lock()
		err = saveConfig(ctx, *config, request.Storage)
		b.lockClientMutex.Unlock()
	}

	if config.TokenExpiresAt.Sub(time.Now()) > config.AutoRotateBefore {
		b.Logger().Debug("Nothing to do it's not yet time to rotate the token")
		return nil
	}

	return b.rotateConfigToken(ctx, request)
}

func (b *Backend) rotateConfigToken(ctx context.Context, request *logical.Request) error {
	if !b.WriteSafeReplicationState() {
		return nil
	}

	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()

	var config *entryConfig
	var client Client
	var err error

	if config, err = getConfig(ctx, request.Storage); err != nil {
		return err
	}
	if config == nil {
		// no configuration yet so we don't need to rotate anything
		return nil
	}

	if client, err = b.getClient(ctx, request.Storage); err != nil {
		return err
	}

	var entryToken *EntryToken
	entryToken, err = client.RotateMainToken()
	if err != nil {
		return nil
	}

	config.Token = entryToken.Token

	err = saveConfig(ctx, *config, request.Storage)
	if err != nil {
		return err
	}

	event(ctx, b.Backend, "config-token-rotate", map[string]string{
		"path": "config",
	})

	return nil
}

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var warnings []string
	var maxTtlRaw, maxTtlOk = data.GetOk("max_ttl")
	var autoTokenRotateRaw, autoTokenRotateTtlOk = data.GetOk("auto_rotate_before")
	var token, tokenOk = data.GetOk("token")
	var err error

	if !tokenOk {
		err = multierror.Append(err, fmt.Errorf("token: %w", ErrFieldRequired))
	}

	var config = entryConfig{
		BaseURL:         data.Get("base_url").(string),
		AutoRotateToken: data.Get("auto_rotate_token").(bool),
	}

	if maxTtlOk {
		maxTtl := maxTtlRaw.(int)
		switch {
		case maxTtl > 0 && maxTtl < int(DefaultAccessTokenMinTTL.Seconds()):
			warnings = append(warnings, "max_ttl is set with less than 24 hours. With current token expiry limitation, this max_ttl is ignored, it's set to 24 hours")
			config.MaxTTL = DefaultAccessTokenMinTTL
		case maxTtl <= 0:
			config.MaxTTL = DefaultAccessTokenMaxPossibleTTL
			warnings = append(warnings, "max_ttl is not set. Token wil be generated with expiration date of '8760 hours'")
		case maxTtl > int(DefaultAccessTokenMaxPossibleTTL.Seconds()):
			warnings = append(warnings, "max_ttl is set to more than '8760 hours'. Token wil be generated with expiration date of '8760 hours'")
			config.MaxTTL = DefaultAccessTokenMaxPossibleTTL
		default:
			config.MaxTTL = time.Duration(maxTtl) * time.Second
		}
	} else if config.MaxTTL == 0 {
		config.MaxTTL = DefaultAccessTokenMaxPossibleTTL
	}

	if autoTokenRotateTtlOk {
		atr, _ := convertToInt(autoTokenRotateRaw)
		if atr > int(config.MaxTTL.Seconds()*DefaultAutoRotateBeforeMaxFraction) {
			err = multierror.Append(err, fmt.Errorf("auto_rotate_token can not be bigger than %d%% (max: %s) of %s: %w", int(DefaultAutoRotateBeforeMaxFraction*100), time.Duration(config.MaxTTL.Seconds()*DefaultAutoRotateBeforeMaxFraction)*time.Second, config.MaxTTL.String(), ErrInvalidValue))
		} else if atr <= int(config.MaxTTL.Seconds()*DefaultAutoRotateBeforeMinFraction) {
			err = multierror.Append(err, fmt.Errorf("auto_rotate_token can not be less than %d%% (max: %s) of %s: %w", int(DefaultAutoRotateBeforeMinFraction*100), time.Duration(config.MaxTTL.Seconds()*DefaultAutoRotateBeforeMinFraction)*time.Second, config.MaxTTL.String(), ErrInvalidValue))
		} else {
			config.AutoRotateBefore = time.Duration(atr) * time.Second
		}
	} else {
		config.AutoRotateBefore = time.Duration(config.MaxTTL.Seconds()*DefaultAutoRotateBeforeMinFraction) * time.Second
		warnings = append(warnings, fmt.Sprintf("auto_rotate_token not specified setting to %v (%d%% of %s)", config.AutoRotateBefore.String(), int(DefaultAutoRotateBeforeMinFraction*100), config.MaxTTL.String()))
	}

	if err != nil {
		return nil, err
	}

	config.Token = token.(string)

	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()

	err = saveConfig(ctx, config, req.Storage)
	if err != nil {
		return nil, err
	}

	event(ctx, b.Backend, "config-write", map[string]string{
		"path":               "config",
		"max_ttl":            config.MaxTTL.String(),
		"auto_rotate_token":  strconv.FormatBool(config.AutoRotateToken),
		"auto_rotate_before": config.AutoRotateBefore.String(),
		"base_url":           config.BaseURL,
	})

	b.Logger().Debug("Wrote new config", "base_url", config.BaseURL, "max_ttl", config.MaxTTL)
	return &logical.Response{
		Data:     config.LogicalResponseData(),
		Warnings: warnings,
	}, nil

}

func pathConfig(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathConfigHelpSynopsis),
		HelpDescription: strings.TrimSpace(pathConfigHelpDescription),
		Pattern:         fmt.Sprintf("%s$", PathConfigStorage),
		Fields:          fieldSchemaConfig,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: operationPrefixGitlabAccessTokens,
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback:     b.pathConfigWrite,
				DisplayAttrs: &framework.DisplayAttributes{OperationVerb: "configure"},
				Summary:      "Configure Backend level settings that are applied to all credentials.",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb:   "read",
					OperationSuffix: "configuration",
				},
				Summary: "Read the Backend level settings.",
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields:      fieldSchemaConfig,
					}},
				},
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb:   "delete",
					OperationSuffix: "configuration",
				},
				Summary: "Delete the Backend level settings.",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
		},
	}
}

const pathConfigHelpSynopsis = `Configure the Gitlab Access Tokens Backend.`

const pathConfigHelpDescription = `
The Gitlab Access Tokens auth Backend requires credentials for managing
private and group access tokens for Gitlab. This endpoint
is used to configure those credentials and the default values for the Backend input general.

You must specify expected Gitlab token with access to allow Vault to create tokens.
`
