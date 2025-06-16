package gitlab

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

const (
	PathRoleStorage   = "roles"
	TypeConfigDefault = DefaultConfigName
)

var (
	FieldSchemaRoles = map[string]*framework.FieldSchema{
		"role_name": {
			Type:        framework.TypeString,
			Description: "Role name",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Role Name",
			},
		},
		"path": {
			Type:        framework.TypeString,
			Description: "Project/Group path to create an access token for. If the token type is set to personal then write the username here.",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "path",
			},
		},
		"name": {
			Type:        framework.TypeString,
			Description: "The name of the access token",
			Required:    true,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Name",
			},
		},
		"scopes": {
			Type:        framework.TypeCommaStringSlice,
			Description: "List of scopes",
			Required:    false,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Scopes",
			},
			AllowedValues: utils.ToAny(ValidPersonalTokenScopes...),
		},
		"ttl": {
			Type:        framework.TypeDurationSecond,
			Description: "The TTL of the token",
			Required:    false,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Token TTL",
			},
		},
		"access_level": {
			Type:        framework.TypeString,
			Description: "access level of access token (only required for Group and Project access tokens)",
			Required:    false,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Access Level",
			},
			AllowedValues: utils.ToAny(ValidAccessLevels...),
		},
		"token_type": {
			Type:          framework.TypeString,
			Description:   "access token type",
			Required:      true,
			AllowedValues: utils.ToAny(token.ValidTokenTypes...),
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Token Type",
			},
		},
		"gitlab_revokes_token": {
			Type:        framework.TypeBool,
			Default:     false,
			Required:    false,
			Description: `Gitlab revokes the token when it's time. Vault will not revoke the token when the lease expires.`,
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Gitlab revokes token.",
			},
		},
		"config_name": {
			Type:        framework.TypeString,
			Default:     TypeConfigDefault,
			Required:    false,
			Description: "The config we use when interacting with the role, this can be specified if you want to use a specific config for the role, otherwise it uses the default one.",
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Configuration.",
			},
		},
	}
)

func (b *Backend) pathRolesList(ctx context.Context, req *logical.Request, data *framework.FieldData) (l *logical.Response, err error) {
	var roles []string
	defer func() {
		b.Logger().Debug("Available", "roles", roles, "err", err)
	}()
	l = logical.ErrorResponse("Error listing roles")
	if roles, err = req.Storage.List(ctx, fmt.Sprintf("%s/", PathRoleStorage)); err == nil {
		l = logical.ListResponse(roles)
	}
	return l, err
}

func pathListRoles(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathListRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathListRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s?/?$", PathRoleStorage),
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: operationPrefixGitlabAccessTokens,
			OperationSuffix: "roles",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.pathRolesList,
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

func (b *Backend) pathRolesDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var resp *logical.Response
	var err error
	var roleName = data.Get("role_name").(string)
	lock := locksutil.LockForKey(b.roleLocks, roleName)
	lock.RLock()
	defer lock.RUnlock()

	_, err = getRole(ctx, roleName, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error getting role: %w", err)
	}

	err = req.Storage.Delete(ctx, fmt.Sprintf("%s/%s", PathRoleStorage, roleName))
	if err != nil {
		return nil, fmt.Errorf("error deleting role: %w", err)
	}

	Event(ctx, b.Backend, "role-delete", map[string]string{
		"path":      "roles",
		"role_name": roleName,
	})

	b.Logger().Debug("Role deleted", "role", roleName)

	return resp, nil
}

func (b *Backend) pathRolesRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var roleName = data.Get("role_name").(string)

	lock := locksutil.LockForKey(b.roleLocks, roleName)
	lock.RLock()
	defer lock.RUnlock()

	role, err := getRole(ctx, roleName, req.Storage)
	if err != nil {
		return logical.ErrorResponse("error reading role"), err
	}

	if role == nil {
		return nil, nil
	}

	b.Logger().Debug("Role read", "role", roleName)

	return &logical.Response{
		Data: role.LogicalResponseData(),
	}, nil
}

