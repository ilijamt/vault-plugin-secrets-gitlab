package gitlab

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
)

func getRole(ctx context.Context, name string, s logical.Storage) (r *role.Role, err error) {
	var entry *logical.StorageEntry
	if entry, err = s.Get(ctx, fmt.Sprintf("%s/%s", PathRoleStorage, name)); err == nil {
		if entry == nil {
			return nil, nil
		}
		r = new(role.Role)
		_ = entry.DecodeJSON(r)
	}
	return r, err
}
