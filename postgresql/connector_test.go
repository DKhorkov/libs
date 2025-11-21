package postgresql_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	loggermock "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/postgresql"
)

var (
	driver = "sqlite3"
	dsn    = "file::memory:?cache=shared"
)

func TestTransaction(t *testing.T) {
	t.Run("should return transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql.New(dsn, driver, logger)
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
		connector := &postgresql.CommonConnector{}

		transaction, err := connector.Transaction(context.Background())
		require.Error(t, err)
		assert.IsTypef(
			t,
			&postgresql.NilDBConnectionError{},
			err,
			"error should be %T", &postgresql.NilDBConnectionError{},
		)

		assert.Nil(t, transaction)
	})

	t.Run("transaction with options", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql.New(dsn, driver, logger)
		require.NoError(t, err)

		defer func() {
			err = connector.Close()
			require.NoError(t, err)
		}()

		transaction, err := connector.Transaction(
			context.Background(),
			postgresql.WithTransactionReadOnly(true),
			postgresql.WithTransactionIsolationLevel(sql.LevelReadUncommitted),
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
		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql.New(dsn, driver, logger)
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
		connector := &postgresql.CommonConnector{}

		connection, err := connector.Connection(context.Background())
		require.Error(t, err)
		require.Error(t, err)
		assert.IsTypef(
			t,
			&postgresql.NilDBConnectionError{},
			err,
			"error should be %T", &postgresql.NilDBConnectionError{},
		)

		assert.Nil(t, connection)
	})
}

func TestNewConnector(t *testing.T) {
	t.Run("new connector without options", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql.New(dsn, driver, logger)
		require.NoError(t, err)
		require.NotNil(t, connector)

		err = connector.Close()
		require.NoError(t, err)
	})

	t.Run("new connector with options", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql.New(
			dsn,
			driver,
			logger,
			postgresql.WithMaxConnectionIdleTime(time.Minute),
			postgresql.WithMaxConnectionLifetime(time.Minute),
			postgresql.WithMaxIdleConnections(1),
			postgresql.WithMaxOpenConnections(1),
		)

		require.NoError(t, err)
		require.NotNil(t, connector)

		err = connector.Close()
		require.NoError(t, err)
	})
}

func TestPool(t *testing.T) {
	t.Run("should return connections pool", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql.New(dsn, driver, logger)
		require.NoError(t, err)

		assert.NotNil(t, connector.Pool())

		err = connector.Close()
		require.NoError(t, err)
	})

	t.Run("nil connections pool", func(t *testing.T) {
		connector := &postgresql.CommonConnector{}
		assert.Nil(t, connector.Pool())
	})
}
