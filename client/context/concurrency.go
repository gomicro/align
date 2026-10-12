package context

import (
	"context"
	"fmt"
)

type concurrencyContextKey struct{}

// WithConcurrency stores an invocation-scoped limit for operations that support it.
func WithConcurrency(ctx context.Context, limit int) context.Context {
	return context.WithValue(ctx, concurrencyContextKey{}, limit)
}

// Concurrency returns the override or fallback, rejecting nonpositive limits.
func Concurrency(ctx context.Context, fallback int) (int, error) {
	limit, ok := ctx.Value(concurrencyContextKey{}).(int)
	if !ok {
		limit = fallback
	}

	if limit < 1 {
		return 0, fmt.Errorf("concurrency must be a positive integer")
	}

	return limit, nil
}
