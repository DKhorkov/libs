package cache

import (
	"time"
)

const (
	defaultHost = "localhost"
	defaultPort = 6379
)

// newOptions creates *options with default values.
func newOptions() *options {
	return &options{
		host: defaultHost,
		port: defaultPort,
	}
}

// options represents options for configuration.
type options struct {
	host string
	port int

	// clientName will execute the `CLIENT SETNAME ClientName` command for each conn.
	clientName string

	// username is used to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	username string

	// password is an optional password. Must match the password specified in the
	// `requirepass` server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	password string

	// db is the database to be selected after connecting to the server.
	db int

	// maxRetries is the maximum number of retries before giving up.
	// -1 (not 0) disables retries.
	//
	// default: 3 retries
	maxRetries int

	// minRetryBackoff is the minimum backoff between each retry.
	// -1 disables backoff.
	//
	// default: 8 milliseconds
	minRetryBackoff time.Duration

	// maxRetryBackoff is the maximum backoff between each retry.
	// -1 disables backoff.
	// default: 512 milliseconds;
	maxRetryBackoff time.Duration

	// dialTimeout for establishing new connections.
	//
	// default: 5 seconds
	dialTimeout time.Duration

	// readTimeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Supported values:
	//
	//	- `-1` - no timeout (block indefinitely).
	//	- `-2` - disables SetReadDeadline calls completely.
	//
	// default: 3 seconds
	readTimeout time.Duration

	// writeTimeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.  Supported values:
	//
	//	- `-1` - no timeout (block indefinitely).
	//	- `-2` - disables SetWriteDeadline calls completely.
	//
	// default: 3 seconds
	writeTimeout time.Duration

	// contextTimeoutEnabled controls whether the client respects context timeouts and deadlines.
	// See https://redis.uptrace.dev/guide/go-redis-debugging.html#timeouts
	contextTimeoutEnabled bool

	// poolFIFO type of connection pool.
	//
	//	- true for FIFO pool
	//	- false for LIFO pool.
	//
	// Note that FIFO has slightly higher overhead compared to LIFO,
	// but it helps closing idle connections faster reducing the pool size.
	poolFIFO bool

	// poolSize is the base number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	// If there is not enough connections in the pool, new connections will be allocated in excess of PoolSize,
	// you can limit it through MaxActiveConns
	//
	// default: 10 * runtime.GOMAXPROCS(0)
	poolSize int

	// poolTimeout is the amount of time client waits for connection if all connections
	// are busy before returning an error.
	//
	// default: ReadTimeout + 1 second
	poolTimeout time.Duration

	// minIdleConnections is the minimum number of idle connections which is useful when establishing
	// new connection is slow. The idle connections are not closed by default.
	//
	// default: 0
	minIdleConnections int

	// maxIdleConnections is the maximum number of idle connections.
	// The idle connections are not closed by default.
	//
	// default: 0
	maxIdleConnections int

	// maxActiveConnections is the maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	// If the pool is full, the next call to Get() will block until a connection is released.
	maxActiveConnections int

	// connectionMaxIdleTime is the maximum amount of time a connection may be idle.
	// Should be less than server's timeout.
	//
	// Expired connections may be closed lazily before reuse.
	// If d <= 0, connections are not closed due to a connection's idle time.
	// -1 disables idle timeout check.
	//
	// default: 30 minutes
	connectionMaxIdleTime time.Duration

	// connectionMaxLifetime is the maximum amount of time a connection may be reused.
	//
	// Expired connections may be closed lazily before reuse.
	// If <= 0, connections are not closed due to a connection's age.
	//
	// default: 0
	connectionMaxLifetime time.Duration
}

// Option represents golang functional option pattern func for configuration.
type Option func(options *options) error

func WithHost(host string) Option {
	return func(options *options) error {
		options.host = host

		return nil
	}
}

func WithPort(port int) Option {
	return func(options *options) error {
		options.port = port

		return nil
	}
}

func WithClientName(clientName string) Option {
	return func(options *options) error {
		options.clientName = clientName

		return nil
	}
}

func WithUsername(username string) Option {
	return func(options *options) error {
		options.username = username

		return nil
	}
}

func WithPassword(password string) Option {
	return func(options *options) error {
		options.password = password

		return nil
	}
}

func WithDB(db int) Option {
	return func(options *options) error {
		options.db = db

		return nil
	}
}

func WithMaxRetries(maxRetries int) Option {
	return func(options *options) error {
		options.maxRetries = maxRetries

		return nil
	}
}

func WithMinRetryBackoff(minRetryBackoff time.Duration) Option {
	return func(options *options) error {
		options.minRetryBackoff = minRetryBackoff

		return nil
	}
}

func WithMaxRetryBackoff(maxRetryBackoff time.Duration) Option {
	return func(options *options) error {
		options.maxRetryBackoff = maxRetryBackoff

		return nil
	}
}

func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(options *options) error {
		options.dialTimeout = dialTimeout

		return nil
	}
}

func WithReadTimeout(readTimeout time.Duration) Option {
	return func(options *options) error {
		options.readTimeout = readTimeout

		return nil
	}
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(options *options) error {
		options.writeTimeout = writeTimeout

		return nil
	}
}

func WithContextTimeoutEnabled(contextTimeoutEnabled bool) Option {
	return func(options *options) error {
		options.contextTimeoutEnabled = contextTimeoutEnabled

		return nil
	}
}

func WithPoolFIFO(poolFIFO bool) Option {
	return func(options *options) error {
		options.poolFIFO = poolFIFO

		return nil
	}
}

func WithPoolSize(poolSize int) Option {
	return func(options *options) error {
		options.poolSize = poolSize

		return nil
	}
}

func WithPoolTimeout(poolTimeout time.Duration) Option {
	return func(options *options) error {
		options.poolTimeout = poolTimeout

		return nil
	}
}

func WithMinIdleConnections(minIdleConnections int) Option {
	return func(options *options) error {
		options.minIdleConnections = minIdleConnections

		return nil
	}
}

func WithMaxIdleConnections(maxIdleConnections int) Option {
	return func(options *options) error {
		options.maxIdleConnections = maxIdleConnections

		return nil
	}
}

func WithMaxActiveConnections(maxActiveConnections int) Option {
	return func(options *options) error {
		options.maxActiveConnections = maxActiveConnections

		return nil
	}
}

func WithConnectionMaxIdleTime(connectionMaxIdleTime time.Duration) Option {
	return func(options *options) error {
		options.connectionMaxIdleTime = connectionMaxIdleTime

		return nil
	}
}

func WithConnectionMaxLifetime(connectionMaxLifetime time.Duration) Option {
	return func(options *options) error {
		options.connectionMaxLifetime = connectionMaxLifetime

		return nil
	}
}
