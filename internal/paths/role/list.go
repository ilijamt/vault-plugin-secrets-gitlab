package role

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

func (p *Provider) pathRolesList(ctx context.Context, req *logical.Request, data *framework.FieldData) (l *logical.Response, err error) {
	var roles []string
	defer func() {
		p.b.Logger().Debug("Available", "roles", roles, "err", err)
	}()
	l = logical.ErrorResponse("Error listing roles")
	if roles, err = req.Storage.List(ctx, fmt.Sprintf("%s/", backend.PathRoleStorage)); err == nil {
		l = logical.ListResponse(roles)
	}
	return l, err
}

func (p *Provider) pathListRoles() *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathListRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathListRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s?/?$", backend.PathRoleStorage),
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: paths.OperationPrefixGitlabAccessTokens,
			OperationSuffix: "roles",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: p.pathRolesList,
				DisplayAttrs: &framework.DisplayAttributes{
					OperationVerb: "list",
				},
				Responses: map[int][]framework.Response{
					http.StatusOK: {{
						Description: http.StatusText(http.StatusOK),
						Fields: map[string]*framework.FieldSchema{
							"role_name": FieldSchemaRoles["role_name"],
						},
					}},
				},
			},
		},
	}
}

const (
	pathListRolesHelpSyn  = `Lists existing roles`
	pathListRolesHelpDesc = `
This path allows you to list all available roles that have been created within the GitLab Access Tokens Backend. 
Each role defines a set of parameters, such as token permissions, scopes, and expiration settings, which are used 
when generating access tokens.`
)
