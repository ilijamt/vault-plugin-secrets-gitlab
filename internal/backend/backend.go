package backend

import (
	"context"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab"
)

// ClientProvider abstracts obtaining a gitlab client by config name.
type ClientProvider interface {
	GetClientByName(ctx context.Context, s logical.Storage, name string) (gitlab.Client, error)
}

// EventSender abstracts sending audit/events from the backend.
type EventSender interface {
	SendTokenEvent(ctx context.Context, eventType string, metadata map[string]string) error
}
