package gitlab

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"net/http"
	"strings"
	"time"
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
	}
)

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

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var warnings []string
	var maxTtlRaw, maxTtlOk = data.GetOk("max_ttl")
	var token, tokenOk = data.GetOk("token")
	var err error

	if !tokenOk {
		err = multierror.Append(err, fmt.Errorf("token: %w", ErrFieldRequired))
	}

	var config = entryConfig{
		BaseURL: data.Get("base_url").(string),
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

	if err != nil {
		return nil, err
	}

	config.Token = token.(string)

	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()

	entry, err := logical.StorageEntryJSON(PathConfigStorage, config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	event(ctx, b.Backend, "config-write", map[string]string{
		"path":     "config",
		"max_ttl":  config.MaxTTL.String(),
		"base_url": config.BaseURL,
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
