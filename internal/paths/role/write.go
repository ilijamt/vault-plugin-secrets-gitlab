package role

import (
	"cmp"
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	modelRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func (p *Provider) pathRolesWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var roleName = data.Get("role_name").(string)
	var config *modelConfig.EntryConfig
	var err error
	var warnings []string
	var tokenType token.Type
	var accessLevel token.AccessLevel
	var configName = cmp.Or(data.Get("config_name").(string), backend.DefaultConfigName)

	config, err = p.b.GetConfig(ctx, req.Storage, configName)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("missing %s configuration for gitlab", configName)), err
	}

	if config == nil {
		return logical.ErrorResponse(errs.ErrBackendNotConfigured.Error()), nil
	}

	tokenType, _ = token.ParseType(data.Get("token_type").(string))
	accessLevel, _ = token.ParseAccessLevel(data.Get("access_level").(string))

	var role = modelRole.Role{
		RoleName:            roleName,
		TTL:                 time.Duration(data.Get("ttl").(int)) * time.Second,
		Path:                data.Get("path").(string),
		Name:                data.Get("name").(string),
		Scopes:              data.Get("scopes").([]string),
		DynamicPath:         data.Get("dynamic_path").(bool),
		AccessLevel:         accessLevel,
		TokenType:           tokenType,
		GitlabRevokesTokens: data.Get("gitlab_revokes_token").(bool),
		ConfigName:          configName,
	}

	// validate the name of the entry role
	if e := utils.ValidateTokenNameName(role); e != nil {
		err = multierror.Append(err, fmt.Errorf("invalid template %s for name: %w", role.Name, e))
	}

	if role.DynamicPath {
		// if we have a dynamic path, and we can override the path, validate the regexp that it compiles
		// this is required as during token creation we will validate the path using this regexp
		if _, err = regexp.Compile(role.Path); err != nil {
			err = multierror.Append(err, fmt.Errorf("invalid regexp %s for path: %w", role.Path, errs.ErrInvalidValue))
		}
	} else {
		// validate the path that it confirms to the correct format for the given
		if !token.IsValidPath(role.Path, role.TokenType) {
			err = multierror.Append(err, fmt.Errorf("invalid path %s for token type %s: %w", role.Path, role.TokenType, errs.ErrInvalidValue))
		}
	}

	// validate token type
	if !slices.Contains(token.ValidTokenTypes, tokenType.String()) {
		err = multierror.Append(err, fmt.Errorf("token_type='%s', should be one of %v: %w", data.Get("token_type").(string), token.ValidTokenTypes, errs.ErrFieldInvalidValue))
	}

	// version-aware validation tables: empty config.GitlabVersion is lenient
	// (every known value is accepted) so first-write roles aren't blocked.
	gitlabVersion := config.GitlabVersion
	allowedAccessLevels, accessLevelApplicable := token.ValidAccessLevelsFor(tokenType, gitlabVersion)
	allowedScopes, scopesApplicable := token.ValidScopesFor(tokenType, gitlabVersion)

	// noEmptyScopes: deploy tokens require at least one scope.
	var noEmptyScopes bool
	var skipFields []string

	switch tokenType {
	case token.TypePersonal, token.TypeUserServiceAccount, token.TypeGroupServiceAccount:
		skipFields = []string{"config_name", "access_level"}
	case token.TypeGroup, token.TypeProject:
		skipFields = []string{"config_name"}
	case token.TypePipelineProjectTrigger:
		skipFields = []string{"config_name", "access_level", "scopes"}
	case token.TypeProjectDeploy, token.TypeGroupDeploy:
		noEmptyScopes = true
		skipFields = []string{"config_name", "access_level"}
	}

	// always skip these fields
	skipFields = append(skipFields, "dynamic_path")

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
			if role.TTL > backend.DefaultAccessTokenMaxPossibleTTL {
				err = multierror.Append(err, fmt.Errorf("ttl = %s [ttl <= max_ttl = %s]: %w", role.TTL.String(), backend.DefaultAccessTokenMaxPossibleTTL, errs.ErrInvalidValue))
			}
			if role.GitlabRevokesTokens && role.TTL < 24*time.Hour {
				err = multierror.Append(err, fmt.Errorf("ttl = %s [%s <= ttl <= %s]: %w", role.TTL, backend.DefaultAccessTokenMinTTL, backend.DefaultAccessTokenMaxPossibleTTL, errs.ErrInvalidValue))
			}
			if !role.GitlabRevokesTokens && role.TTL < time.Hour {
				err = multierror.Append(err, fmt.Errorf("ttl = %s [ttl >= 1h]: %w", role.TTL, errs.ErrInvalidValue))
			}
		}
	}

	if !accessLevelApplicable {
		// Token type does not take an access_level — only the empty value is allowed.
		if accessLevel != token.AccessLevelUnknown {
			err = multierror.Append(err, fmt.Errorf("access_level='%s', should be one of %v: %w", data.Get("access_level").(string), allowedAccessLevels, errs.ErrFieldInvalidValue))
		}
	} else if !token.IsAccessLevelAllowed(tokenType, accessLevel, gitlabVersion) {
		err = multierror.Append(err, fmt.Errorf("access_level='%s' not allowed for token_type='%s' on gitlab %s, should be one of %v: %w", data.Get("access_level").(string), tokenType, gitlabVersion, allowedAccessLevels, errs.ErrFieldInvalidValue))
	}

	if !scopesApplicable {
		if len(role.Scopes) > 0 {
			err = multierror.Append(err, fmt.Errorf("token_type='%s' does not support scopes: %w", tokenType, errs.ErrFieldInvalidValue))
		}
	} else {
		for _, scope := range role.Scopes {
			if !token.IsScopeAllowed(tokenType, token.Scope(scope), gitlabVersion) {
				invalidScopes = append(invalidScopes, scope)
			}
		}
		if len(invalidScopes) > 0 {
			err = multierror.Append(err, fmt.Errorf("scopes='%v' not allowed for token_type='%s' on gitlab %s, should be one or more of '%v': %w", invalidScopes, tokenType, gitlabVersion, allowedScopes, errs.ErrFieldInvalidValue))
		}
		if noEmptyScopes && len(role.Scopes) == 0 {
			err = multierror.Append(err, fmt.Errorf("should be one or more of '%v': %w", allowedScopes, errs.ErrFieldInvalidValue))
		}
	}

	if tokenType == token.TypeUserServiceAccount && (config.Type == gitlabTypes.TypeSaaS || config.Type == gitlabTypes.TypeDedicated) {
		err = multierror.Append(err, fmt.Errorf("cannot create %s with %s: %w", tokenType, config.Type, errs.ErrInvalidValue))
	}

	if err != nil {
		return logical.ErrorResponse(err.Error()), err
	}

	lock := p.b.LockForKey("role", roleName)
	lock.Lock()
	defer lock.Unlock()

	entry, err := logical.StorageEntryJSON(fmt.Sprintf("%s/%s", backend.PathRoleStorage, role.RoleName), role)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	_ = p.b.SendEvent(ctx, eventWrite, map[string]string{
		"path":         "roles",
		"role_name":    roleName,
		"config_name":  role.ConfigName,
		"role_path":    role.Path,
		"dynamic_path": strconv.FormatBool(role.DynamicPath),
	})

	p.b.Logger().Debug("Role written", "role", roleName)

	return &logical.Response{
		Data:     role.LogicalResponseData(),
		Warnings: warnings,
	}, nil
}
