package config

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

var _ model.Named = (*EntryConfig)(nil)

type InstanceInfo struct {
	TokenOwnerIsAdmin           bool   `json:"token_owner_is_admin" structs:"token_owner_is_admin" mapstructure:"token_owner_is_admin"`
	SettingsKnown               bool   `json:"settings_known" structs:"settings_known" mapstructure:"settings_known"`
	EnforceGranularTokens       bool   `json:"enforce_granular_tokens" structs:"enforce_granular_tokens" mapstructure:"enforce_granular_tokens"`
	GranularTokensEnforcedAfter string `json:"granular_tokens_enforced_after" structs:"granular_tokens_enforced_after" mapstructure:"granular_tokens_enforced_after"`
	MaxTokenLifetimeDays        int64  `json:"max_token_lifetime_days" structs:"max_token_lifetime_days" mapstructure:"max_token_lifetime_days"`
	RequireTokenExpiry          bool   `json:"require_token_expiry" structs:"require_token_expiry" mapstructure:"require_token_expiry"`
	ServiceTokenExpiryEnforced  bool   `json:"service_token_expiry_enforced" structs:"service_token_expiry_enforced" mapstructure:"service_token_expiry_enforced"`
	PersonalTokensDisabled      *bool  `json:"personal_tokens_disabled" structs:"personal_tokens_disabled" mapstructure:"personal_tokens_disabled"`
}

type EntryConfig struct {
	TokenId            int64            `json:"token_id" yaml:"token_id" mapstructure:"token_id"`
	BaseURL            string           `json:"base_url" structs:"base_url" mapstructure:"base_url"`
	Token              string           `json:"token" structs:"token" mapstructure:"token"`
	AutoRotateToken    bool             `json:"auto_rotate_token" structs:"auto_rotate_token" mapstructure:"auto_rotate_token"`
	AutoRotateBefore   time.Duration    `json:"auto_rotate_before" structs:"auto_rotate_before" mapstructure:"auto_rotate_before"`
	TokenCreatedAt     time.Time        `json:"token_created_at" structs:"token_created_at" mapstructure:"token_created_at"`
	TokenExpiresAt     time.Time        `json:"token_expires_at" structs:"token_expires_at" mapstructure:"token_expires_at"`
	Scopes             []string         `json:"scopes" structs:"scopes" mapstructure:"scopes"`
	Type               gitlabTypes.Type `json:"type" structs:"type" mapstructure:"type"`
	Name               string           `json:"name" structs:"name" mapstructure:"name"`
	GitlabVersion      string           `json:"gitlab_version" structs:"gitlab_version" mapstructure:"gitlab_version"`
	GitlabRevision     string           `json:"gitlab_revision" structs:"gitlab_revision" mapstructure:"gitlab_revision"`
	GitlabIsEnterprise bool             `json:"gitlab_is_enterprise" structs:"gitlab_is_enterprise" mapstructure:"gitlab_is_enterprise"`
	GitlabInstance     InstanceInfo     `json:"gitlab_instance" structs:"gitlab_instance" mapstructure:"gitlab_instance"`
	GitlabRefreshedAt  time.Time        `json:"gitlab_refreshed_at" structs:"gitlab_refreshed_at" mapstructure:"gitlab_refreshed_at"`
}

func (e *EntryConfig) GetName() string { return e.Name }

func (e *EntryConfig) Merge(data *framework.FieldData) (warnings []string, changes map[string]string, err error) {
	var er error
	if data == nil {
		return warnings, changes, multierror.Append(fmt.Errorf("data: %w", errs.ErrNilValue))
	}

	if err = data.Validate(); err != nil {
		return warnings, changes, multierror.Append(err)
	}

	changes = make(map[string]string)

	if val, ok := data.GetOk("auto_rotate_token"); ok {
		e.AutoRotateToken = val.(bool)
		changes["auto_rotate_token"] = strconv.FormatBool(e.AutoRotateToken)
	}

	if typ, ok := data.GetOk("type"); ok {
		var pType gitlabTypes.Type
		if pType, er = gitlabTypes.TypeParse(typ.(string)); er != nil {
			err = multierror.Append(err, er)
		} else {
			e.Type = pType
			changes["type"] = pType.String()
		}
	}

	if _, ok := data.GetOk("auto_rotate_before"); ok {
		w, er := e.updateAutoRotateBefore(data)
		if er != nil {
			err = multierror.Append(err, er.Errors...)
		} else {
			changes["auto_rotate_before"] = e.AutoRotateBefore.String()
		}
		warnings = append(warnings, w...)
	}

	if val, ok := data.GetOk("base_url"); ok && len(val.(string)) > 0 {
		e.BaseURL = val.(string)
		changes["base_url"] = e.BaseURL
	}

	if val, ok := data.GetOk("token"); ok && len(val.(string)) > 0 {
		e.Token = val.(string)
		changes["token"] = strings.Repeat("*", len(e.Token))
	}

	return warnings, changes, err
}

