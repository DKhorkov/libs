package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultBatchSize int64 = 500
)

type CommonProvider struct {
	client *redis.Client
}

func New(opts ...Option) (*CommonProvider, error) {
	cacheOptions := newOptions()
	for _, opt := range opts {
		err := opt(cacheOptions)
		if err != nil {
			return nil, err
		}
	}

	clientOptions := &redis.Options{
		Addr:                  fmt.Sprintf("%s:%d", cacheOptions.host, cacheOptions.port),
		ClientName:            cacheOptions.clientName,
		Username:              cacheOptions.username,
		Password:              cacheOptions.password,
		DB:                    cacheOptions.db,
		MaxRetries:            cacheOptions.maxRetries,
		MinRetryBackoff:       cacheOptions.minRetryBackoff,
		MaxRetryBackoff:       cacheOptions.maxRetryBackoff,
		DialTimeout:           cacheOptions.dialTimeout,
		ReadTimeout:           cacheOptions.readTimeout,
		WriteTimeout:          cacheOptions.writeTimeout,
		ContextTimeoutEnabled: cacheOptions.contextTimeoutEnabled,
		PoolFIFO:              cacheOptions.poolFIFO,
		PoolSize:              cacheOptions.poolSize,
		PoolTimeout:           cacheOptions.poolTimeout,
		MinIdleConns:          cacheOptions.minIdleConnections,
		MaxIdleConns:          cacheOptions.maxIdleConnections,
		MaxActiveConns:        cacheOptions.maxActiveConnections,
		ConnMaxIdleTime:       cacheOptions.connectionMaxIdleTime,
		ConnMaxLifetime:       cacheOptions.connectionMaxLifetime,
	}

	client := redis.NewClient(clientOptions)

	provider := &CommonProvider{client: client}
	if _, err := provider.Ping(context.Background()); err != nil {
		return nil, err
	}

	return provider, nil
}

// Set sets key.
func (p *CommonProvider) Set(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) error {
	return p.client.Set(ctx, key, value, expiration).Err()
}

// SetNX sets key, if not already exists.
func (p *CommonProvider) SetNX(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) error {
	return p.client.SetNX(ctx, key, value, expiration).Err()
}

// Get gets key.
func (p *CommonProvider) Get(ctx context.Context, key string) (string, error) {
	return p.client.Get(ctx, key).Result()
}

// GetEx gets key and expires it, if ttl is expired.
func (p *CommonProvider) GetEx(
	ctx context.Context,
	key string,
	expiration time.Duration,
) (string, error) {
	return p.client.GetEx(ctx, key, expiration).Result()
}

// GetDel gets key and deletes it.
func (p *CommonProvider) GetDel(ctx context.Context, key string) (string, error) {
	return p.client.GetDel(ctx, key).Result()
}

// Incr increments key.
func (p *CommonProvider) Incr(ctx context.Context, key string) (int64, error) {
	return p.client.Incr(ctx, key).Result()
}

// IncrBy increments key by value (numeric such as +1, +2 and so on).
func (p *CommonProvider) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return p.client.IncrBy(ctx, key, value).Result()
}

// Decr decrements key.
func (p *CommonProvider) Decr(ctx context.Context, key string) (int64, error) {
	return p.client.Decr(ctx, key).Result()
}

// DecrBy decrements key by value (numeric such as -1, -2 and so on).
func (p *CommonProvider) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return p.client.DecrBy(ctx, key, decrement).Result()
}

// Del deletes keys.
func (p *CommonProvider) Del(ctx context.Context, keys ...string) error {
	return p.client.Del(ctx, keys...).Err()
}

