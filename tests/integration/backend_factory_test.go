//go:build paths || saas || selfhosted || e2e

package integration_test

import (
	"cmp"
	"context"
	"fmt"
	"io"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

func getBackendWithEvents(ctx context.Context) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	return getBackendWithFlagsWithEvents(ctx, flags.Flags{})
}

func getBackendWithFlagsWithEvents(ctx context.Context, flags flags.Flags) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	events := &mockEventsSender{}
	config := &logical.BackendConfig{
		Logger:       logging.NewVaultLoggerWithWriter(io.Discard, log.NoLevel),
		System:       &logical.StaticSystemView{},
		StorageView:  &logical.InmemStorage{},
		BackendUUID:  "test",
		EventsSender: events,
	}

	b, err := gitlab.Factory(flags)(ctx, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to create Backend: %w", err)
	}

	return b.(*gitlab.Backend), config.StorageView, events, nil
}

func writeBackendConfigWithName(ctx context.Context, b *gitlab.Backend, l logical.Storage, config map[string]any, name string) error {
	var _, err = b.HandleRequest(ctx, &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/%s", backend.PathConfigStorage, cmp.Or(name, backend.DefaultConfigName)), Storage: l,
		Data: config,
	})
	return err
}

func writeBackendConfig(ctx context.Context, b *gitlab.Backend, l logical.Storage, config map[string]any) error {
	return writeBackendConfigWithName(ctx, b, l, config, backend.DefaultConfigName)
}

func getBackendWithEventsAndConfig(ctx context.Context, config map[string]any) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	var b, storage, events, _ = getBackendWithEvents(ctx)
	return b, storage, events, writeBackendConfig(ctx, b, storage, config)
}

func getBackendWithEventsAndConfigName(ctx context.Context, config map[string]any, name string) (*gitlab.Backend, logical.Storage, *mockEventsSender, error) {
	var b, storage, events, _ = getBackendWithEvents(ctx)
	return b, storage, events, writeBackendConfigWithName(ctx, b, storage, config, name)
}

func getBackendWithConfig(ctx context.Context, config map[string]any) (*gitlab.Backend, logical.Storage, error) {
	var b, storage, _, _ = getBackendWithEvents(ctx)
	return b, storage, writeBackendConfig(ctx, b, storage, config)
}

func getBackend(ctx context.Context) (*gitlab.Backend, logical.Storage, error) {
	b, storage, _, err := getBackendWithEvents(ctx)
	return b, storage, err
}
