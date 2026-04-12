package config

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths"
)

const (
	pathConfigHelpSynopsis = `Configure the Gitlab Access Tokens Backend.`

	pathConfigHelpDescription = `
The Gitlab Access Tokens auth Backend requires credentials for managing
private and group access tokens for Gitlab. This endpoint
is used to configure those credentials and the default values for the Backend input general.

You must specify expected Gitlab token with access to allow Vault to create tokens.
`
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
				gitlabTypes.TypeSelfManaged,
				gitlabTypes.TypeSaaS,
				gitlabTypes.TypeDedicated,
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
			Default:     backend.DefaultConfigFieldAccessTokenRotate,
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

// configBackend defines the narrow interface this provider needs.
type configBackend interface {
	backend.Logging
	backend.FlagsProvider
	backend.ClientReader
	backend.ClientSetter
	backend.ClientDeleter
	backend.ConfigStore
	backend.EventSender
	backend.Locker
}

// Provider implements backend.PathProvider, backend.PeriodicHandler,
// and backend.InvalidateHandler for config paths.
type Provider struct {
	b configBackend
}

func (p *Provider) Name() string { return "config" }

// New creates a new config path provider.
func New(b configBackend) *Provider {
	return &Provider{b: b}
}

// Paths returns all config-related framework paths.
func (p *Provider) Paths() []*framework.Path {
	return []*framework.Path{
		p.pathConfig(),
		p.pathListConfig(),
		p.pathConfigTokenRotate(),
	}
}

func (p *Provider) pathConfig() *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathConfigHelpSynopsis),
		HelpDescription: strings.TrimSpace(pathConfigHelpDescription),
		Pattern:         fmt.Sprintf("%s/%s", backend.PathConfigStorage, framework.GenericNameRegex("config_name")),
		Fields:          FieldSchemaConfig,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: paths.OperationPrefixGitlabAccessTokens,
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.PatchOperation: &framework.PathOperation{
				Callback:     p.pathConfigPatch,
				DisplayAttrs: &framework.DisplayAttributes{OperationVerb: "configure"},
				Summary:      "Configure Backend level settings that are applied to all credentials.",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback:     p.pathConfigWrite,
				DisplayAttrs: &framework.DisplayAttributes{OperationVerb: "configure"},
				Summary:      "Configure Backend level settings that are applied to all credentials.",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: p.pathConfigRead,
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
				Callback: p.pathConfigDelete,
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
