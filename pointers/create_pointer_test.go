package pointers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DKhorkov/libs/pointers"
)

func TestCreatePointer(t *testing.T) {
	t.Run("simple pointer", func(t *testing.T) {
		value := 2
		ptr := pointers.Pointer(value)
		assert.Equal(t, &value, ptr)
		assert.IsType(t, &value, ptr)
	})
}
