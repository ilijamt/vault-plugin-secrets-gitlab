package config

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths"
)

const (
	pathListConfigHelpSyn  = `Lists existing configs`
	pathListConfigHelpDesc = `
This path allows you to list all available configurations that have been set up within the GitLab Access Tokens Backend.
These configurations typically include credentials, base URLs, and other settings required for managing access tokens
across different GitLab environments.`
)

func (p *Provider) pathListConfig() *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathListConfigHelpSyn),
		HelpDescription: strings.TrimSpace(pathListConfigHelpDesc),
		Pattern:         fmt.Sprintf("%s?/?$", backend.PathConfigStorage),
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: paths.OperationPrefixGitlabAccessTokens,
			OperationSuffix: "config",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: p.pathConfigList,
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb: "list",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields: map[string]*framework.FieldSchema{
							"config_name": {
								Type:        framework.TypeString,
								Default:     backend.DefaultConfigName,
								Required:    false,
								Description: "The config we use when interacting with the role, this can be specified if you want to use a specific config for the role, otherwise it uses the default one.",
								DisplayAttrs: &framework.DisplayAttributes{
									Name: "Configuration.",
								},
							},
						},
					}},
				},
			},
		},
	}
}

func (p *Provider) pathConfigList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	configs, err := req.Storage.List(ctx, fmt.Sprintf("%s/", backend.PathConfigStorage))
	if err != nil {
		return logical.ErrorResponse("Error listing configs"), err
	}
	p.b.Logger().Debug("Available", "configs", configs)
	return logical.ListResponse(configs), nil
}
