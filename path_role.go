package gitlab

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
	"golang.org/x/exp/slices"
	"net/http"
	"strings"
	"time"
)

const (
	PathRoleStorage = "roles"
)

var (
	fieldSchemaRoles = map[string]*framework.FieldSchema{
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
			AllowedValues: allowedValues(append(validTokenScopes, ValidPersonalTokenScopes...)...),
		},
		"token_ttl": {
			Type:        framework.TypeDurationSecond,
			Description: "The TTL of the token",
			Required:    false,
			Default:     DefaultRoleFieldAccessTokenMaxTTL,
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
			AllowedValues: allowedValues(ValidAccessLevels...),
		},
		"token_type": {
			Type:          framework.TypeString,
			Description:   "access token type",
			Required:      true,
			AllowedValues: allowedValues(validTokenTypes...),
			DisplayAttrs: &framework.DisplayAttributes{
				Name: "Token Type",
			},
		},
	}
)

func (b *Backend) pathRolesList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roles, err := req.Storage.List(ctx, fmt.Sprintf("%s/", PathRoleStorage))
	if err != nil {
		return logical.ErrorResponse("Error listing roles"), err
	}
	b.Logger().Debug("Available roles input the system", "roles", roles)
	return logical.ListResponse(roles), nil
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
							"role_name": fieldSchemaRoles["role_name"],
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
	var roleName string

	if roleName = data.Get("role_name").(string); roleName == "" {
		return logical.ErrorResponse("Unable to delete, missing role name"), nil
	}

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

	event(ctx, b.Backend, "role-delete", map[string]string{
		"path":      "roles",
		"role_name": roleName,
	})

	b.Logger().Debug("Role deleted", "role", roleName)

	return resp, nil
}

func (b *Backend) pathRolesRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var roleName string
	if roleName = data.Get("role_name").(string); roleName == "" {
		return logical.ErrorResponse("Unable to read, missing role name"), nil
	}

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
	var roleName string
	if roleName = data.Get("role_name").(string); roleName == "" {
		return logical.ErrorResponse("Unable to write, missing role name"), nil
	}

	var config *EntryConfig
	var err error
	var warnings []string
	var tokenType TokenType
	var accessLevel AccessLevel

	b.lockClientMutex.RLock()
	defer b.lockClientMutex.RUnlock()
	config, err = getConfig(ctx, req.Storage)
	if err != nil {
		return logical.ErrorResponse("missing configuration for gitlab"), err
	}

	if config == nil {
		return logical.ErrorResponse(ErrBackendNotConfigured.Error()), nil
	}

	tokenType, _ = TokenTypeParse(data.Get("token_type").(string))
	accessLevel, _ = AccessLevelParse(data.Get("access_level").(string))

	var role = entryRole{
		RoleName:    roleName,
		TokenTTL:    time.Duration(data.Get("token_ttl").(int)) * time.Second,
		Path:        data.Get("path").(string),
		Name:        data.Get("name").(string),
		Scopes:      data.Get("scopes").([]string),
		AccessLevel: accessLevel,
		TokenType:   tokenType,
	}

	// validate token type
	if !slices.Contains(validTokenTypes, tokenType.String()) {
		err = multierror.Append(err, fmt.Errorf("token_type='%s', should be one of %v: %w", data.Get("token_type").(string), validTokenTypes, ErrFieldInvalidValue))
	}

	// check if all required fields are set
	for name, field := range fieldSchemaRoles {
		val, ok, _ := data.GetOkErr(name)
		if tokenType == TokenTypePersonal && name == "access_level" {
			continue
		}
		if field.Required && !ok {
			err = multierror.Append(err, fmt.Errorf("%s: %w", name, ErrFieldRequired))
		} else if !field.Required && val == nil {
			warnings = append(warnings, fmt.Sprintf("field '%s' is using expected default value of %v", name, val))
		}
	}

	// validate ttl to make sure it's less than max ttl in config, and above 24h
	if role.TokenTTL > 0 && role.TokenTTL < DefaultAccessTokenMinTTL {
		warnings = append(warnings, "token_ttl is set with less than 24 hours. With current token expiry limitation, this token_ttl is ignored, it will be set to 24 hours")
		role.TokenTTL = DefaultAccessTokenMinTTL
	} else if role.TokenTTL > config.MaxTTL && config.MaxTTL > 0 {
		warnings = append(warnings, fmt.Sprintf("token_ttl needs to be less than %s, setting 'token_ttl' to %s", config.MaxTTL, config.MaxTTL))
		role.TokenTTL = config.MaxTTL
	} else if role.TokenTTL <= 0 {
		role.TokenTTL = config.MaxTTL
		warnings = append(warnings, fmt.Sprintf("token_ttl is set to 0. Tokens without ttls are not supported since Gitlab 16.0 setting to %d based on config max_ttl", config.MaxTTL))
	}

	// validate access level
	var validAccessLevels []string
	switch tokenType {
	case TokenTypePersonal:
		validAccessLevels = ValidPersonalAccessLevels
	case TokenTypeGroup:
		validAccessLevels = ValidGroupAccessLevels
	case TokenTypeProject:
		validAccessLevels = ValidProjectAccessLevels
	}

	if !slices.Contains(validAccessLevels, accessLevel.String()) {
		err = multierror.Append(err, fmt.Errorf("access_level='%s', should be one of %v: %w", data.Get("access_level").(string), validAccessLevels, ErrFieldInvalidValue))
	}

	// validate scopes
	var invalidScopes []string
	var validScopes = validTokenScopes
	if tokenType == TokenTypePersonal {
		validScopes = append(validScopes, ValidPersonalTokenScopes...)
	}
	for _, scope := range role.Scopes {
		if !slices.Contains(validScopes, scope) {
			invalidScopes = append(invalidScopes, scope)
		}
	}

	if len(invalidScopes) > 0 {
		err = multierror.Append(err, fmt.Errorf("scopes='%v', should be one or more of '%v': %w", invalidScopes, validScopes, ErrFieldInvalidValue))
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

	event(ctx, b.Backend, "role-write", map[string]string{
		"path":      "roles",
		"role_name": roleName,
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
		Pattern:         fmt.Sprintf("%s/%s", PathRoleStorage, framework.GenericNameWithAtRegex("role_name")),
		Fields:          fieldSchemaRoles,
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
						Fields: fieldSchemaRoles,
					}},
				},
			},
		},
		ExistenceCheck: b.pathRoleExistenceCheck,
	}
}

const (
	pathRolesHelpSyn      = `Create a role with parameters that are used to generate a project, group or personal access token.`
	pathRolesHelpDesc     = `This path allows you to create a role whose parameters will be used to generate a project, group or personal access access token.`
	pathListRolesHelpSyn  = `Lists existing roles`
	pathListRolesHelpDesc = `This path allows you to list all available roles.`
)
