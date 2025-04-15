package cookies_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/cookies"
)

func TestNotFoundError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := cookies.NotFoundError{
			Message: "session_cookie",
		}
		expected := "session_cookie cookie not found"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Empty message without base error", func(t *testing.T) {
		err := cookies.NotFoundError{}
		expected := " cookie not found"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := cookies.NotFoundError{
			Message: "session_cookie",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("session_cookie cookie not found. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Empty message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := cookies.NotFoundError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf(" cookie not found. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
