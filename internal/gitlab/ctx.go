package gitlab

import "context"

type contextKey string

var (
	ctxKeyGitlabClient = contextKey("vpsg-ctx-key-gitlab-client")
)

func ClientNewContext(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, ctxKeyGitlabClient, client)
}

func ClientFromContext(ctx context.Context) (Client, bool) {
	u, ok := ctx.Value(ctxKeyGitlabClient).(Client)
	if !ok {
		u = nil
	}
	return u, ok
}
