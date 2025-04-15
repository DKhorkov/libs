package contextlib_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/contextlib"
)

func TestValueNotFoundError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := contextlib.ValueNotFoundError{
			Message: "test-key",
		}
		expected := "context with value test-key not found"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Empty message without base error", func(t *testing.T) {
		err := contextlib.ValueNotFoundError{}
		expected := "context with value  not found"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := contextlib.ValueNotFoundError{
			Message: "test-key",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("context with value test-key not found. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Empty message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := contextlib.ValueNotFoundError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("context with value  not found. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
