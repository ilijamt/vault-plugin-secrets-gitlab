package utils

import (
	"context"
	"time"
)

var (
	ctxKeyTimeNow = contextKey("vpsg-ctx-key-time-now")
)

// WithStaticTime returns a new context.Context that carries a specific static time.
//
// This function embeds a given time.Time value into the provided context, allowing
// parts of an application to operate with a fixed notion of the current time. This
// can be particularly useful in testing scenarios where you need to control or
// simulate time progression.
func WithStaticTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, ctxKeyTimeNow, t)
}

// TimeFromContext extracts a time.Time from the given context.
//
// This function retrieves a time value that was previously embedded in the context.
// If the context does not contain such a time value, it defaults to returning time.Now(),
// effectively providing the current system time.
func TimeFromContext(ctx context.Context) time.Time {
	t, ok := ctx.Value(ctxKeyTimeNow).(time.Time)
	if !ok {
		return time.Now()
	}
	return t
}
