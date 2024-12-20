package db

import (
	"time"
)

// poolOptions represents options for *sql.DB configuration.
type poolOptions struct {
	maxOpenConnections    int
	maxIdleConnections    int
	maxConnectionLifetime time.Duration
	maxConnectionIdleTime time.Duration
}

// PoolOption represents golang functional option pattern func for connections pool settings.
type PoolOption func(options *poolOptions) error

// WithMaxOpenConnections sets maximum opened connections in database connections pool.
func WithMaxOpenConnections(num int) PoolOption {
	return func(options *poolOptions) error {
		options.maxOpenConnections = num
		return nil
	}
}

// WithMaxIdleConnections sets maximum idle connections in database connections pool.
func WithMaxIdleConnections(num int) PoolOption {
	return func(options *poolOptions) error {
		options.maxIdleConnections = num
		return nil
	}
}

// WithMaxConnectionLifetime sets maximum connection lifetime before closure in database connections pool.
func WithMaxConnectionLifetime(lifetime time.Duration) PoolOption {
	return func(options *poolOptions) error {
		options.maxConnectionLifetime = lifetime
		return nil
	}
}

// WithMaxConnectionIdleTime sets maximum connection idle time (without usage) before closure in database
// connections pool.
func WithMaxConnectionIdleTime(idleTime time.Duration) PoolOption {
	return func(options *poolOptions) error {
		options.maxConnectionIdleTime = idleTime
		return nil
	}
}
