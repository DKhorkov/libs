//go:build integration

package cache_test

import (
	"context"
	"fmt"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/pointers"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/cache"
)

// Connection params of Redis, which integration tests are run against.
//
// Defaults point to hmtm's local Redis, but both are overridable via env: the
// same tests are run against any project's local infrastructure without editing
// this file.
var (
	password = loadenv.GetEnv("CACHE_PASSWORD", "hmtm_sso")
	port     = loadenv.GetEnvAsInt("CACHE_PORT", 8072)
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
			t.Fatal(err)
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
			t.Fatal(err)
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
			t.Fatal(err)
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
			t.Fatal(err)
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

func TestCommonProvider_DelByPattern(t *testing.T) {
	ctx := context.Background()
	provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
	require.NoError(t, err)

	defer func(provider *cache.CommonProvider) {
		err = provider.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(provider)

	// Вспомогательная функция для создания тестовых ключей
	setupKeys := func(prefix string, count int) []string {
		keys := make([]string, 0, count)
		for i := 0; i < count; i++ {
			key := fmt.Sprintf("%s:%d", prefix, i)
			err = provider.Set(ctx, key, fmt.Sprintf("value%d", i), time.Second*15)
			require.NoError(t, err)
			keys = append(keys, key)
		}
		return keys
	}

	// Вспомогательная функция для проверки существования ключей
	assertKeysExist := func(t *testing.T, keys []string, shouldExist bool) {
		for _, key := range keys {
			_, err = provider.Get(ctx, key)
			if shouldExist {
				assert.NoError(t, err, "key %s should exist", key)
			} else {
				assert.Equal(t, redis.Nil, err, "key %s should not exist", key)
			}
		}
	}

	tests := []struct {
		name          string
		pattern       string
		setupKeys     []string // ключи, которые должны быть созданы
		extraKeys     []string // дополнительные ключи, которые не должны удаляться
		batchSize     *int64
		wantErr       bool
		expectedErr   error
		checkLeftKeys bool // проверять оставшиеся ключи
	}{
		{
			name:      "delete single key",
			pattern:   "testpattern:single*",
			setupKeys: []string{"testpattern:single-key"},
			batchSize: nil,
			wantErr:   false,
		},
		{
			name:      "delete multiple keys",
			pattern:   "testpattern:multi*",
			setupKeys: setupKeys("testpattern:multi", 5),
			batchSize: nil,
			wantErr:   false,
		},
		{
			name:          "delete with batch size",
			pattern:       "testpattern:batch*",
			setupKeys:     setupKeys("testpattern:batch", 10),
			batchSize:     pointers.New[int64](3),
			wantErr:       false,
			checkLeftKeys: true,
		},
		{
			name:      "no keys to delete",
			pattern:   "testpattern:nonexistent*",
			setupKeys: []string{},
			batchSize: nil,
			wantErr:   false,
		},
		{
			name:          "partial pattern match",
			pattern:       "testpattern:partial*",
			setupKeys:     []string{"testpattern:partial-match", "testpattern:partial-match2"},
			extraKeys:     []string{"otherprefix:partial-match"},
			batchSize:     nil,
			wantErr:       false,
			checkLeftKeys: true,
		},
		{
			name:    "complex pattern",
			pattern: "testpattern:complex:*_*",
			setupKeys: []string{
				"testpattern:complex:123_abc",
				"testpattern:complex:456_def",
			},
			batchSize:     nil,
			wantErr:       false,
			checkLeftKeys: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup - удаляем все тестовые ключи перед началом
			for _, key := range append(tt.setupKeys, tt.extraKeys...) {
				_ = provider.Del(ctx, key)
			}

			// Создаем тестовые ключи
			for _, key := range tt.setupKeys {
				err = provider.Set(ctx, key, "test-value", time.Second*15)
				require.NoError(t, err)
			}

			// Создаем дополнительные ключи, которые не должны удаляться
			for _, key := range tt.extraKeys {
				err = provider.Set(ctx, key, "test-value", time.Second*15)
				require.NoError(t, err)
			}

			// Проверяем, что ключи созданы
			assertKeysExist(t, tt.setupKeys, true)
			if len(tt.extraKeys) > 0 {
				assertKeysExist(t, tt.extraKeys, true)
			}

			// Выполняем тестируемый метод
			err = provider.DelByPattern(ctx, tt.pattern, tt.batchSize)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			require.NoError(t, err)

			// Проверяем, что целевые ключи удалены
			assertKeysExist(t, tt.setupKeys, false)

			// Проверяем, что дополнительные ключи остались
			if tt.checkLeftKeys && len(tt.extraKeys) > 0 {
				assertKeysExist(t, tt.extraKeys, true)
			}
		})
	}
}

// newProvider builds provider for a single test and closes it at the end.
func newProvider(t *testing.T) *cache.CommonProvider {
	t.Helper()

	provider, err := cache.New(cache.WithPassword(password), cache.WithPort(port))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, provider.Close())
	})

	return provider
}

