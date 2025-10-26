package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

// Delete removes the storage entry at the specified fullPath.
func Delete(ctx context.Context, s logical.Storage, fullPath string) (err error) {
	if s == nil {
		return fmt.Errorf("%w: local.Storage", errs.ErrNilValue)
	}
	return s.Delete(ctx, fullPath)
}

// Save marshals and stores 'data' at the given rootPath in storage 's'.
func Save(ctx context.Context, s logical.Storage, rootPath string, data Named) (err error) {
	if s == nil {
		err = errors.Join(err, fmt.Errorf("local.Storage: %w", errs.ErrNilValue))
	}

	if data == nil {
		err = errors.Join(err, fmt.Errorf("%w: model", errs.ErrNilValue))
	}

	if err != nil {
		return err
	}

	var entry *logical.StorageEntry
	if entry, err = logical.StorageEntryJSON(fmt.Sprintf("%s/%s", rootPath, data.GetName()), data); err == nil {
		err = s.Put(ctx, entry)
	}
	return err
}

// Get retrieves and decodes an entry at fullPath into a new T.
// Returns (nil, nil) if the entry is not found.
func Get[T any](ctx context.Context, s logical.Storage, fullPath string) (data *T, err error) {
	if s == nil {
		return nil, fmt.Errorf("%w: local.Storage", errs.ErrNilValue)
	}
	var entry *logical.StorageEntry
	if entry, err = s.Get(ctx, fullPath); err == nil {
		if entry == nil {
			return nil, nil
		}
		data = new(T)
		err = entry.DecodeJSON(data)
	}

	return data, err
}

// List returns the list of keys at fullPath in storage.
func List(ctx context.Context, s logical.Storage, fullPath string) (entries []string, err error) {
	if s == nil {
		return entries, fmt.Errorf("%w: local.Storage", errs.ErrNilValue)
	}
	return s.List(ctx, fullPath)
}
