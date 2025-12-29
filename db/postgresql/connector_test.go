package postgresql_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	postgresql2 "github.com/DKhorkov/libs/db/postgresql"
	loggermock "github.com/DKhorkov/libs/logging/mocks"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	driver = "sqlite3"
	dsn    = "file::memory:?cache=shared"
)

func TestTransaction(t *testing.T) {
	t.Parallel()

	t.Run("should return transaction", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql2.New(dsn, driver, logger)
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
		t.Parallel()

		connector := &postgresql2.CommonConnector{}

		transaction, err := connector.Transaction(context.Background())
		require.Error(t, err)
		assert.IsTypef(
			t,
			&postgresql2.NilDBConnectionError{},
			err,
			"error should be %T", &postgresql2.NilDBConnectionError{},
		)

		assert.Nil(t, transaction)
	})

	t.Run("transaction with options", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql2.New(dsn, driver, logger)
		require.NoError(t, err)

		defer func() {
			err = connector.Close()
			require.NoError(t, err)
		}()

		transaction, err := connector.Transaction(
			context.Background(),
			postgresql2.WithTransactionReadOnly(true),
			postgresql2.WithTransactionIsolationLevel(sql.LevelReadUncommitted),
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
	t.Parallel()

	t.Run("should return connection", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql2.New(dsn, driver, logger)
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
		t.Parallel()

		connector := &postgresql2.CommonConnector{}

		connection, err := connector.Connection(context.Background())
		require.Error(t, err)
		require.Error(t, err)
		assert.IsTypef(
			t,
			&postgresql2.NilDBConnectionError{},
			err,
			"error should be %T", &postgresql2.NilDBConnectionError{},
		)

		assert.Nil(t, connection)
	})
}

func TestNewConnector(t *testing.T) {
	t.Parallel()

	t.Run("new connector without options", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql2.New(dsn, driver, logger)
		require.NoError(t, err)
		require.NotNil(t, connector)

		err = connector.Close()
		require.NoError(t, err)
	})

	t.Run("new connector with options", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql2.New(
			dsn,
			driver,
			logger,
			postgresql2.WithMaxConnectionIdleTime(time.Minute),
			postgresql2.WithMaxConnectionLifetime(time.Minute),
			postgresql2.WithMaxIdleConnections(1),
			postgresql2.WithMaxOpenConnections(1),
		)

		require.NoError(t, err)
		require.NotNil(t, connector)

		err = connector.Close()
		require.NoError(t, err)
	})
}

func TestPool(t *testing.T) {
	t.Parallel()

	t.Run("should return connections pool", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql2.New(dsn, driver, logger)
		require.NoError(t, err)

		assert.NotNil(t, connector.Pool())

		err = connector.Close()
		require.NoError(t, err)
	})

	t.Run("nil connections pool", func(t *testing.T) {
		t.Parallel()

		connector := &postgresql2.CommonConnector{}
		assert.Nil(t, connector.Pool())
	})
}
