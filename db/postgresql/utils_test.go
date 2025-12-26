package postgresql_test

import (
	"context"
	"database/sql"
	"testing"

	postgresql2 "github.com/DKhorkov/libs/db/postgresql"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	loggermock "github.com/DKhorkov/libs/logging/mocks"
)

func TestGetEntityColumns(t *testing.T) {
	t.Run("should return slice of correct len and capacity", func(t *testing.T) {
		testStruct := &struct {
			Column1 int
			Column2 string
		}{}

		columns := postgresql2.GetEntityColumns(testStruct)
		assert.Len(t, columns, 2)
		assert.IsTypef(
			t,
			[]interface{}{},
			columns,
			"should return a slice of []interface{}")
	})
}

func TestBuildDsn(t *testing.T) {
	expected := "host=0.0.0.0 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	config := postgresql2.Config{
		Host:         "0.0.0.0",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
		SSLMode:      "disable",
		Driver:       "postgres",
	}

	actual := postgresql2.BuildDsn(config)
	assert.Equal(t, expected, actual)
}

func TestCloseConnectionContext(t *testing.T) {
	t.Run("should close connection context", func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		logger := loggermock.NewMockLogger(ctrl)
		connector, err := postgresql2.New(dsn, driver, logger)
		require.NoError(t, err)

		defer func() {
			if err = connector.Close(); err != nil {
				t.Fatal(err)
			}
		}()

		connection, err := connector.Connection(ctx)
		require.NoError(t, err)

		conn, ok := connection.(*sql.Conn)
		require.True(t, ok)

		postgresql2.CloseConnectionContext(ctx, conn, logger)
	})
}
