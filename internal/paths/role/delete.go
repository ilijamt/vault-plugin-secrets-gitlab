package role

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
)

func (p *Provider) pathRolesDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var resp *logical.Response
	var err error
	var roleName = data.Get("role_name").(string)
	lock := p.b.RoleLockForKey(roleName)
	lock.Lock()
	defer lock.Unlock()

	_, err = p.b.GetRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, fmt.Errorf("error getting role: %w", err)
	}

	err = req.Storage.Delete(ctx, fmt.Sprintf("%s/%s", backend.PathRoleStorage, roleName))
	if err != nil {
		return nil, fmt.Errorf("error deleting role: %w", err)
	}

	_ = p.b.SendEvent(ctx, eventDelete, map[string]string{
		"path":      "roles",
		"role_name": roleName,
	})

	p.b.Logger().Debug("Role deleted", "role", roleName)

	return resp, nil
}
