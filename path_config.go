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
			Required:    true,
			Description: `The address to access Gitlab.`,
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
			Description: `How much time should be remaining on the token validity before we should rotate it? Minimum can be set to 24h and maximum to 730h`,
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

	lrd := config.LogicalResponseData()
	b.Logger().Debug("Reading configuration info", "info", lrd)
	return &logical.Response{
		Data: config.LogicalResponseData(),
	}, nil
}

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var warnings []string
	var autoTokenRotateRaw, autoTokenRotateTtlOk = data.GetOk("auto_rotate_before")
	var token, tokenOk = data.GetOk("token")
	var err error

	if !tokenOk {
		err = multierror.Append(err, fmt.Errorf("token: %w", ErrFieldRequired))
	}

	var config = EntryConfig{
		BaseURL:         data.Get("base_url").(string),
		AutoRotateToken: data.Get("auto_rotate_token").(bool),
	}

	if autoTokenRotateTtlOk {
		atr, _ := convertToInt(autoTokenRotateRaw)
		if atr > int(DefaultAutoRotateBeforeMaxTTL.Seconds()) {
			err = multierror.Append(err, fmt.Errorf("auto_rotate_token can not be bigger than %s: %w", DefaultAutoRotateBeforeMaxTTL, ErrInvalidValue))
		} else if atr <= int(DefaultAutoRotateBeforeMinTTL.Seconds())-1 {
			err = multierror.Append(err, fmt.Errorf("auto_rotate_token can not be less than %s: %w", DefaultAutoRotateBeforeMinTTL, ErrInvalidValue))
		} else {
			config.AutoRotateBefore = time.Duration(atr) * time.Second
		}
	} else {
		config.AutoRotateBefore = DefaultAutoRotateBeforeMinTTL
		warnings = append(warnings, fmt.Sprintf("auto_rotate_token not specified setting to %s", DefaultAutoRotateBeforeMinTTL))
	}

	if err != nil {
		return nil, err
	}

	config.Token = token.(string)

	var httpClient *http.Client
	var client Client
	httpClient, _ = HttpClientFromContext(ctx)
	if client, _ = GitlabClientFromContext(ctx); client == nil {
		if client, err = NewGitlabClient(&config, httpClient, b.Logger()); err == nil {
			b.SetClient(client)
		}
	}

	var et *EntryToken
	et, err = client.CurrentTokenInfo()
	if err != nil {
		return nil, fmt.Errorf("token cannot be validated: %s", ErrInvalidValue)
	}

	config.TokenCreatedAt = *et.CreatedAt
	config.TokenExpiresAt = *et.ExpiresAt
	config.TokenId = et.TokenID
	config.Scopes = et.Scopes

	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()
	err = saveConfig(ctx, config, req.Storage)
	if err != nil {
		return nil, err
	}

	event(ctx, b.Backend, "config-write", map[string]string{
		"path":               "config",
		"auto_rotate_token":  strconv.FormatBool(config.AutoRotateToken),
		"auto_rotate_before": config.AutoRotateBefore.String(),
		"base_url":           config.BaseURL,
		"token_id":           strconv.Itoa(config.TokenId),
		"created_at":         config.TokenCreatedAt.Format(time.RFC3339),
		"expires_at":         config.TokenExpiresAt.Format(time.RFC3339),
		"scopes":             strings.Join(config.Scopes, ", "),
	})

	b.SetClient(nil)
	lrd := config.LogicalResponseData()
	b.Logger().Debug("Wrote new config", "lrd", lrd, "warnings", warnings)
	return &logical.Response{
		Data:     lrd,
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
