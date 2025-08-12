package gitlab

import "context"

type contextKey string

var (
	ctxKeyGitlabClient = contextKey("vpsg-ctx-key-gitlab-client")
)

// ClientNewContext returns a new context.Context that carries the provided Client.
//
// This function embeds the specified GitLab client into the given context, allowing
// it to be retrieved later in the execution flow. It's particularly useful for passing
// around client information across different layers of an application.
func ClientNewContext(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, ctxKeyGitlabClient, client)
}

// ClientFromContext extracts the GitLab Client from the provided context.
//
// This function attempts to retrieve a Client from the given context. If it was
// not present or if it cannot be asserted as a Client, the returned Client will
// be nil and the boolean will be false.
func ClientFromContext(ctx context.Context) (Client, bool) {
	u, ok := ctx.Value(ctxKeyGitlabClient).(Client)
	if !ok {
		u = nil
	}
	return u, ok
}
