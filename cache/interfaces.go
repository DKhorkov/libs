package cache

import (
	"context"
	"time"
)

// SortedSetProvider provides methods for sorted sets — the only structure here,
// whose members are kept ordered by a numeric score.
//
// Separate from Provider, because a consumer, which keeps a queue of due work,
// needs these four methods and none of the string ones.
type SortedSetProvider interface {
	// ZAdd adds members to sorted set of key. Repeated add updates score of an
	// already existing member instead of adding a second one.
	ZAdd(ctx context.Context, key string, members ...Z) error

	// ZRem removes members from sorted set of key and returns, how many of them
	// were actually removed.
	ZRem(ctx context.Context, key string, members ...string) (int64, error)

	// ZScore returns score of sorted set member. ErrNotFound, if there is no such
	// member or no such key.
	ZScore(ctx context.Context, key, member string) (float64, error)

	// ZRangeByScore returns members of sorted set, whose scores fall into provided
	// range, ordered by score.
	ZRangeByScore(ctx context.Context, key string, by ZRangeBy) ([]string, error)
}

// Provider provides methods for setting cache and getting cached data.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/provider.go -package=mocks -exclude_interfaces=
type Provider interface {
	SortedSetProvider

	// Set sets key.
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// SetNX sets key, if not already exists.
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) error

	// TrySet sets key, if not already exists, and reports whether it was set.
	// Unlike SetNX it can be used as a lock: out of many concurrent attempts
	// exactly one gets true.
	TrySet(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)

	// Get gets key.
	Get(ctx context.Context, key string) (string, error)

	// GetEx gets key and expires it, if ttl is expired.
	GetEx(ctx context.Context, key string, expiration time.Duration) (string, error)

	// GetDel gets key and deletes it.
	GetDel(ctx context.Context, key string) (string, error)

	// MGet gets values of all provided keys at once. Result follows the order of
	// keys, and nil at a position means, that this key is missing.
	MGet(ctx context.Context, keys ...string) ([]*string, error)

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

	// DelByPattern deletes all keys, which matches provided pattern.
	DelByPattern(ctx context.Context, pattern string, batchSize *int64) error

	// ExpireAt sets the moment, at which key expires. Needed by keys, created by
	// commands, which do not take expiration themselves, such as Incr.
	ExpireAt(ctx context.Context, key string, at time.Time) error

	// TTL returns remaining time to live of key: -1 for key without expiration,
	// -2 for missing key.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Ping checks status.
	Ping(ctx context.Context) (string, error)

	Close() error
}
