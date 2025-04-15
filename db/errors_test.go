package db_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/db"
)

func TestNilDBConnectionError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := db.NilDBConnectionError{}
		expected := "DB connections pool error. Making operation on nil database connections pool"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := db.NilDBConnectionError{
			Message: "custom db connection error",
		}
		expected := "custom db connection error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := db.NilDBConnectionError{
			Message: "custom db connection error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom db connection error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := db.NilDBConnectionError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("DB connections pool error. Making operation on nil database connections pool. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
