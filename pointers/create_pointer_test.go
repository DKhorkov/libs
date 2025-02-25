package pointers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DKhorkov/libs/pointers"
)

func TestNewPointer(t *testing.T) {
	t.Run("simple pointer", func(t *testing.T) {
		value := 2
		ptr := pointers.New(value)
		assert.Equal(t, &value, ptr)
		assert.IsType(t, &value, ptr)
	})
}