func (e *EntryConfig) updateAutoRotateBefore(data *framework.FieldData) (warnings []string, err *multierror.Error) {
	if val, ok := data.GetOk("auto_rotate_before"); ok {
		atr, _ := utils.ConvertToInt(val)
		switch {
		case atr > int(DefaultAutoRotateBeforeMaxTTL.Seconds()):
			err = multierror.Append(err, fmt.Errorf("auto_rotate_before can not be bigger than %s: %w", DefaultAutoRotateBeforeMaxTTL, errs.ErrInvalidValue))
		case atr <= int(DefaultAutoRotateBeforeMinTTL.Seconds())-1:
			err = multierror.Append(err, fmt.Errorf("auto_rotate_before can not be less than %s: %w", DefaultAutoRotateBeforeMinTTL, errs.ErrInvalidValue))
		default:
			e.AutoRotateBefore = time.Duration(atr) * time.Second
		}
	} else {
		e.AutoRotateBefore = DefaultAutoRotateBeforeMinTTL
		warnings = append(warnings, fmt.Sprintf("auto_rotate_before not specified setting to %s", DefaultAutoRotateBeforeMinTTL))
	}
	return warnings, err
}

func (e *EntryConfig) UpdateFromFieldData(data *framework.FieldData) (warnings []string, err error) {
	if data == nil {
		return warnings, multierror.Append(fmt.Errorf("data: %w", errs.ErrNilValue))
	}

	if err = data.Validate(); err != nil {
		return warnings, multierror.Append(err)
	}

	var er error
	e.AutoRotateToken = data.Get("auto_rotate_token").(bool)

	if token, ok := data.GetOk("token"); ok && len(token.(string)) > 0 {
		e.Token = token.(string)
	} else {
		err = multierror.Append(err, fmt.Errorf("token: %w", errs.ErrFieldRequired))
	}

	if typ, ok := data.GetOk("type"); ok {
		if e.Type, er = gitlabTypes.TypeParse(typ.(string)); er != nil {
			err = multierror.Append(err, er)
		}
	} else {
		err = multierror.Append(err, fmt.Errorf("gitlab type: %w", errs.ErrFieldRequired))
	}

	if baseUrl, ok := data.GetOk("base_url"); ok && len(baseUrl.(string)) > 0 {
		e.BaseURL = baseUrl.(string)
	} else {
		err = multierror.Append(err, fmt.Errorf("base_url: %w", errs.ErrFieldRequired))
	}

	{
		w, er := e.updateAutoRotateBefore(data)
		if er != nil {
			err = multierror.Append(err, er.Errors...)
		}
		warnings = append(warnings, w...)
	}

	return warnings, err
}

func (e *EntryConfig) LogicalResponseData(includeToken bool) (data map[string]any) {
	var tokenExpiresAt, tokenCreatedAt, refreshedAt = "", "", ""
	if !e.TokenExpiresAt.IsZero() {
		tokenExpiresAt = e.TokenExpiresAt.Format(time.RFC3339)
	}
	if !e.TokenCreatedAt.IsZero() {
		tokenCreatedAt = e.TokenCreatedAt.Format(time.RFC3339)
	}

	if !e.GitlabRefreshedAt.IsZero() {
		refreshedAt = e.GitlabRefreshedAt.Format(time.RFC3339)
	}

	data = map[string]any{
		"base_url":             e.BaseURL,
		"auto_rotate_token":    e.AutoRotateToken,
		"auto_rotate_before":   e.AutoRotateBefore.String(),
		"token_id":             e.TokenId,
		"gitlab_version":       e.GitlabVersion,
		"gitlab_revision":      e.GitlabRevision,
		"gitlab_is_enterprise": e.GitlabIsEnterprise,
		"gitlab_instance":      e.GitlabInstance,
		"last_refreshed_at":    refreshedAt,
		"token_created_at":     tokenCreatedAt,
		"token_expires_at":     tokenExpiresAt,
		"token_sha1_hash":      fmt.Sprintf("%x", sha1.Sum([]byte(e.Token))),
		"scopes":               strings.Join(e.Scopes, ", "),
		"type":                 e.Type.String(),
		"name":                 e.Name,
	}

	if includeToken {
		data["token"] = e.Token
	}

	return data
}
