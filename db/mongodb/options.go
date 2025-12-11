package mongodb

import (
	"time"
)

// options represents options for *mongo.Client configuration.
type options struct {
	username                 string
	password                 string
	authSource               string // The database where the user is stored
	maxConnections           uint64
	maxPoolSize              uint64
	minPoolSize              uint64
	maxConnectionTimeout     time.Duration
	maxConnectionIdleTimeout time.Duration
}

// Option represents golang functional option pattern func for mongo.Client configuration.
type Option func(options *options) error

// WithUsername sets username for mongo.Client connection.
func WithUsername(username string) Option {
	return func(options *options) error {
		options.username = username

		return nil
	}
}

// WithPassword sets password for mongo.Client connection.
func WithPassword(password string) Option {
	return func(options *options) error {
		options.password = password

		return nil
	}
}

// WithAuthSource sets AuthSource for mongo.Client connection.
func WithAuthSource(authSource string) Option {
	return func(options *options) error {
		options.authSource = authSource

		return nil
	}
}

// WithMaxConnections sets maximum opened connections.
func WithMaxConnections(num uint64) Option {
	return func(options *options) error {
		options.maxConnections = num

		return nil
	}
}

// WithMaxPoolSize specifies that maximum number of connections allowed in the driver's connection pool to each server.
func WithMaxPoolSize(num uint64) Option {
	return func(options *options) error {
		options.maxPoolSize = num

		return nil
	}
}

// WithMinPoolSize specifies that minimum number of connections allowed in the driver's connection pool to each server.
func WithMinPoolSize(num uint64) Option {
	return func(options *options) error {
		options.minPoolSize = num

		return nil
	}
}

// WithMaxConnectionTimeout sets maximum amount of time for connection lifetime.
func WithMaxConnectionTimeout(timeout time.Duration) Option {
	return func(options *options) error {
		options.maxConnectionTimeout = timeout

		return nil
	}
}

// WithMaxConnectionIdleTime sets maximum amount of time for connection idle lifetime.
func WithMaxConnectionIdleTime(timeout time.Duration) Option {
	return func(options *options) error {
		options.maxConnectionIdleTimeout = timeout

		return nil
	}
}
