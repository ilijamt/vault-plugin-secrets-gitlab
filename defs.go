package gitlab

import (
	"context"
	"errors"
	"net/http"
	"time"
)

var (
	ErrNilValue             = errors.New("nil value")
	ErrInvalidValue         = errors.New("invalid value")
	ErrFieldRequired        = errors.New("required field")
	ErrFieldInvalidValue    = errors.New("invalid value for field")
	ErrBackendNotConfigured = errors.New("backend not configured")
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

func WithStaticTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, ctxKeyTimeNow, t)
}

func TimeFromContext(ctx context.Context) time.Time {
	t, ok := ctx.Value(ctxKeyTimeNow).(time.Time)
	if !ok {
		return time.Now()
	}
	return t
}

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
