package gitlab

import (
	"context"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

type EntryConfig struct {
	TokenId                int           `json:"token_id" yaml:"token_id" mapstructure:"token_id"`
	BaseURL                string        `json:"base_url" structs:"base_url" mapstructure:"base_url"`
	Token                  string        `json:"token" structs:"token" mapstructure:"token"`
	AutoRotateToken        bool          `json:"auto_rotate_token" structs:"auto_rotate_token" mapstructure:"auto_rotate_token"`
	AutoRotateBefore       time.Duration `json:"auto_rotate_before" structs:"auto_rotate_before" mapstructure:"auto_rotate_before"`
	TokenExpiresAt         time.Time     `json:"token_expires_at" structs:"token_expires_at" mapstructure:"token_expires_at"`
	RevokeAutoRotatedToken bool          `json:"revoke_auto_rotated_token" structs:"revoke_auto_rotated_token" mapstructure:"revoke_auto_rotated_token"`
}

func (e EntryConfig) LogicalResponseData() map[string]any {
	var tokenExpiresAt = ""
	if !e.TokenExpiresAt.IsZero() {
		tokenExpiresAt = e.TokenExpiresAt.Format(time.RFC3339)
	}

	return map[string]any{
		"base_url":                  e.BaseURL,
		"auto_rotate_token":         e.AutoRotateToken,
		"auto_rotate_before":        e.AutoRotateBefore.String(),
		"token_id":                  e.TokenId,
		"token_expires_at":          tokenExpiresAt,
		"token_sha1_hash":           fmt.Sprintf("%x", sha1.Sum([]byte(e.Token))),
		"revoke_auto_rotated_token": e.RevokeAutoRotatedToken,
	}
}

func getConfig(ctx context.Context, s logical.Storage) (*EntryConfig, error) {
	entry, err := s.Get(ctx, PathConfigStorage)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	cfg := new(EntryConfig)
	if err := entry.DecodeJSON(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func saveConfig(ctx context.Context, config EntryConfig, s logical.Storage) error {
	var err error
	var storageEntry *logical.StorageEntry
	storageEntry, err = logical.StorageEntryJSON(PathConfigStorage, config)
	if err != nil {
		return nil
	}

	return s.Put(ctx, storageEntry)
}
