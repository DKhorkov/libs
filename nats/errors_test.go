package nats_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	customnats "github.com/DKhorkov/libs/nats"
)

func TestConsumerAlreadyRunningError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := customnats.ConsumerAlreadyRunningError{}
		expected := "consumer is already running"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := customnats.ConsumerAlreadyRunningError{
			Message: "custom consumer running error",
		}
		expected := "custom consumer running error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.ConsumerAlreadyRunningError{
			Message: "custom consumer running error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom consumer running error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.ConsumerAlreadyRunningError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("consumer is already running. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}

func TestConsumerAlreadyStoppedError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := customnats.ConsumerAlreadyStoppedError{}
		expected := "consumer is already stopped"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := customnats.ConsumerAlreadyStoppedError{
			Message: "custom consumer stopped error",
		}
		expected := "custom consumer stopped error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.ConsumerAlreadyStoppedError{
			Message: "custom consumer stopped error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom consumer stopped error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.ConsumerAlreadyStoppedError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("consumer is already stopped. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
