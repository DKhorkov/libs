package db

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq" // Postgres driver

	"github.com/DKhorkov/libs/logging"
)

// New is constructor of CommonConnector. Gets database Config and logging.Logger to create an instance.
func New(dsn, driver string, logger logging.Logger, opts ...PoolOption) (*CommonConnector, error) {
	pool, err := connect(dsn, driver, opts...)
	if err != nil {
		return nil, err
	}

	dbConnector := &CommonConnector{
		connectionsPool: pool,
		logger:          logger,
	}

	return dbConnector, nil
}

// CommonConnector is base connector to work with database.
type CommonConnector struct {
	connectionsPool *sql.DB
	logger          logging.Logger
}

// connect connects to database and stores connections pool for later usage.
func connect(dsn, driver string, opts ...PoolOption) (*sql.DB, error) {
	var options poolOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	pool, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(); err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(options.maxOpenConnections)
	pool.SetMaxIdleConns(options.maxIdleConnections)
	pool.SetConnMaxLifetime(options.maxConnectionLifetime)
	pool.SetConnMaxIdleTime(options.maxConnectionIdleTime)
	return pool, nil
}

// Connection creates connection with database, if not exists. Returns connection for external usage.
func (connector *CommonConnector) Connection(ctx context.Context) (*sql.Conn, error) {
	if connector.connectionsPool == nil {
		return nil, &NilDBConnectionError{}
	}

	return connector.connectionsPool.Conn(ctx)
}

// Transaction return database transaction object for external usage with atomicity of operations.
func (connector *CommonConnector) Transaction(ctx context.Context, opts ...TransactionOption) (*sql.Tx, error) {
	if connector.connectionsPool == nil {
		return nil, &NilDBConnectionError{}
	}

	var options transactionOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	return connector.connectionsPool.BeginTx(
		ctx,
		&sql.TxOptions{
			ReadOnly:  options.readOnly,
			Isolation: options.isolationLevel,
		},
	)
}

// Pool returns database connections pool.
func (connector *CommonConnector) Pool() *sql.DB {
	return connector.connectionsPool
}

// Close closes pool of connections.
func (connector *CommonConnector) Close() error {
	if connector.connectionsPool == nil {
		return nil
	}

	return connector.connectionsPool.Close()
}
