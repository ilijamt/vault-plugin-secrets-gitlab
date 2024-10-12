package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	pathListConfigHelpSyn  = `Lists existing configs`
	pathListConfigHelpDesc = `This path allows you to list all available configurations.`
)

func pathListConfig(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathListConfigHelpSyn),
		HelpDescription: strings.TrimSpace(pathListConfigHelpDesc),
		Pattern:         fmt.Sprintf("%s?/?$", PathConfigStorage),
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: operationPrefixGitlabAccessTokens,
			OperationSuffix: "config",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.pathConfigList,
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb: "list",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields: map[string]*framework.FieldSchema{
							"config_name": FieldSchemaRoles["config_name"],
						},
					}},
				},
			},
		},
	}
}

func (b *Backend) pathConfigList(ctx context.Context, req *logical.Request, data *framework.FieldData) (lResp *logical.Response, err error) {
	var configs []string
	configs, err = req.Storage.List(ctx, fmt.Sprintf("%s/", PathConfigStorage))
	lResp = logical.ErrorResponse("Error listing configs")
	if err == nil {
		lResp = logical.ListResponse(configs)
	}
	b.Logger().Debug("Available configs input the system", "configs", configs)
	return lResp, err
}
