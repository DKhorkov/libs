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
func (p *CommonProvider) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return p.client.Set(ctx, key, value, expiration).Err()
}

// SetNX sets key, if not already exists.
func (p *CommonProvider) SetNX(ctx context.Context, key string, value any, expiration time.Duration) error {
	return p.client.SetNX(ctx, key, value, expiration).Err()
}

// Get gets key.
func (p *CommonProvider) Get(ctx context.Context, key string) (string, error) {
	return p.client.Get(ctx, key).Result()
}

// GetEx gets key and expires it, if ttl is expired.
func (p *CommonProvider) GetEx(ctx context.Context, key string, expiration time.Duration) (string, error) {
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
	var cursor uint64
	var err error

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
