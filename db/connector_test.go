package db_test

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	"github.com/DKhorkov/libs/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

var (
	driver = "sqlite3"
	dsn    = "file::memory:?cache=shared"
)

func TestTransaction(t *testing.T) {
	t.Run("should return transaction", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		defer func() {
			err = connector.Close()
			require.NoError(t, err)
		}()

		transaction, err := connector.Transaction(context.Background())
		require.NoError(t, err)
		assert.IsTypef(
			t,
			&sql.Tx{},
			transaction,
			"transaction type should be sql.Tx")
	})

	t.Run("nil connections pool", func(t *testing.T) {
		connector := &db.CommonDBConnector{}

		transaction, err := connector.Transaction(context.Background())
		require.Error(t, err)
		assert.IsTypef(
			t,
			&db.NilDBConnectionError{},
			err,
			"error should be %T", &db.NilDBConnectionError{},
		)

		assert.Nil(t, transaction)
	})

	t.Run("transaction with options", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		defer func() {
			err = connector.Close()
			require.NoError(t, err)
		}()

		transaction, err := connector.Transaction(
			context.Background(),
			db.WithTransactionReadOnly(true),
			db.WithTransactionIsolationLevel(sql.LevelReadUncommitted),
		)

		require.NoError(t, err)
		assert.IsTypef(
			t,
			&sql.Tx{},
			transaction,
			"transaction type should be sql.Tx")
	})
}

func TestConnection(t *testing.T) {
	t.Run("should return connection", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		defer func() {
			err = connector.Close()
			require.NoError(t, err)
		}()

		connection, err := connector.Connection(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, connection)
		assert.IsTypef(
			t,
			&sql.Conn{},
			connection,
			"connection type should be sql.Conn")

		err = connection.Close()
		require.NoError(t, err)
	})

	t.Run("nil connections pool", func(t *testing.T) {
		connector := &db.CommonDBConnector{}

		connection, err := connector.Connection(context.Background())
		require.Error(t, err)
		require.Error(t, err)
		assert.IsTypef(
			t,
			&db.NilDBConnectionError{},
			err,
			"error should be %T", &db.NilDBConnectionError{},
		)

		assert.Nil(t, connection)
	})
}

func TestNewConnector(t *testing.T) {
	t.Run("new connector without options", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)
		require.NotNil(t, connector)

		err = connector.Close()
		require.NoError(t, err)
	})

	t.Run("new connector with options", func(t *testing.T) {
		connector, err := db.New(
			dsn,
			driver,
			&slog.Logger{},
			db.WithMaxConnectionIdleTime(time.Minute),
			db.WithMaxConnectionLifetime(time.Minute),
			db.WithMaxIdleConnections(1),
			db.WithMaxOpenConnections(1),
		)

		require.NoError(t, err)
		require.NotNil(t, connector)

		err = connector.Close()
		require.NoError(t, err)
	})
}

func TestPool(t *testing.T) {
	t.Run("should return connections pool", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		assert.NotNil(t, connector.Pool())

		err = connector.Close()
		require.NoError(t, err)
	})

	t.Run("nil connections pool", func(t *testing.T) {
		connector := &db.CommonDBConnector{}
		assert.Nil(t, connector.Pool())
	})
}
