package database__test

import (
	"database/sql"
	"testing"

	"github.com/DKhorkov/libs/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestDatabaseConnect(t *testing.T) {
	testsConfig := db.NewTestConfig()

	t.Run("should connect to database", func(t *testing.T) {
		connector := &db.CommonDBConnector{
			DSN:    testsConfig.DSN,
			Driver: testsConfig.Driver,
		}

		err := connector.Connect()
		require.NoError(t, err)
	})

	t.Run("should fail to connect to non existing database", func(t *testing.T) {
		connector := &db.CommonDBConnector{
			DSN:    "non existing database",
			Driver: "error",
		}

		err := connector.Connect()
		require.Error(t, err)
	})
}

func TestDatabaseGetTransaction(t *testing.T) {
	testsConfig := db.NewTestConfig()

	t.Run("should return transaction", func(t *testing.T) {
		connector := &db.CommonDBConnector{
			DSN:    testsConfig.DSN,
			Driver: testsConfig.Driver,
		}

		if err := connector.Connect(); err != nil {
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

	t.Run("should fail to get transaction from nil connection", func(t *testing.T) {
		connector := &db.CommonDBConnector{
			DSN:    "non existing database",
			Driver: "error",
		}

		transaction, err := connector.GetTransaction()
		require.Error(t, err)
		assert.IsTypef(
			t,
			&db.NilDBConnectionError{},
			err,
			"should be customerrors.NilDBConnectionError")
		assert.Nil(t, transaction)
	})
}

func TestDatabaseGetConnection(t *testing.T) {
	testsConfig := db.NewTestConfig()

	t.Run("should return connection", func(t *testing.T) {
		connector := &db.CommonDBConnector{
			DSN:    testsConfig.DSN,
			Driver: testsConfig.Driver,
		}

		if err := connector.Connect(); err != nil {
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
		connector := &db.CommonDBConnector{
			DSN:    testsConfig.DSN,
			Driver: testsConfig.Driver,
		}

		connection := connector.GetConnection()
		assert.NotNil(t, connection)
		assert.IsTypef(
			t,
			&sql.DB{},
			connection,
			"connection type should be sql.DB")
	})

	t.Run("should return nil if connect to database is not possible", func(t *testing.T) {
		connector := &db.CommonDBConnector{
			DSN:    "non existing database",
			Driver: "error",
		}

		connection := connector.GetConnection()
		assert.Nil(t, connection)
	})
}