// DelByPattern deletes all keys, which matches provided pattern.
func (p *CommonProvider) DelByPattern(ctx context.Context, pattern string, batchSize *int64) error {
	var (
		cursor uint64
		err    error
	)

	bs := defaultBatchSize
	if batchSize != nil {
		bs = *batchSize
	}

	for {
		var keys []string

		keys, cursor, err = p.client.Scan(ctx, cursor, pattern, bs).Result()
		if err != nil {
			return fmt.Errorf("error scanning keys: %w", err)
		}

		if len(keys) > 0 {
			if err = p.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("error deleting keys: %w", err)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

// Ping checks status.
func (p *CommonProvider) Ping(ctx context.Context) (string, error) {
	return p.client.Ping(ctx).Result()
}

// Close closes connection to cache.
func (p *CommonProvider) Close() error {
	return p.client.Close()
}

// TrySet sets key, if it does not already exist, and reports whether it was set.
//
// Unlike SetNX, which hides the outcome behind error alone, it can be used as a
// lock: out of many concurrent attempts to take the same key exactly one gets
// true, and the rest get false.
func (p *CommonProvider) TrySet(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) (bool, error) {
	return p.client.SetNX(ctx, key, value, expiration).Result()
}

// MGet gets values of all provided keys at once.
//
// Result follows the order of keys, and nil at a position means, that this key
// is missing. One round trip instead of a Get per key: values, read separately,
// belong to different moments in time, and decisions made on them may be based
// on a state, which never existed.
func (p *CommonProvider) MGet(ctx context.Context, keys ...string) ([]*string, error) {
	if len(keys) == 0 {
		return []*string{}, nil
	}

	rawValues, err := p.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	values := make([]*string, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(string)
		if !ok {
			values = append(values, nil)

			continue
		}

		values = append(values, &value)
	}

	return values, nil
}

// ExpireAt sets the moment, at which key expires.
//
// Needed by keys, created by commands, which do not take expiration themselves:
// Incr creates a counter without any, and such counter stays in cache forever.
// Moment in the past deletes key. Missing key is not an error.
func (p *CommonProvider) ExpireAt(ctx context.Context, key string, at time.Time) error {
	return p.client.ExpireAt(ctx, key, at).Err()
}

// TTL returns remaining time to live of key.
//
// Negative results repeat the ones of Redis itself: -1 means key without
// expiration, -2 means missing key.
func (p *CommonProvider) TTL(ctx context.Context, key string) (time.Duration, error) {
	return p.client.TTL(ctx, key).Result()
}

// ZAdd adds members to sorted set of key, creating it, if needed.
//
// Member is addressed by its value, so repeated add updates score of an already
// existing member instead of adding a second one.
func (p *CommonProvider) ZAdd(ctx context.Context, key string, members ...Z) error {
	if len(members) == 0 {
		return nil
	}

	values := make([]redis.Z, 0, len(members))
	for _, member := range members {
		values = append(values, redis.Z{Score: member.Score, Member: member.Member})
	}

	return p.client.ZAdd(ctx, key, values...).Err()
}

// ZRem removes members from sorted set of key and returns, how many of them
// were actually removed.
//
// The number matters: it is the one, who got 1, that owns the member. Two
// readers, which saw the same entry, are told apart only this way.
func (p *CommonProvider) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
	if len(members) == 0 {
		return 0, nil
	}

	values := make([]any, 0, len(members))
	for _, member := range members {
		values = append(values, member)
	}

	return p.client.ZRem(ctx, key, values...).Result()
}

// ZScore returns score of sorted set member.
//
// ErrNotFound, if there is no such member or no such key at all.
func (p *CommonProvider) ZScore(ctx context.Context, key, member string) (float64, error) {
	return p.client.ZScore(ctx, key, member).Result()
}

// ZRangeByScore returns members of sorted set, whose scores fall into provided
// range, ordered by score.
//
// Count is what makes it usable for queues of due work: a range without limit
// returns everything, that piled up, while a batch is handled and repeated.
func (p *CommonProvider) ZRangeByScore(
	ctx context.Context,
	key string,
	by ZRangeBy,
) ([]string, error) {
	return p.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    by.Min,
		Max:    by.Max,
		Offset: by.Offset,
		Count:  by.Count,
	}).Result()
}
