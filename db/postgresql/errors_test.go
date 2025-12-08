package postgresql_test

import (
	"errors"
	"fmt"
	"github.com/DKhorkov/libs/db/postgresql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNilDBConnectionError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := postgresql.NilDBConnectionError{}
		expected := "DB connections pool error. Making operation on nil database connections pool"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := postgresql.NilDBConnectionError{
			Message: "custom postgresql connection error",
		}
		expected := "custom postgresql connection error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := postgresql.NilDBConnectionError{
			Message: "custom postgresql connection error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom postgresql connection error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := postgresql.NilDBConnectionError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("DB connections pool error. Making operation on nil database connections pool. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
