package gitlab

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
)

func getConfig(ctx context.Context, s logical.Storage, name string) (cfg *config.EntryConfig, err error) {
	if s == nil {
		return nil, fmt.Errorf("%w: local.Storage", errs.ErrNilValue)
	}
	var entry *logical.StorageEntry
	if entry, err = s.Get(ctx, fmt.Sprintf("%s/%s", PathConfigStorage, name)); err == nil {
		if entry == nil {
			return nil, nil
		}
		cfg = new(config.EntryConfig)
		_ = entry.DecodeJSON(cfg)
	}
	return cfg, err
}

func saveConfig(ctx context.Context, config config.EntryConfig, s logical.Storage) (err error) {
	var storageEntry *logical.StorageEntry
	if storageEntry, err = logical.StorageEntryJSON(fmt.Sprintf("%s/%s", PathConfigStorage, config.Name), config); err == nil {
		err = s.Put(ctx, storageEntry)
	}
	return err
}
