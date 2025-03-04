package contextlib

import (
	"context"
)

// ContextKey - context-keys-type: should not use basic type string as key in context.WithValue (revive)
// https://vishnubharathi.codes/blog/context-with-value-pitfall/
type contextKey string

func ValueFromContext[T any](ctx context.Context, key string) (T, error) {
	value, ok := ctx.Value(contextKey(key)).(T)
	if !ok {
		return value, &ValueNotFoundError{Message: key}
	}

	return value, nil
}

func WithValue(ctx context.Context, key string, value any) context.Context {
	return context.WithValue(ctx, contextKey(key), value)
}
