package grpc_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	customgrpc "github.com/DKhorkov/libs/grpc"
)

func TestBaseError_Error(t *testing.T) {
	t.Run("Message without base error", func(t *testing.T) {
		err := customgrpc.BaseError{
			Message: "something went wrong",
			Status:  codes.InvalidArgument,
		}
		expected := "something went wrong"
		require.Equal(t, expected, err.Error())
	})

	t.Run("Message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customgrpc.BaseError{
			Message: "something went wrong",
			BaseErr: baseErr,
			Status:  codes.InvalidArgument,
		}
		expected := fmt.Sprintf("something went wrong. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
	})

	t.Run("Empty message without base error", func(t *testing.T) {
		err := customgrpc.BaseError{
			Status: codes.InvalidArgument,
		}
		expected := ""
		require.Equal(t, expected, err.Error())
	})

	t.Run("Empty message with base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customgrpc.BaseError{
			BaseErr: baseErr,
			Status:  codes.InvalidArgument,
		}
		expected := fmt.Sprintf(". Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
	})
}

func TestBaseError_Unwrap(t *testing.T) {
	t.Run("With base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customgrpc.BaseError{
			Message: "something went wrong",
			BaseErr: baseErr,
			Status:  codes.InvalidArgument,
		}
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Without base error", func(t *testing.T) {
		err := customgrpc.BaseError{
			Message: "something went wrong",
			Status:  codes.InvalidArgument,
		}
		require.Nil(t, err.Unwrap())
	})
}

func TestBaseError_GRPCStatus(t *testing.T) {
	t.Run("With message and base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := customgrpc.BaseError{
			Message: "something went wrong",
			BaseErr: baseErr,
			Status:  codes.InvalidArgument,
		}
		expected := status.New(codes.InvalidArgument, fmt.Sprintf("something went wrong. Base error: %v", baseErr))
		result := err.GRPCStatus()
		require.Equal(t, expected.Code(), result.Code())
		require.Equal(t, expected.Message(), result.Message())
	})

	t.Run("With message only", func(t *testing.T) {
		err := customgrpc.BaseError{
			Message: "something went wrong",
			Status:  codes.InvalidArgument,
		}
		expected := status.New(codes.InvalidArgument, "something went wrong")
		result := err.GRPCStatus()
		require.Equal(t, expected.Code(), result.Code())
		require.Equal(t, expected.Message(), result.Message())
	})

	t.Run("Empty message without base error", func(t *testing.T) {
		err := customgrpc.BaseError{
			Status: codes.InvalidArgument,
		}
		expected := status.New(codes.InvalidArgument, "")
		result := err.GRPCStatus()
		require.Equal(t, expected.Code(), result.Code())
		require.Equal(t, expected.Message(), result.Message())
	})
}
