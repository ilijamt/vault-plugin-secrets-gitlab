package gitlab

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type EntryRole struct {
	RoleName            string        `json:"role_name" structs:"role_name" mapstructure:"role_name"`
	TTL                 time.Duration `json:"ttl" structs:"ttl" mapstructure:"ttl"`
	Path                string        `json:"path" structs:"path" mapstructure:"path"`
	Name                string        `json:"name" structs:"name" mapstructure:"name"`
	Scopes              []string      `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	AccessLevel         AccessLevel   `json:"access_level" structs:"access_level" mapstructure:"access_level,omitempty"`
	TokenType           TokenType     `json:"token_type" structs:"token_type" mapstructure:"token_type"`
	GitlabRevokesTokens bool          `json:"gitlab_revokes_token" structs:"gitlab_revokes_token" mapstructure:"gitlab_revokes_token"`
	ConfigName          string        `json:"config_name" structs:"config_name" mapstructure:"config_name"`
}

func (e *EntryRole) LogicalResponseData() map[string]any {
	return map[string]any{
		"role_name":            e.RoleName,
		"path":                 e.Path,
		"name":                 e.Name,
		"scopes":               e.Scopes,
		"access_level":         e.AccessLevel.String(),
		"ttl":                  int64(e.TTL / time.Second),
		"token_type":           e.TokenType.String(),
		"gitlab_revokes_token": e.GitlabRevokesTokens,
		"config_name":          e.ConfigName,
	}
}

func (e *EntryRole) Merge(data *framework.FieldData, gitlabType Type) (warnings []string, changes map[string]string, err error) {
	if data == nil {
		return warnings, changes, multierror.Append(fmt.Errorf("data: %w", ErrNilValue))
	}

	if err = data.Validate(); err != nil {
		return warnings, changes, multierror.Append(err)
	}

	changes = make(map[string]string)

	if _, ok := data.GetOk("path"); ok {
		e.Path = data.Get("path").(string)
		changes["path"] = e.Path
	}

	if v, ok := data.GetOk("gitlab_revokes_tokens"); ok {
		e.GitlabRevokesTokens = v.(bool)
		changes["gitlab_revokes_tokens"] = strconv.FormatBool(e.GitlabRevokesTokens)
	}

	if _, ok := data.GetOk("config_name"); ok {
		e.ConfigName = cmp.Or(data.Get("config_name").(string), TypeConfigDefault)
		changes["config_name"] = e.ConfigName
	}

	if val, ok := data.GetOk("name"); ok {
		if _, er := template.New("name").Funcs(tplFuncMap).Parse(val.(string)); er != nil {
			err = multierror.Append(err, fmt.Errorf("invalid template %s for name: %w", val, er))
		} else {
			e.Name = val.(string)
			changes["name"] = e.Name
		}
	}

	if _, ok := data.GetOk("ttl"); ok {
		if er := e.updateTTL(data, e.GitlabRevokesTokens); er == nil {
			changes["ttl"] = e.TTL.String()
		} else {
			err = multierror.Append(err, er)
		}
	}

	if _, ok := data.GetOk("token_type"); ok {
		if er := e.updateTokenType(data, gitlabType); er == nil {
			changes["token_type"] = e.TokenType.String()
		} else {
			err = multierror.Append(err, er)
		}
	}

	if _, ok := data.GetOk("access_level"); ok {
		if er := e.updateAccessLevel(data, e.TokenType); er == nil {
			changes["access_level"] = e.AccessLevel.String()
		} else {
			err = multierror.Append(err, er)
		}
	}

	// if access_level, token_type or scopes change we need to recalculate the whole thing

	return warnings, changes, err
}

func (e *EntryRole) updateTTL(data *framework.FieldData, gitlabRevokesTokens bool) (err error) {
	if val, ok := data.GetOk("ttl"); ok {
		ttl := time.Duration(val.(int)) * time.Second

		if ttl > DefaultAccessTokenMaxPossibleTTL {
			err = multierror.Append(err, fmt.Errorf("ttl = %s [ttl <= max_ttl = %s]: %w", ttl.String(), DefaultAccessTokenMaxPossibleTTL, ErrInvalidValue))
		}

		if gitlabRevokesTokens && ttl < 24*time.Hour {
			err = multierror.Append(err, fmt.Errorf("ttl = %s [%s <= ttl <= %s]: %w", ttl, DefaultAccessTokenMinTTL, DefaultAccessTokenMaxPossibleTTL, ErrInvalidValue))
		}

		if !gitlabRevokesTokens && ttl < time.Hour {
			err = multierror.Append(err, fmt.Errorf("ttl = %s [ttl >= 1h]: %w", ttl, ErrInvalidValue))
		}

		if err == nil {
			e.TTL = ttl
		}
	} else {
		err = multierror.Append(err, fmt.Errorf("ttl: %w", ErrFieldRequired))
	}
	return err
}

func (e *EntryRole) updateTokenType(data *framework.FieldData, gitlabType Type) (err error) {
	var tokenType TokenType
	if val, ok := data.GetOk("token_type"); ok {
		if tokenType, err = TokenTypeParse(val.(string)); err == nil {
			if slices.Contains(validTokenTypes, tokenType.String()) {
				e.TokenType = tokenType
			} else {
				err = multierror.Append(err, fmt.Errorf("token_type='%s', should be one of %v: %w", tokenType.String(), validTokenTypes, ErrFieldInvalidValue))
			}
		}
	} else {
		err = multierror.Append(err, fmt.Errorf("token_type: %w", ErrFieldRequired))
	}

	if tokenType == TokenTypeUserServiceAccount && (gitlabType == TypeSaaS || gitlabType == TypeDedicated) {
		err = multierror.Append(err, fmt.Errorf("cannot create %s with %s: %w", tokenType, gitlabType, ErrInvalidValue))
	}

	return err
}

func (e *EntryRole) updateAccessLevel(data *framework.FieldData, tokenType TokenType) (err error) {
	if val, ok := data.GetOk("access_level"); ok {
		var validAccessLevels []string

		switch tokenType {
		case TokenTypePersonal:
			validAccessLevels = ValidPersonalAccessLevels
		case TokenTypeGroup:
			validAccessLevels = ValidGroupAccessLevels
		case TokenTypeProject:
			validAccessLevels = ValidProjectAccessLevels
		case TokenTypeUserServiceAccount:
			validAccessLevels = ValidUserServiceAccountAccessLevels
		case TokenTypeGroupServiceAccount:
			validAccessLevels = ValidGroupServiceAccountAccessLevels
		default:
			return multierror.Append(fmt.Errorf("unknown '%s' token type: %w", tokenType.String(), ErrFieldInvalidValue))
		}

		if !slices.Contains(validAccessLevels, val.(string)) {
			err = multierror.Append(err, fmt.Errorf("access_level='%s', should be one of %v: %w", val.(string), validAccessLevels, ErrFieldInvalidValue))
		}
	}

	return err
}

func (e *EntryRole) updateScopes(data *framework.FieldData, tokenType TokenType) (err error) {
	if val, ok := data.GetOk("scopes"); ok {
		var scopes = val.([]string)
		var invalidScopes []string
		var validScopes = validTokenScopes

		switch tokenType {
		case TokenTypePersonal:
			validScopes = append(validScopes, ValidPersonalTokenScopes...)
		case TokenTypeUserServiceAccount:
			validScopes = append(validScopes, ValidUserServiceAccountTokenScopes...)
		case TokenTypeGroupServiceAccount:
			validScopes = append(validScopes, ValidGroupServiceAccountTokenScopes...)
		}

		for _, scope := range scopes {
			if !slices.Contains(validScopes, scope) {
				invalidScopes = append(invalidScopes, scope)
			}
		}

		if len(invalidScopes) > 0 {
			err = multierror.Append(err, fmt.Errorf("scopes='%v', should be one or more of '%v': %w", invalidScopes, validScopes, ErrFieldInvalidValue))
		}
	}

	return err
}

func (e *EntryRole) UpdateFromFieldData(data *framework.FieldData, gitlabType Type) (warnings []string, err error) {
	if data == nil {
		return warnings, multierror.Append(fmt.Errorf("data: %w", ErrNilValue))
	}

	if err = data.Validate(); err != nil {
		return warnings, multierror.Append(err)
	}

	e.RoleName = data.Get("role_name").(string)
	e.GitlabRevokesTokens = data.Get("gitlab_revokes_token").(bool)
	e.Path = data.Get("path").(string)
	e.ConfigName = cmp.Or(data.Get("config_name").(string), TypeConfigDefault)

	err = multierror.Append(err, e.updateTTL(data, e.GitlabRevokesTokens))

	if val, ok := data.GetOk("name"); ok {
		if _, er := template.New("name").Funcs(tplFuncMap).Parse(val.(string)); er != nil {
			err = multierror.Append(err, fmt.Errorf("invalid template %s for name: %w", val, er))
		} else {
			e.Name = val.(string)
		}
	} else {
		err = multierror.Append(err, fmt.Errorf("config_name: %w", ErrFieldRequired))
	}

	return warnings, multierror.Append(err,
		e.updateTokenType(data, gitlabType),
		e.updateAccessLevel(data, e.TokenType),
		e.updateScopes(data, e.TokenType),
	)
}

func getRole(ctx context.Context, name string, s logical.Storage) (role *EntryRole, err error) {
	var entry *logical.StorageEntry
	if entry, err = s.Get(ctx, fmt.Sprintf("%s/%s", PathRoleStorage, name)); err == nil {
		if entry == nil {
			return nil, nil
		}
		role = new(EntryRole)
		_ = entry.DecodeJSON(role)
	}
	return role, err
}
