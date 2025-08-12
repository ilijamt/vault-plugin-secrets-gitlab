package gitlab

import (
	"context"
	"time"
)

type contextKey string

const (
	DefaultConfigFieldAccessTokenMaxTTL = 7 * 24 * time.Hour
	DefaultConfigFieldAccessTokenRotate = DefaultAutoRotateBeforeMinTTL
	DefaultRoleFieldAccessTokenMaxTTL   = 24 * time.Hour
	DefaultAccessTokenMinTTL            = 24 * time.Hour
	DefaultAccessTokenMaxPossibleTTL    = 365 * 24 * time.Hour
	DefaultAutoRotateBeforeMinTTL       = 24 * time.Hour
	DefaultAutoRotateBeforeMaxTTL       = 730 * time.Hour
	ctxKeyHttpClient                    = contextKey("vpsg-ctx-key-http-client")
	ctxKeyGitlabClient                  = contextKey("vpsg-ctx-key-gitlab-client")
	ctxKeyTimeNow                       = contextKey("vpsg-ctx-key-time-now")
	DefaultConfigName                   = "default"
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
