package db_test

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

func TestBuildDsn(t *testing.T) {
	expected := "host=0.0.0.0 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	config := db.Config{
		Host:         "0.0.0.0",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
		SSLMode:      "disable",
		Driver:       "postgres",
	}

	actual := db.BuildDsn(config)
	assert.Equal(t, expected, actual)
}
