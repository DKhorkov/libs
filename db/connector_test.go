package db_test

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/DKhorkov/libs/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

var (
	driver = "sqlite3"
	dsn    = "file::memory:?cache=shared"
)

func TestConnect(t *testing.T) {
	t.Run("should connect to database", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		err = connector.Connect()
		require.NoError(t, err)
	})

	t.Run("should fail to connect to non existing database", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		err = connector.Connect()
		require.NoError(t, err)
	})

	t.Run("should return error to unknown driver", func(t *testing.T) {
		driver, dsn := "fake", "fake"
		_, err := db.New(dsn, driver, &slog.Logger{})
		require.Error(t, err)
	})
}

func TestGetTransaction(t *testing.T) {
	t.Run("should return transaction", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		if err = connector.Connect(); err != nil {
			t.Fatal(err)
		}

		transaction, err := connector.GetTransaction()
		require.NoError(t, err)
		assert.IsTypef(
			t,
			&sql.Tx{},
			transaction,
			"transaction type should be sql.Tx")
	})
}

func TestGetConnection(t *testing.T) {
	t.Run("should return connection", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		if err = connector.Connect(); err != nil {
			t.Fatal(err)
		}

		connection := connector.GetConnection()
		assert.NotNil(t, connection)
		assert.IsTypef(
			t,
			&sql.DB{},
			connection,
			"connection type should be sql.DB")
	})

	t.Run("should return connection, even if it was nil", func(t *testing.T) {
		connector, err := db.New(dsn, driver, &slog.Logger{})
		require.NoError(t, err)

		connection := connector.GetConnection()
		assert.NotNil(t, connection)
		assert.IsTypef(
			t,
			&sql.DB{},
			connection,
			"connection type should be sql.DB")
	})
}
