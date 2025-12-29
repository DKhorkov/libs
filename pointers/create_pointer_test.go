package pointers_test

import (
	"testing"

	"github.com/DKhorkov/libs/pointers"
	"github.com/stretchr/testify/assert"
)

func TestNewPointer(t *testing.T) {
	t.Parallel()

	t.Run("simple pointer", func(t *testing.T) {
		t.Parallel()

		value := 2
		ptr := pointers.New(value)
		assert.Equal(t, &value, ptr)
		assert.IsType(t, &value, ptr)
	})
}
