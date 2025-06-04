//go:build integration

package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/cache"
)

const (
	password = "hmtm_sso"
	port     = 8072
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		opts        []cache.Option
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "default options",
			opts:    []cache.Option{},
			wantErr: false,
		},
		{
			name: "all options",
			opts: []cache.Option{
				cache.WithHost("localhost"),
				cache.WithPort(port),
				cache.WithClientName("test-client"),
				cache.WithUsername(""),
				cache.WithPassword(password),
				cache.WithDB(0),
				cache.WithMaxRetries(3),
				cache.WithMinRetryBackoff(8 * time.Millisecond),
				cache.WithMaxRetryBackoff(512 * time.Millisecond),
				cache.WithDialTimeout(5 * time.Second),
				cache.WithReadTimeout(3 * time.Second),
				cache.WithWriteTimeout(3 * time.Second),
				cache.WithContextTimeoutEnabled(true),
				cache.WithPoolFIFO(false),
				cache.WithPoolSize(10),
				cache.WithPoolTimeout(4 * time.Second),
				cache.WithMinIdleConnections(0),
				cache.WithMaxIdleConnections(0),
				cache.WithMaxActiveConnections(0),
				cache.WithConnectionMaxIdleTime(30 * time.Minute),
				cache.WithConnectionMaxLifetime(0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := cache.New(tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, provider)
			assert.NoError(t, provider.Close())
		})
	}
}

func TestCommonProvider_CRUD(t *testing.T) {
	ctx := context.Background()
	provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
	require.NoError(t, err)

	defer func(provider *cache.CommonProvider) {
		err = provider.Close()
		if err != nil {

		}
	}(provider)

	tests := []struct {
		name        string
		action      func() error
		key         string
		value       interface{}
		expiration  time.Duration
		want        string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "set and get string",
			action: func() error {
				return provider.Set(ctx, "key1", "value1", time.Minute)
			},
			key:     "key1",
			want:    "value1",
			wantErr: false,
		},
		{
			name: "set and get empty string",
			action: func() error {
				return provider.Set(ctx, "key2", "", time.Minute)
			},
			key:     "key2",
			want:    "",
			wantErr: false,
		},
		{
			name: "setnx new key",
			action: func() error {
				return provider.SetNX(ctx, "key3", "value3", time.Minute)
			},
			key:     "key3",
			want:    "value3",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			_ = provider.Del(ctx, tt.key)

			// Execute action
			err = tt.action()
			if tt.wantErr && tt.action != nil {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else if tt.action != nil {
				require.NoError(t, err)
			}

			// Verify result
			if tt.key != "" {
				got, err := provider.Get(ctx, tt.key)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.want, got)
				}
			}
		})
	}
}

