package cache

import (
	"context"
	"time"
)

// Provider provides methods for setting cache and getting cached data.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/provider.go -package=mocks -exclude_interfaces=
type Provider interface {
	// Set sets key.
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// SetNX sets key, if not already exists.
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) error

	// Get gets key.
	Get(ctx context.Context, key string) (string, error)

	// GetEx gets key and expires it, if ttl is expired.
	GetEx(ctx context.Context, key string, expiration time.Duration) (string, error)

	// GetDel gets key and deletes it.
	GetDel(ctx context.Context, key string) (string, error)

	// Incr increments key.
	Incr(ctx context.Context, key string) (int64, error)

	// IncrBy increments key by value (numeric such as +1, +2 and so on).
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	// Decr decrements key.
	Decr(ctx context.Context, key string) (int64, error)

	// DecrBy decrements key by value (numeric such as -1, -2 and so on).
	DecrBy(ctx context.Context, key string, decrement int64) (int64, error)

	// Del deletes key.
	Del(ctx context.Context, keys ...string) error

	// Ping checks status.
	Ping(ctx context.Context) (string, error)

	Close() error
}
