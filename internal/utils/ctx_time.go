package utils

import (
	"context"
	"time"
)

var (
	ctxKeyTimeNow = contextKey("vpsg-ctx-key-time-now")
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
