package postgresql

import "time"

// Config is a database config, on base of which new connector is created.
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSLMode      string
	Driver       string
	Pool         PoolConfig
}

type PoolConfig struct {
	MaxOpenConnections    int
	MaxIdleConnections    int
	MaxConnectionLifetime time.Duration
	MaxConnectionIdleTime time.Duration
}
