package nats_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	customnats "github.com/DKhorkov/libs/nats"
)

func TestWorkerAlreadyRunningError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := customnats.WorkerAlreadyRunningError{}
		expected := "worker is already running"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := customnats.WorkerAlreadyRunningError{
			Message: "custom worker running error",
		}
		expected := "custom worker running error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.WorkerAlreadyRunningError{
			Message: "custom worker running error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom worker running error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.WorkerAlreadyRunningError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("worker is already running. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}

func TestWorkerAlreadyStoppedError(t *testing.T) {
	t.Run("Default message without base error", func(t *testing.T) {
		err := customnats.WorkerAlreadyStoppedError{}
		expected := "worker is already stopped"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		err := customnats.WorkerAlreadyStoppedError{
			Message: "custom worker stopped error",
		}
		expected := "custom worker stopped error"
		require.Equal(t, expected, err.Error())
		require.Nil(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.WorkerAlreadyStoppedError{
			Message: "custom worker stopped error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom worker stopped error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customnats.WorkerAlreadyStoppedError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("worker is already stopped. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
