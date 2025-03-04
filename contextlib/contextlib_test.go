package contextlib_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/contextlib"
)

func TestGetValueFromContextAndSetValueToContext(t *testing.T) {
	t.Run("Set value to context and Get value from context successfully", func(t *testing.T) {
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
		result, err := contextlib.ValueFromContext[string](context.Background(), "anyKey")
		require.Error(t, err)
		require.IsType(t, &contextlib.ValueNotFoundError{}, err)
		assert.Equal(t, "", result)
	})
}