// TestErrNotFound checks, that missing key is reported via package sentinel and
// not only via redis.Nil: consumers should not import redis driver just to tell
// "no such key" from a real failure.
func TestErrNotFound(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	require.NoError(t, provider.Del(ctx, "errnotfound-key"))

	_, err := provider.Get(ctx, "errnotfound-key")
	require.Error(t, err)
	assert.ErrorIs(t, err, cache.ErrNotFound)
	assert.ErrorIs(t, err, redis.Nil)
}

func TestCommonProvider_TrySet(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	tests := []struct {
		name        string
		key         string
		value       string
		expiration  time.Duration
		setup       func(t *testing.T, key string)
		want        bool
		wantValue   string
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "sets absent key",
			key:        "tryset-absent",
			value:      "value1",
			expiration: time.Minute,
			want:       true,
			wantValue:  "value1",
		},
		{
			name:       "does not overwrite existing key",
			key:        "tryset-existing",
			value:      "value2",
			expiration: time.Minute,
			setup: func(t *testing.T, key string) {
				t.Helper()
				require.NoError(t, provider.Set(ctx, key, "first", time.Minute))
			},
			want:      false,
			wantValue: "first",
		},
		{
			name:       "sets key, expired since previous attempt",
			key:        "tryset-expired",
			value:      "value3",
			expiration: time.Minute,
			setup: func(t *testing.T, key string) {
				t.Helper()
				require.NoError(t, provider.Set(ctx, key, "first", time.Millisecond*10))
				time.Sleep(time.Millisecond * 20)
			},
			want:      true,
			wantValue: "value3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, provider.Del(ctx, tt.key))

			if tt.setup != nil {
				tt.setup(t, tt.key)
			}

			got, err := provider.TrySet(ctx, tt.key, tt.value, tt.expiration)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			stored, err := provider.Get(ctx, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, stored)
		})
	}
}

// TestCommonProvider_TrySetIsExclusive checks the whole point of the method:
// out of many attempts to take the same key exactly one wins. Without it the
// method could not be used as a lock.
func TestCommonProvider_TrySetIsExclusive(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	const attempts = 10

	require.NoError(t, provider.Del(ctx, "tryset-exclusive"))

	won := 0

	for range attempts {
		ok, err := provider.TrySet(ctx, "tryset-exclusive", "owner", time.Minute)
		require.NoError(t, err)

		if ok {
			won++
		}
	}

	assert.Equal(t, 1, won)
}