func (b *Backend) pathRolesWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var roleName = data.Get("role_name").(string)
	var config *EntryConfig
	var err error
	var warnings []string
	var tokenType token.Type
	var accessLevel AccessLevel
	var configName = cmp.Or(data.Get("config_name").(string), TypeConfigDefault)

	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()
	config, err = getConfig(ctx, req.Storage, configName)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("missing %s configuration for gitlab", configName)), err
	}

	if config == nil {
		return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
	}

	tokenType, _ = token.TypeParse(data.Get("token_type").(string))
	accessLevel, _ = AccessLevelParse(data.Get("access_level").(string))

	var role = EntryRole{
		RoleName:            roleName,
		TTL:                 time.Duration(data.Get("ttl").(int)) * time.Second,
		Path:                data.Get("path").(string),
		Name:                data.Get("name").(string),
		Scopes:              data.Get("scopes").([]string),
		AccessLevel:         accessLevel,
		TokenType:           tokenType,
		GitlabRevokesTokens: data.Get("gitlab_revokes_token").(bool),
		ConfigName:          configName,
	}

	// validate name of the entry role
	if _, e := template.New("name").Funcs(tplFuncMap).Parse(role.Name); e != nil {
		err = multierror.Append(err, fmt.Errorf("invalid template %s for name: %w", role.Name, e))
	}

	// validate token type
	if !slices.Contains(token.ValidTokenTypes, tokenType.String()) {
		err = multierror.Append(err, fmt.Errorf("token_type='%s', should be one of %v: %w", data.Get("token_type").(string), token.ValidTokenTypes, errs.ErrFieldInvalidValue))
	}

	// validate access level and which fields to skip for validation
	var validAccessLevels []string
	var validScopes []string
	var noEmptyScopes bool
	var skipFields []string

	switch tokenType {
	case token.TypePersonal:
		validAccessLevels = ValidPersonalAccessLevels
		validScopes = ValidPersonalTokenScopes
		noEmptyScopes = false
		skipFields = []string{"config_name", "access_level"}
	case token.TypeGroup:
		validAccessLevels = ValidGroupAccessLevels
		validScopes = ValidGroupTokenScopes
		noEmptyScopes = false
		skipFields = []string{"config_name"}
	case token.TypeProject:
		validAccessLevels = ValidProjectAccessLevels
		validScopes = ValidProjectTokenScopes
		noEmptyScopes = false
		skipFields = []string{"config_name"}
	case token.TypeUserServiceAccount:
		validAccessLevels = ValidUserServiceAccountAccessLevels
		validScopes = ValidUserServiceAccountTokenScopes
		noEmptyScopes = false
		skipFields = []string{"config_name", "access_level"}
	case token.TypeGroupServiceAccount:
		validAccessLevels = ValidGroupServiceAccountAccessLevels
		validScopes = ValidGroupServiceAccountTokenScopes
		noEmptyScopes = false
		skipFields = []string{"config_name", "access_level"}
	case token.TypePipelineProjectTrigger:
		validAccessLevels = ValidPipelineProjectTriggerAccessLevels
		validScopes = []string{}
		noEmptyScopes = false
		skipFields = []string{"config_name", "access_level", "scopes"}
	case token.TypeProjectDeploy:
		validAccessLevels = ValidProjectDeployAccessLevels
		validScopes = ValidProjectDeployTokenScopes
		noEmptyScopes = true
		skipFields = []string{"config_name", "access_level"}
	case token.TypeGroupDeploy:
		validAccessLevels = ValidGroupDeployAccessLevels
		validScopes = ValidGroupDeployTokenScopes
		noEmptyScopes = true
		skipFields = []string{"config_name", "access_level"}
	}

	var invalidScopes []string

	// check if all required fields are set
	for name, field := range FieldSchemaRoles {
		if slices.Contains(skipFields, name) {
			continue
		}

		val, ok, _ := data.GetOkErr(name)
		if (tokenType == token.TypePersonal && name == "access_level") ||
			name == "gitlab_revokes_token" {
			continue
		}

		var required = field.Required
		if name == "ttl" && !slices.Contains([]token.Type{token.TypePipelineProjectTrigger}, tokenType) {
			required = true
		}

		if required && !ok {
			err = multierror.Append(err, fmt.Errorf("%s: %w", name, errs.ErrFieldRequired))
		} else if !required && val == nil {
			warnings = append(warnings, fmt.Sprintf("field '%s' is using expected default value of %v", name, val))
		}

		if required && name == "ttl" {
			if role.TTL > DefaultAccessTokenMaxPossibleTTL {
				err = multierror.Append(err, fmt.Errorf("ttl = %s [ttl <= max_ttl = %s]: %w", role.TTL.String(), DefaultAccessTokenMaxPossibleTTL, errs.ErrInvalidValue))
			}
			if role.GitlabRevokesTokens && role.TTL < 24*time.Hour {
				err = multierror.Append(err, fmt.Errorf("ttl = %s [%s <= ttl <= %s]: %w", role.TTL, DefaultAccessTokenMinTTL, DefaultAccessTokenMaxPossibleTTL, errs.ErrInvalidValue))
			}
			if !role.GitlabRevokesTokens && role.TTL < time.Hour {
				err = multierror.Append(err, fmt.Errorf("ttl = %s [ttl >= 1h]: %w", role.TTL, errs.ErrInvalidValue))
			}
		}
	}

	if !slices.Contains(validAccessLevels, accessLevel.String()) {
		err = multierror.Append(err, fmt.Errorf("access_level='%s', should be one of %v: %w", data.Get("access_level").(string), validAccessLevels, errs.ErrFieldInvalidValue))
	}

	for _, scope := range role.Scopes {
		if !slices.Contains(validScopes, scope) {
			invalidScopes = append(invalidScopes, scope)
		}
	}

	if len(invalidScopes) > 0 {
		err = multierror.Append(err, fmt.Errorf("scopes='%v', should be one or more of '%v': %w", invalidScopes, validScopes, errs.ErrFieldInvalidValue))
	}

	if noEmptyScopes && len(role.Scopes) == 0 {
		err = multierror.Append(err, fmt.Errorf("should be one or more of '%v': %w", validScopes, errs.ErrFieldInvalidValue))
	}

	if tokenType == token.TypeUserServiceAccount && (config.Type == gitlab.TypeSaaS || config.Type == gitlab.TypeDedicated) {
		err = multierror.Append(err, fmt.Errorf("cannot create %s with %s: %w", tokenType, config.Type, errs.ErrInvalidValue))
	}

	if err != nil {
		return logical.ErrorResponse(err.Error()), err
	}

	lock := locksutil.LockForKey(b.roleLocks, roleName)
	lock.Lock()
	defer lock.Unlock()

	entry, err := logical.StorageEntryJSON(fmt.Sprintf("%s/%s", PathRoleStorage, role.RoleName), role)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	Event(ctx, b.Backend, "role-write", map[string]string{
		"path":        "roles",
		"role_name":   roleName,
		"config_name": role.ConfigName,
	})

	b.Logger().Debug("Role written", "role", roleName)

	return &logical.Response{
		Data:     role.LogicalResponseData(),
		Warnings: warnings,
	}, nil
}

