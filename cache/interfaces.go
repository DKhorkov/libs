package cache

import (
	"context"
	"time"
)

// Provider provides methods for setting cache and getting cached data.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/provider.go -package=mocks -exclude_interfaces=
type Provider interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error   // sets key
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) error // sets key, if not already exists
	Get(ctx context.Context, key string) (string, error)                              // gets key
	GetEx(ctx context.Context, key string, expiration time.Duration) (string, error)  // gets key and expires it, if ttl is expired
	GetDel(ctx context.Context, key string) (string, error)                           // gets key and deletes it
	Incr(ctx context.Context, key string) (int64, error)                              // increments key
	IncrBy(ctx context.Context, key string, value int64) (int64, error)               // increments key by value (numeric such as +1, +2 and so on)
	Decr(ctx context.Context, key string) (int64, error)                              // decrements key
	DecrBy(ctx context.Context, key string, decrement int64) (int64, error)           // decrements key by value (numeric such as -1, -2 and so on)
	Del(ctx context.Context, keys ...string) error                                    // deletes key
	Ping(ctx context.Context) (string, error)                                         // checks status
	Close() error
}