func TestCommonProvider_MGet(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	require.NoError(t, provider.Del(ctx, "mget-1", "mget-2", "mget-3"))
	require.NoError(t, provider.Set(ctx, "mget-1", "value1", time.Minute))
	require.NoError(t, provider.Set(ctx, "mget-3", "value3", time.Minute))

	tests := []struct {
		name string
		keys []string
		want []*string
	}{
		{
			name: "all keys exist",
			keys: []string{"mget-1", "mget-3"},
			want: []*string{pointers.New("value1"), pointers.New("value3")},
		},
		{
			name: "missing key is nil at its position",
			keys: []string{"mget-1", "mget-2", "mget-3"},
			want: []*string{pointers.New("value1"), nil, pointers.New("value3")},
		},
		{
			name: "order follows requested keys",
			keys: []string{"mget-3", "mget-1"},
			want: []*string{pointers.New("value3"), pointers.New("value1")},
		},
		{
			name: "no keys requested",
			keys: []string{},
			want: []*string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.MGet(ctx, tt.keys...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommonProvider_TTL(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	require.NoError(t, provider.Del(ctx, "ttl-with", "ttl-without", "ttl-missing"))
	require.NoError(t, provider.Set(ctx, "ttl-with", "value", time.Minute))
	require.NoError(t, provider.Set(ctx, "ttl-without", "value", 0))

	tests := []struct {
		name   string
		key    string
		assert func(t *testing.T, ttl time.Duration)
	}{
		{
			name: "key with expiration",
			key:  "ttl-with",
			assert: func(t *testing.T, ttl time.Duration) {
				t.Helper()
				assert.Positive(t, ttl)
				assert.LessOrEqual(t, ttl, time.Minute)
			},
		},
		{
			name: "key without expiration",
			key:  "ttl-without",
			assert: func(t *testing.T, ttl time.Duration) {
				t.Helper()
				assert.Equal(t, time.Duration(-1), ttl)
			},
		},
		{
			name: "missing key",
			key:  "ttl-missing",
			assert: func(t *testing.T, ttl time.Duration) {
				t.Helper()
				assert.Equal(t, time.Duration(-2), ttl)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttl, err := provider.TTL(ctx, tt.key)
			require.NoError(t, err)
			tt.assert(t, ttl)
		})
	}
}

// TestCommonProvider_ExpireAt checks, that key, created without expiration, can
// be given one afterwards: counters are created by Incr, which never sets TTL,
// and without this method they would stay in cache forever.
func TestCommonProvider_ExpireAt(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	t.Run("gives expiration to key without one", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "expireat-counter"))

		_, err := provider.Incr(ctx, "expireat-counter")
		require.NoError(t, err)

		ttl, err := provider.TTL(ctx, "expireat-counter")
		require.NoError(t, err)
		assert.Equal(t, time.Duration(-1), ttl)

		require.NoError(t, provider.ExpireAt(ctx, "expireat-counter", time.Now().Add(time.Hour)))

		ttl, err = provider.TTL(ctx, "expireat-counter")
		require.NoError(t, err)
		assert.Positive(t, ttl)
		assert.LessOrEqual(t, ttl, time.Hour)
	})

	t.Run("moment in the past deletes key", func(t *testing.T) {
		require.NoError(t, provider.Set(ctx, "expireat-past", "value", time.Minute))
		require.NoError(t, provider.ExpireAt(ctx, "expireat-past", time.Now().Add(-time.Hour)))

		_, err := provider.Get(ctx, "expireat-past")
		require.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrNotFound)
	})

	t.Run("missing key is not an error", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "expireat-missing"))
		require.NoError(t, provider.ExpireAt(ctx, "expireat-missing", time.Now().Add(time.Hour)))
	})
}

func TestCommonProvider_ZAddAndZScore(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	t.Run("member round trips through its score", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-roundtrip"))
		require.NoError(t, provider.ZAdd(ctx, "zset-roundtrip", cache.Z{Score: 42, Member: "first"}))

		score, err := provider.ZScore(ctx, "zset-roundtrip", "first")
		require.NoError(t, err)
		assert.InDelta(t, 42.0, score, 0)
	})

	t.Run("repeated add updates score instead of doubling member", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-update"))
		require.NoError(t, provider.ZAdd(ctx, "zset-update", cache.Z{Score: 1, Member: "only"}))
		require.NoError(t, provider.ZAdd(ctx, "zset-update", cache.Z{Score: 2, Member: "only"}))

		score, err := provider.ZScore(ctx, "zset-update", "only")
		require.NoError(t, err)
		assert.InDelta(t, 2.0, score, 0)

		members, err := provider.ZRangeByScore(
			ctx,
			"zset-update",
			cache.ZRangeBy{Min: "-inf", Max: "+inf"},
		)
		require.NoError(t, err)
		assert.Equal(t, []string{"only"}, members)
	})

	t.Run("several members at once", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-batch"))
		require.NoError(
			t,
			provider.ZAdd(
				ctx,
				"zset-batch",
				cache.Z{Score: 2, Member: "second"},
				cache.Z{Score: 1, Member: "first"},
			),
		)

		members, err := provider.ZRangeByScore(
			ctx,
			"zset-batch",
			cache.ZRangeBy{Min: "-inf", Max: "+inf"},
		)
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "second"}, members)
	})

	t.Run("missing member", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-missing"))
		require.NoError(t, provider.ZAdd(ctx, "zset-missing", cache.Z{Score: 1, Member: "present"}))

		_, err := provider.ZScore(ctx, "zset-missing", "absent")
		require.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrNotFound)
	})

	t.Run("missing key", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-absent"))

		_, err := provider.ZScore(ctx, "zset-absent", "whatever")
		require.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrNotFound)
	})
}

