package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"google.golang.org/protobuf/types/known/structpb"
)

func Event(ctx context.Context, b *framework.Backend, prefix, eventType string, metadata map[string]string) error {
	var err error
	var ev *logical.EventData
	if ev, err = logical.NewEvent(); err == nil {
		var metadataBytes []byte
		metadataBytes, _ = json.Marshal(metadata)
		ev.Metadata = &structpb.Struct{}
		_ = ev.Metadata.UnmarshalJSON(metadataBytes)
		err = b.SendEvent(ctx, logical.EventType(fmt.Sprintf("%s/%s", prefix, eventType)), ev)
	}
	return err
}
