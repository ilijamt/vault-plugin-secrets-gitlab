package utils

import (
	"context"
	"net/http"
)

var (
	ctxKeyHttpClient = contextKey("vpsg-ctx-key-http-client")
)

// HttpClientNewContext returns a new context.Context that carries the provided http.Client.
//
// This function embeds a given HTTP client into the provided context, allowing it
// to be passed through the application and retrieved later. This is useful for
// managing HTTP client configurations and dependency injection across different
// parts of an application that require HTTP clients.
func HttpClientNewContext(ctx context.Context, httpClient *http.Client) context.Context {
	return context.WithValue(ctx, ctxKeyHttpClient, httpClient)
}

// HttpClientFromContext extracts the http.Client from a given context.
//
// This function retrieves an HTTP client that was previously embedded in the context.
// If the context does not contain an HTTP client or it cannot be asserted as an
// *http.Client, the function returns nil and false.
func HttpClientFromContext(ctx context.Context) (*http.Client, bool) {
	u, ok := ctx.Value(ctxKeyHttpClient).(*http.Client)
	if !ok {
		u = nil
	}
	return u, ok
}
