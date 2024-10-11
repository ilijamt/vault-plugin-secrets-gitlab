package gitlab

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"google.golang.org/protobuf/types/known/structpb"
)

func event(ctx context.Context, b *framework.Backend, eventType string, metadata map[string]string) {
	var err error
	var ev *logical.EventData
	if ev, err = logical.NewEvent(); err == nil {
		var metadataBytes []byte
		metadataBytes, _ = json.Marshal(metadata)
		ev.Metadata = &structpb.Struct{}
		_ = ev.Metadata.UnmarshalJSON(metadataBytes)
		_ = b.SendEvent(ctx, logical.EventType(fmt.Sprintf("%s/%s", operationPrefixGitlabAccessTokens, eventType)), ev)
	}
}
