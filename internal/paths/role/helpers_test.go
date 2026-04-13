package role_test

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	modelConfig "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/config"
	modelRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
	pathRole "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/role"
)

// mockRoleBackend is a hand-written mock satisfying the roleBackend interface.
type mockRoleBackend struct {
	role      *modelRole.Role
	roleErr   error
	config    *modelConfig.EntryConfig
	configErr error
	sendEvent func(ctx context.Context, eventType event.EventType, metadata map[string]string) error
}

func (m *mockRoleBackend) Logger() hclog.Logger { return hclog.NewNullLogger() }
func (m *mockRoleBackend) LockForKey(_, _ string) *locksutil.LockEntry {
	return locksutil.CreateLocks()[0]
}
func (m *mockRoleBackend) GetRole(_ context.Context, _ logical.Storage, _ string) (*modelRole.Role, error) {
	return m.role, m.roleErr
}
func (m *mockRoleBackend) GetConfig(_ context.Context, _ logical.Storage, _ string) (*modelConfig.EntryConfig, error) {
	return m.config, m.configErr
}
func (m *mockRoleBackend) SaveConfig(_ context.Context, _ logical.Storage, _ *modelConfig.EntryConfig) error {
	return nil
}
func (m *mockRoleBackend) SendEvent(ctx context.Context, eventType event.EventType, metadata map[string]string) error {
	if m.sendEvent != nil {
		return m.sendEvent(ctx, eventType, metadata)
	}
	return nil
}

// testConfig returns a minimal EntryConfig for test use.
func testConfig() *modelConfig.EntryConfig {
	return &modelConfig.EntryConfig{
		BaseURL: "https://gitlab.example.com",
		Token:   "glpat-test-token-value",
		Type:    gitlabTypes.TypeSelfManaged,
		Name:    "default",
	}
}

// newFieldData creates a FieldData using the exported role schema.
func newFieldData(raw map[string]interface{}) *framework.FieldData {
	return &framework.FieldData{Raw: raw, Schema: pathRole.FieldSchemaRoles}
}

// writeHandler returns the CreateOperation handler for the role CRUD path.
func writeHandler(mb *mockRoleBackend) framework.OperationFunc {
	return pathRole.New(mb).Paths()[1].Operations[logical.CreateOperation].Handler()
}

// readHandler returns the ReadOperation handler for the role CRUD path.
func readHandler(mb *mockRoleBackend) framework.OperationFunc {
	return pathRole.New(mb).Paths()[1].Operations[logical.ReadOperation].Handler()
}

// deleteHandler returns the DeleteOperation handler for the role CRUD path.
func deleteHandler(mb *mockRoleBackend) framework.OperationFunc {
	return pathRole.New(mb).Paths()[1].Operations[logical.DeleteOperation].Handler()
}

// listHandler returns the ListOperation handler for the role list path.
func listHandler(mb *mockRoleBackend) framework.OperationFunc {
	return pathRole.New(mb).Paths()[0].Operations[logical.ListOperation].Handler()
}

// newRequest creates a minimal logical.Request with in-memory storage.
func newRequest() *logical.Request {
	return &logical.Request{Storage: &logical.InmemStorage{}}
}
