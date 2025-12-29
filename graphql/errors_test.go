package graphql_test

import (
	"errors"
	"fmt"
	"testing"

	customgraphql "github.com/DKhorkov/libs/graphql"
	"github.com/stretchr/testify/require"
)

func TestParseError(t *testing.T) {
	t.Parallel()

	t.Run("Default message without base error", func(t *testing.T) {
		t.Parallel()

		err := customgraphql.ParseError{}
		expected := "failed to parse GraphQL query"
		require.Equal(t, expected, err.Error())
		require.NoError(t, err.Unwrap())
	})

	t.Run("Custom message without base error", func(t *testing.T) {
		t.Parallel()

		err := customgraphql.ParseError{
			Message: "custom parse error",
		}
		expected := "custom parse error"
		require.Equal(t, expected, err.Error())
		require.NoError(t, err.Unwrap())
	})

	t.Run("Custom message with base error", func(t *testing.T) {
		t.Parallel()

		baseErr := errors.New("base error")
		err := customgraphql.ParseError{
			Message: "custom parse error",
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("custom parse error. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})

	t.Run("Default message with base error", func(t *testing.T) {
		t.Parallel()

		baseErr := errors.New("base error")
		err := customgraphql.ParseError{
			BaseErr: baseErr,
		}
		expected := fmt.Sprintf("failed to parse GraphQL query. Base error: %v", baseErr)
		require.Equal(t, expected, err.Error())
		require.Equal(t, baseErr, err.Unwrap())
	})
}
