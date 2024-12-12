package contextlib_test

import (
	"context"
	"testing"

	"github.com/DKhorkov/libs/contextlib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValueFromContextAndSetValueToContext(t *testing.T) {
	t.Run("Set value to context and Get value from context successfully", func(t *testing.T) {
		var (
			key   = "key"
			value = "value"
		)

		// Only that way due to inner type:
		ctx := contextlib.SetValueToContext(context.Background(), key, value)

		result, err := contextlib.GetValueFromContext[string](ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("Get value from context fail", func(t *testing.T) {
		result, err := contextlib.GetValueFromContext[string](context.Background(), "anyKey")
		require.Error(t, err)
		require.IsType(t, &contextlib.ContextValueNotFoundError{}, err)
		assert.Equal(t, "", result)
	})
}
