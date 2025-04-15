package security_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/security"
)

func TestInvalidJWTError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := security.InvalidJWTError{}
		expected := "JWT token is invalid or has expired"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := security.InvalidJWTError{
			Message: "custom JWT error",
		}
		expected := "custom JWT error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := security.InvalidJWTError{
			Message: "custom JWT error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom JWT error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := security.InvalidJWTError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("JWT token is invalid or has expired. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}

func TestJWTClaimsError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := security.JWTClaimsError{}
		expected := "JWT claims error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := security.JWTClaimsError{
			Message: "custom claims error",
		}
		expected := "custom claims error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := security.JWTClaimsError{
			Message: "custom claims error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom claims error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := security.JWTClaimsError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("JWT claims error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
