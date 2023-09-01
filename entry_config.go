package gitlab

import (
	"context"
	"github.com/hashicorp/vault/sdk/logical"
	"time"
)

type entryConfig struct {
	BaseURL          string        `json:"base_url" structs:"base_url" mapstructure:"base_url"`
	Token            string        `json:"token" structs:"token" mapstructure:"token"`
	MaxTTL           time.Duration `json:"max_ttl" structs:"max_ttl" mapstructure:"max_ttl"`
	AutoRotateToken  bool          `json:"auto_rotate_token" structs:"auto_rotate_token" mapstructure:"auto_rotate_token"`
	AutoRotateBefore time.Duration `json:"auto_rotate_before" structs:"auto_rotate_before" mapstructure:"auto_rotate_before"`
}

func (e entryConfig) LogicalResponseData() map[string]interface{} {
	return map[string]interface{}{
		"max_ttl":            int64(e.MaxTTL / time.Second),
		"base_url":           e.BaseURL,
		"token":              e.Token,
		"auto_rotate_token":  e.AutoRotateToken,
		"auto_rotate_before": e.AutoRotateBefore.String(),
	}
}

func getConfig(ctx context.Context, s logical.Storage) (*entryConfig, error) {
	entry, err := s.Get(ctx, PathConfigStorage)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	cfg := new(entryConfig)
	if err := entry.DecodeJSON(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
