package database__test

import (
	"testing"

	"github.com/DKhorkov/libs/db"
	"github.com/stretchr/testify/assert"
)

func TestGetEntityColumns(t *testing.T) {
	t.Run("should return slice of correct len and capacity", func(t *testing.T) {
		testStruct := &struct {
			Column1 int
			Column2 string
		}{}

		columns := db.GetEntityColumns(testStruct)
		assert.Len(t, columns, 2)
		assert.IsTypef(
			t,
			[]interface{}{},
			columns,
			"should return a slice of []interface{}")
	})
}
