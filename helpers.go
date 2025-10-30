package gitlab

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
)

func getConfig(ctx context.Context, s logical.Storage, name string) (cfg *config.EntryConfig, err error) {
	return model.Get[config.EntryConfig](ctx, s, fmt.Sprintf("%s/%s", PathConfigStorage, name))
}

func saveConfig(ctx context.Context, config *config.EntryConfig, s logical.Storage) error {
	return model.Save(ctx, s, PathConfigStorage, config)
}

func getRole(ctx context.Context, name string, s logical.Storage) (r *role.Role, err error) {
	return model.Get[role.Role](ctx, s, fmt.Sprintf("%s/%s", PathRoleStorage, name))
}
