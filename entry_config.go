package gitlab

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

type EntryConfig struct {
	TokenId          int           `json:"token_id" yaml:"token_id" mapstructure:"token_id"`
	BaseURL          string        `json:"base_url" structs:"base_url" mapstructure:"base_url"`
	Token            string        `json:"token" structs:"token" mapstructure:"token"`
	AutoRotateToken  bool          `json:"auto_rotate_token" structs:"auto_rotate_token" mapstructure:"auto_rotate_token"`
	AutoRotateBefore time.Duration `json:"auto_rotate_before" structs:"auto_rotate_before" mapstructure:"auto_rotate_before"`
	TokenCreatedAt   time.Time     `json:"token_created_at" structs:"token_created_at" mapstructure:"token_created_at"`
	TokenExpiresAt   time.Time     `json:"token_expires_at" structs:"token_expires_at" mapstructure:"token_expires_at"`
	Scopes           []string      `json:"scopes" structs:"scopes" mapstructure:"scopes"`
}

func (e EntryConfig) Response() *logical.Response {
	return &logical.Response{
		Secret: &logical.Secret{
			LeaseOptions: logical.LeaseOptions{},
			InternalData: map[string]any{
				"token_id": e.TokenId,
				"token":    e.Token,
			},
		},
		Data: e.LogicalResponseData(),
	}
}

func (e EntryConfig) LogicalResponseData() map[string]any {
	var tokenExpiresAt, tokenCreatedAt = "", ""
	if !e.TokenExpiresAt.IsZero() {
		tokenExpiresAt = e.TokenExpiresAt.Format(time.RFC3339)
	}
	if !e.TokenCreatedAt.IsZero() {
		tokenCreatedAt = e.TokenCreatedAt.Format(time.RFC3339)
	}

	return map[string]any{
		"base_url":           e.BaseURL,
		"auto_rotate_token":  e.AutoRotateToken,
		"auto_rotate_before": e.AutoRotateBefore.String(),
		"token_id":           e.TokenId,
		"token_created_at":   tokenCreatedAt,
		"token_expires_at":   tokenExpiresAt,
		"token_sha1_hash":    fmt.Sprintf("%x", sha1.Sum([]byte(e.Token))),
		"scopes":             strings.Join(e.Scopes, ", "),
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
