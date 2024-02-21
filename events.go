package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"google.golang.org/protobuf/types/known/structpb"
)

func event(ctx context.Context, b *framework.Backend, eventType string, metadata map[string]string) {
	ev, err := logical.NewEvent()
	if err != nil {
		b.Logger().Warn("Error creating event", "error", err)
		return
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		b.Logger().Warn("Error marshaling metadata", "error", err)
		return
	}
	ev.Metadata = &structpb.Struct{}
	if err := ev.Metadata.UnmarshalJSON(metadataBytes); err != nil {
		b.Logger().Warn("Error unmarshalling metadata into proto", "error", err)
		return
	}
	err = b.SendEvent(ctx, logical.EventType(fmt.Sprintf("%s/%s", operationPrefixGitlabAccessTokens, eventType)), ev)
	// ignore events are disabled error
	if errors.Is(err, framework.ErrNoEvents) {
		return
	} else if err != nil {
		b.Logger().Warn("Error sending event", "error", err)
		return
	}
}
