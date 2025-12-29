package contextlib_test

import (
	"context"
	"testing"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValueFromContextAndSetValueToContext(t *testing.T) {
	t.Parallel()

	t.Run("Set value to context and Get value from context successfully", func(t *testing.T) {
		t.Parallel()

		var (
			key   = "key"
			value = "value"
		)

		// Only that way due to inner type:
		ctx := contextlib.WithValue(context.Background(), key, value)

		result, err := contextlib.ValueFromContext[string](ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("Get value from context fail", func(t *testing.T) {
		t.Parallel()

		result, err := contextlib.ValueFromContext[string](context.Background(), "anyKey")
		require.Error(t, err)
		require.IsType(t, &contextlib.ValueNotFoundError{}, err)
		assert.Empty(t, result)
	})
}
