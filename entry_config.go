package gitlab

import (
	"context"
	"github.com/hashicorp/vault/sdk/logical"
	"time"
)

type entryConfig struct {
	BaseURL string        `json:"base_url" structs:"base_url" mapstructure:"base_url"`
	Token   string        `json:"token" structs:"token" mapstructure:"token"`
	MaxTTL  time.Duration `json:"max_ttl" structs:"max_ttl" mapstructure:"max_ttl"`
}

func (e entryConfig) LogicalResponseData() map[string]interface{} {
	return map[string]interface{}{
		"max_ttl":  int64(e.MaxTTL / time.Second),
		"base_url": e.BaseURL,
		"token":    e.Token,
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
