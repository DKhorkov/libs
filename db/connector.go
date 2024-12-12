package db

import (
	"database/sql"
	"log/slog"

	"github.com/DKhorkov/libs/logging"

	_ "github.com/lib/pq" // Postgres driver
)

// CommonDBConnector is base connector to work with database.
type CommonDBConnector struct {
	connection *sql.DB
	driver     string
	dsn        string
	logger     *slog.Logger
}

// Connect connects to database and stores database connection for later usage.
func (connector *CommonDBConnector) Connect() error {
	if connector.connection == nil {
		connection, err := sql.Open(connector.driver, connector.dsn)

		if err != nil {
			return err
		}

		if err = connection.Ping(); err != nil {
			return err
		}

		connector.connection = connection
	}

	return nil
}

// GetConnection creates connection with database, if not exists. Returns connection for external usage.
func (connector *CommonDBConnector) GetConnection() *sql.DB {
	if connector.connection == nil {
		if err := connector.Connect(); err != nil {
			return nil
		}
	}

	return connector.connection
}

// GetTransaction return database transaction object for external usage with atomicity of operations.
func (connector *CommonDBConnector) GetTransaction() (*sql.Tx, error) {
	if connector.connection == nil {
		return nil, &NilDBConnectionError{}
	}

	return connector.connection.Begin()
}

// CloseConnection closes stored connection to database.
func (connector *CommonDBConnector) CloseConnection() {
	if connector.connection == nil {
		return
	}

	if err := connector.connection.Close(); err != nil {
		connector.logger.Error(
			"Failed to close db connection",
			"Traceback",
			logging.GetLogTraceback(1),
			"Error",
			err,
		)
	}
}

// New is constructor of CommonDBConnector. Gets database Config and *slog.Logger to create an instance.
func New(dsn, driver string, logger *slog.Logger) (*CommonDBConnector, error) {
	dbConnector := &CommonDBConnector{
		driver: driver,
		dsn:    dsn,
		logger: logger,
	}

	if err := dbConnector.Connect(); err != nil {
		return nil, err
	}

	if err := dbConnector.GetConnection().Ping(); err != nil {
		return nil, err
	}

	return dbConnector, nil
}
