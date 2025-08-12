package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	g "gitlab.com/gitlab-org/api/client-go"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/models"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

const (
	PathConfigStorage = "config"
)

var (
	FieldSchemaConfig = map[string]*framework.FieldSchema{
		"token": {
			Type:        framework.TypeString,
			Description: "The API access token required for authenticating requests to the GitLab API. This token must be a valid personal access token or any other type of token supported by GitLab for API access.",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name:      "Token",
				Sensitive: true,
			},
		},
		"base_url": {
			Type:     framework.TypeString,
			Required: true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "GitLab Base URL",
			},
			Description: `The base URL of your GitLab instance. This could be the URL of a self-managed GitLab instance or the URL of the GitLab SaaS service (https://gitlab.com). The URL must be properly formatted, including the scheme (http or https). This field is essential as it determines the endpoint where API requests will be directed.`,
		},
		"type": {
			Type:     framework.TypeString,
			Required: true,
			AllowedValues: []any{
				gitlab.TypeSelfManaged,
				gitlab.TypeSaaS,
				gitlab.TypeDedicated,
			},
			Description: `The type of GitLab instance you are connecting to. This could typically distinguish between 'self-managed' for on-premises GitLab installations or 'saas' or 'dedicated' for the GitLab SaaS offering. This field helps the plugin to adjust any necessary configurations or request patterns specific to the type of GitLab instance.`,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "GitLab Type",
			},
		},
		"auto_rotate_token": {
			Type:        framework.TypeBool,
			Default:     false,
			Description: `Determines whether the plugin should automatically rotate the API access token as it approaches its expiration date. Enabling this feature ensures that the token is refreshed without manual intervention, reducing the risk of service disruption due to expired tokens.`,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Auto Rotate Token",
			},
		},
		"auto_rotate_before": {
			Type:        framework.TypeDurationSecond,
			Description: `Specifies the duration, in seconds, before the token's expiration at which the auto-rotation should occur. The value must be set between a minimum of 24 hours (86400 seconds) and a maximum of 730 hours (2628000 seconds). This setting allows you to control how early the token rotation should happen, balancing between proactive rotation and maximizing token lifespan.`,
			Default:     DefaultConfigFieldAccessTokenRotate,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Auto Rotate Before",
			},
		},
		"config_name": {
			Type:        framework.TypeString,
			Description: "Config name",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Config name",
			},
		},
	}
)

func (b *Backend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()
	var err error
	var name = data.Get("config_name").(string)

	if config, err := getConfig(ctx, req.Storage, name); err == nil {
		if config == nil {
			return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
		}

		if err = req.Storage.Delete(ctx, fmt.Sprintf("%s/%s", PathConfigStorage, name)); err == nil {
			_ = event.Event(ctx, b.Backend, operationPrefixGitlabAccessTokens, "config-delete", map[string]string{
				"path": fmt.Sprintf("%s/%s", PathConfigStorage, name),
			})
			b.SetClient(nil, name)
		}
	}

	return nil, err
}

func (b *Backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()

	var name = data.Get("config_name").(string)
	var config *EntryConfig
	if config, err = getConfig(ctx, req.Storage, name); err == nil {
		if config == nil {
			return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
		}
		lrd := config.LogicalResponseData(b.flags.ShowConfigToken)
		b.Logger().Debug("Reading configuration info", "info", lrd)
		lResp = &logical.Response{Data: lrd}
	}
	return lResp, err
}

func (b *Backend) pathConfigPatch(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	var name = data.Get("config_name").(string)
	var warnings []string
	var changes map[string]string
	var config *EntryConfig
	config, err = getConfig(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
	}

	warnings, changes, err = config.Merge(data)
	if err != nil {
		return nil, err
	}

	if _, ok := data.GetOk("token"); ok {
		if _, err = b.updateConfigClientInfo(ctx, config); err != nil {
			return nil, err
		}
	}

	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()
	if err = saveConfig(ctx, *config, req.Storage); err == nil {
		lrd := config.LogicalResponseData(b.flags.ShowConfigToken)
		_ = event.Event(ctx, b.Backend, operationPrefixGitlabAccessTokens, "config-patch", changes)
		b.SetClient(nil, name)
		b.Logger().Debug("Patched config", "lrd", lrd, "warnings", warnings)
		lResp = &logical.Response{Data: lrd, Warnings: warnings}
	}

	return lResp, err
}

func (b *Backend) updateConfigClientInfo(ctx context.Context, config *EntryConfig) (et *models.TokenConfig, err error) {
	var httpClient *http.Client
	var client Client
	httpClient, _ = utils.HttpClientFromContext(ctx)
	if client, _ = ClientFromContext(ctx); client == nil {
		if client, err = NewGitlabClient(config, httpClient, b.Logger()); err == nil {
			b.SetClient(client, config.Name)
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

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var name = data.Get("config_name").(string)
	var config = new(EntryConfig)
	var warnings, err = config.UpdateFromFieldData(data)
	if err != nil {
		return nil, err
	}
	config.Name = name

	if _, err = b.updateConfigClientInfo(ctx, config); err != nil {
		return nil, err
	}

	b.lockClientMutex.Lock()
	defer b.lockClientMutex.Unlock()
	var lResp *logical.Response

	if err = saveConfig(ctx, *config, req.Storage); err == nil {
		_ = event.Event(ctx, b.Backend, operationPrefixGitlabAccessTokens, "config-write", map[string]string{
			"path":               fmt.Sprintf("%s/%s", PathConfigStorage, name),
			"auto_rotate_token":  strconv.FormatBool(config.AutoRotateToken),
			"auto_rotate_before": config.AutoRotateBefore.String(),
			"base_url":           config.BaseURL,
			"token_id":           strconv.Itoa(config.TokenId),
			"created_at":         config.TokenCreatedAt.Format(time.RFC3339),
			"expires_at":         config.TokenExpiresAt.Format(time.RFC3339),
			"scopes":             strings.Join(config.Scopes, ", "),
			"type":               config.Type.String(),
			"config_name":        config.Name,
		})

		b.SetClient(nil, name)
		lrd := config.LogicalResponseData(b.flags.ShowConfigToken)
		b.Logger().Debug("Wrote new config", "lrd", lrd, "warnings", warnings)
		lResp = &logical.Response{Data: lrd, Warnings: warnings}
	}

	return lResp, err
}

func pathConfig(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathConfigHelpSynopsis),
		HelpDescription: strings.TrimSpace(pathConfigHelpDescription),
		Pattern:         fmt.Sprintf("%s/%s", PathConfigStorage, framework.GenericNameRegex("config_name")),
		Fields:          FieldSchemaConfig,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: operationPrefixGitlabAccessTokens,
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.PatchOperation: &framework.PathOperation{
				Callback:     b.pathConfigPatch,
				DisplayAttrs: &framework.DisplayAttributes{OperationVerb: "configure"},
				Summary:      "Configure Backend level settings that are applied to all credentials.",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
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
						Fields:      FieldSchemaConfig,
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