func TestCommonProvider_ZRangeByScore(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	require.NoError(t, provider.Del(ctx, "zset-range"))
	require.NoError(
		t,
		provider.ZAdd(
			ctx,
			"zset-range",
			cache.Z{Score: 10, Member: "a"},
			cache.Z{Score: 20, Member: "b"},
			cache.Z{Score: 30, Member: "c"},
			cache.Z{Score: 40, Member: "d"},
		),
	)

	tests := []struct {
		name string
		by   cache.ZRangeBy
		want []string
	}{
		{
			name: "everything in score order",
			by:   cache.ZRangeBy{Min: "-inf", Max: "+inf"},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "everything, that is already due",
			by:   cache.ZRangeBy{Min: "-inf", Max: "25"},
			want: []string{"a", "b"},
		},
		{
			name: "bounds are inclusive",
			by:   cache.ZRangeBy{Min: "20", Max: "30"},
			want: []string{"b", "c"},
		},
		{
			name: "exclusive bound",
			by:   cache.ZRangeBy{Min: "(20", Max: "30"},
			want: []string{"c"},
		},
		{
			name: "count limits batch",
			by:   cache.ZRangeBy{Min: "-inf", Max: "+inf", Count: 2},
			want: []string{"a", "b"},
		},
		{
			name: "offset skips head of batch",
			by:   cache.ZRangeBy{Min: "-inf", Max: "+inf", Offset: 1, Count: 2},
			want: []string{"b", "c"},
		},
		{
			name: "nothing matches range",
			by:   cache.ZRangeBy{Min: "100", Max: "200"},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.ZRangeByScore(ctx, "zset-range", tt.by)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestCommonProvider_ZRem checks the number of removed members, and not only
// the fact of removal: it is the one who got 1 that owns the member, and a
// method reporting error alone could not be used to hand a queue entry to
// exactly one of two competing readers.
func TestCommonProvider_ZRem(t *testing.T) {
	ctx := context.Background()
	provider := newProvider(t)

	t.Run("removes member and reports it", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-rem"))
		require.NoError(t, provider.ZAdd(ctx, "zset-rem", cache.Z{Score: 1, Member: "taken"}))

		removed, err := provider.ZRem(ctx, "zset-rem", "taken")
		require.NoError(t, err)
		assert.Equal(t, int64(1), removed)

		_, err = provider.ZScore(ctx, "zset-rem", "taken")
		require.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrNotFound)
	})

	t.Run("second remove of the same member reports zero", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-rem-twice"))
		require.NoError(t, provider.ZAdd(ctx, "zset-rem-twice", cache.Z{Score: 1, Member: "taken"}))

		first, err := provider.ZRem(ctx, "zset-rem-twice", "taken")
		require.NoError(t, err)
		assert.Equal(t, int64(1), first)

		second, err := provider.ZRem(ctx, "zset-rem-twice", "taken")
		require.NoError(t, err)
		assert.Equal(t, int64(0), second)
	})

	t.Run("several members at once", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-rem-batch"))
		require.NoError(
			t,
			provider.ZAdd(
				ctx,
				"zset-rem-batch",
				cache.Z{Score: 1, Member: "first"},
				cache.Z{Score: 2, Member: "second"},
			),
		)

		removed, err := provider.ZRem(ctx, "zset-rem-batch", "first", "second", "absent")
		require.NoError(t, err)
		assert.Equal(t, int64(2), removed)
	})

	t.Run("missing key is not an error", func(t *testing.T) {
		require.NoError(t, provider.Del(ctx, "zset-rem-absent"))

		removed, err := provider.ZRem(ctx, "zset-rem-absent", "whatever")
		require.NoError(t, err)
		assert.Equal(t, int64(0), removed)
	})
}
