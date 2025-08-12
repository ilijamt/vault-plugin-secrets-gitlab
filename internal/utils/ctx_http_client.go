package utils

import (
	"context"
	"net/http"
)

var (
	ctxKeyHttpClient = contextKey("vpsg-ctx-key-http-client")
)

func HttpClientNewContext(ctx context.Context, httpClient *http.Client) context.Context {
	return context.WithValue(ctx, ctxKeyHttpClient, httpClient)
}

func HttpClientFromContext(ctx context.Context) (*http.Client, bool) {
	u, ok := ctx.Value(ctxKeyHttpClient).(*http.Client)
	if !ok {
		u = nil
	}
	return u, ok
}