func (b *Backend) pathRoleExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	name := data.Get("role_name").(string)
	role, err := getRole(ctx, name, req.Storage)
	if err != nil {
		if strings.Contains(err.Error(), logical.ErrReadOnly.Error()) {
			return false, nil
		}

		return false, fmt.Errorf("error reading role: %w", err)
	}

	return role != nil, nil
}

func pathRoles(b *Backend) *framework.Path {
	return &framework.Path{
		HelpSynopsis:    strings.TrimSpace(pathRolesHelpSyn),
		HelpDescription: strings.TrimSpace(pathRolesHelpDesc),
		Pattern:         fmt.Sprintf("%s/%s", PathRoleStorage, framework.GenericNameRegex("role_name")),
		Fields:          FieldSchemaRoles,
		DisplayAttrs: &framework.DisplayAttributes{
			OperationPrefix: operationPrefixGitlabAccessTokens,
			OperationSuffix: "role",
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathRolesDelete,
				Summary:  "Deletes a role",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathRolesWrite,
				Summary:  "Creates a new role",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathRolesWrite,
				Summary:  "Updates an existing role",
				Responses: map[int][]framework.Response{
					http.StatusNoContent: {{
						Description: http.StatusText(http.StatusNoContent),
					}},
				},
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathRolesRead,
				Summary:  "Reads an existing role",
				Responses: map[int][]framework.Response{
					http.StatusNotFound: {{
						Description: http.StatusText(http.StatusNotFound),
					}},
					http.StatusOK: {{
						Fields: FieldSchemaRoles,
					}},
				},
			},
		},
		ExistenceCheck: b.pathRoleExistenceCheck,
	}
}

const (
	pathRolesHelpSyn  = `Create a role with parameters that are used to generate a various access tokens.`
	pathRolesHelpDesc = `
This path allows you to create a role with predefined parameters that will be used to generate tokens for different 
access types in GitLab. The role defines the configuration for generating project, group, personal access tokens,
user service accounts, or group service accounts.`
	pathListRolesHelpSyn  = `Lists existing roles`
	pathListRolesHelpDesc = `
This path allows you to list all available roles that have been created within the GitLab Access Tokens Backend. 
Each role defines a set of parameters, such as token permissions, scopes, and expiration settings, which are used 
when generating access tokens.`
)
