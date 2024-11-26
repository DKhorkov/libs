package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/DKhorkov/libs/logging"

	_ "github.com/lib/pq" // Postgres driver
)

// CommonDBConnector is base connector to work with database.
type CommonDBConnector struct {
	Connection *sql.DB
	Driver     string
	DSN        string
	Logger     *slog.Logger
}

// Connect connects to database and stores database connection for later usage.
func (connector *CommonDBConnector) Connect() error {
	if connector.Connection == nil {
		connection, err := sql.Open(connector.Driver, connector.DSN)

		if err != nil {
			return err
		}

		connector.Connection = connection
	}

	return nil
}

// GetConnection creates connection with database, if not exists. Returns connection for external usage.
func (connector *CommonDBConnector) GetConnection() *sql.DB {
	if connector.Connection == nil {
		if err := connector.Connect(); err != nil {
			return nil
		}
	}

	return connector.Connection
}

// GetTransaction return database transaction object for external usage with atomicity of operations.
func (connector *CommonDBConnector) GetTransaction() (*sql.Tx, error) {
	if connector.Connection == nil {
		return nil, &NilDBConnectionError{}
	}

	return connector.Connection.Begin()
}

// CloseConnection closes stored connection to database.
func (connector *CommonDBConnector) CloseConnection() {
	if connector.Connection == nil {
		return
	}

	if err := connector.Connection.Close(); err != nil {
		connector.Logger.Error(
			"Failed to close db connection",
			"Traceback",
			logging.GetLogTraceback(),
			"Error",
			err,
		)
	}
}

// New is constructor of CommonDBConnector. Gets database Config and *slog.Logger to create an instance.
func New(dbConfig Config, logger *slog.Logger) (*CommonDBConnector, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DatabaseName,
		dbConfig.SSLMode,
	)

	dbConnector := &CommonDBConnector{
		Driver: dbConfig.Driver,
		DSN:    dsn,
		Logger: logger,
	}

	if err := dbConnector.Connect(); err != nil {
		return nil, err
	}

	return dbConnector, nil
}