func TestCommonProvider_IncrDecr(t *testing.T) {
	ctx := context.Background()
	provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
	require.NoError(t, err)

	defer func(provider *cache.CommonProvider) {
		err = provider.Close()
		if err != nil {

		}
	}(provider)

	tests := []struct {
		name        string
		key         string
		action      func(context.Context, string, ...int64) (int64, error)
		value       int64
		want        int64
		wantErr     bool
		expectedErr error
	}{
		{
			name: "incr",
			key:  "counter1",
			action: func(ctx context.Context, key string, _ ...int64) (int64, error) {
				return provider.Incr(ctx, key)
			},
			want: 1,
		},
		{
			name: "incrby 5",
			key:  "counter2",
			action: func(ctx context.Context, key string, _ ...int64) (int64, error) {
				return provider.IncrBy(ctx, key, 5)
			},
			want: 5,
		},
		{
			name: "decr",
			key:  "counter3",
			action: func(ctx context.Context, key string, _ ...int64) (int64, error) {
				_, _ = provider.IncrBy(ctx, key, 10)
				return provider.Decr(ctx, key)
			},
			want: 9,
		},
		{
			name: "decrby 3",
			key:  "counter4",
			action: func(ctx context.Context, key string, _ ...int64) (int64, error) {
				_, _ = provider.IncrBy(ctx, key, 10)
				return provider.DecrBy(ctx, key, 3)
			},
			want: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			_ = provider.Del(ctx, tt.key)

			// Execute
			got, err := tt.action(ctx, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommonProvider_Ping(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func() *cache.CommonProvider
		want        string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful ping",
			setup: func() *cache.CommonProvider {
				provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
				require.NoError(t, err)

				return provider
			},
			want:    "PONG",
			wantErr: false,
		},
		{
			name: "ping closed connection",
			setup: func() *cache.CommonProvider {
				provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
				require.NoError(t, err)
				require.NoError(t, provider.Close())

				return provider
			},
			want:        "",
			wantErr:     true,
			expectedErr: redis.ErrClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setup()
			defer func(provider *cache.CommonProvider) {
				err := provider.Close()
				if err != nil {

				}
			}(provider)

			got, err := provider.Ping(ctx)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommonProvider_GetEx(t *testing.T) {
	ctx := context.Background()
	provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
	require.NoError(t, err)

	defer func(provider *cache.CommonProvider) {
		err = provider.Close()
		if err != nil {

		}
	}(provider)

	tests := []struct {
		name        string
		key         string
		value       string
		setTTL      time.Duration
		getExTTL    time.Duration
		waitTime    time.Duration
		wantValue   string
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "successful getex with new expiration",
			key:       "getex-key1",
			value:     "value1",
			setTTL:    time.Minute,
			getExTTL:  time.Second * 2,
			wantValue: "value1",
			wantErr:   false,
		},
		{
			name:      "getex extends key lifetime",
			key:       "getex-key2",
			value:     "value2",
			setTTL:    time.Second,
			getExTTL:  time.Second * 3,
			waitTime:  time.Second * 2,
			wantValue: "value2",
			wantErr:   false,
		},
		{
			name:        "getex non-existent key",
			key:         "getex-non-existent",
			getExTTL:    time.Second,
			wantValue:   "",
			wantErr:     true,
			expectedErr: redis.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			_ = provider.Del(ctx, tt.key)

			// Setup initial key if value provided
			if tt.value != "" {
				err = provider.Set(ctx, tt.key, tt.value, tt.setTTL)
				require.NoError(t, err)
			}

			// Test GetEx
			got, err := provider.GetEx(ctx, tt.key, tt.getExTTL)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, got)

			// Verify TTL if wait time specified
			if tt.waitTime > 0 {
				time.Sleep(tt.waitTime)
				_, err = provider.Get(ctx, tt.key)
				require.NoError(t, err, "key should still exist after wait time")
			}
		})
	}
}

func TestCommonProvider_GetDel(t *testing.T) {
	ctx := context.Background()
	provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
	require.NoError(t, err)

	defer func(provider *cache.CommonProvider) {
		err = provider.Close()
		if err != nil {

		}
	}(provider)

	tests := []struct {
		name        string
		key         string
		value       string
		ttl         time.Duration
		setup       func() // Additional setup if needed
		wantValue   string
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "successful getdel",
			key:       "getdel-key1",
			value:     "value1",
			ttl:       time.Minute,
			wantValue: "value1",
			wantErr:   false,
		},
		{
			name:        "getdel non-existent key",
			key:         "getdel-non-existent",
			wantValue:   "",
			wantErr:     true,
			expectedErr: redis.Nil,
		},
		{
			name: "getdel expired key",
			key:  "getdel-expired",
			setup: func() {
				_ = provider.Set(ctx, "getdel-expired", "value", time.Millisecond*10)
				time.Sleep(time.Millisecond * 20)
			},
			wantValue:   "",
			wantErr:     true,
			expectedErr: redis.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			_ = provider.Del(ctx, tt.key)

			// Additional setup if specified
			if tt.setup != nil {
				tt.setup()
			} else if tt.value != "" {
				// Default setup - set key with value
				err = provider.Set(ctx, tt.key, tt.value, tt.ttl)
				require.NoError(t, err)
			}

			// Test GetDel
			got, err := provider.GetDel(ctx, tt.key)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, got)

			// Verify key deleted
			_, err = provider.Get(ctx, tt.key)
			require.Error(t, err)
			assert.Equal(t, redis.Nil, err)
		})
	}
}
