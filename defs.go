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
	DefaultConfigName                   = "default"
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

func GitlabClientNewContext(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, ctxKeyGitlabClient, client)
}

func GitlabClientFromContext(ctx context.Context) (Client, bool) {
	u, ok := ctx.Value(ctxKeyGitlabClient).(Client)
	if !ok {
		u = nil
	}
	return u, ok
}
