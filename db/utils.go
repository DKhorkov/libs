package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DKhorkov/libs/logging"
	"log/slog"
	"reflect"
)

// GetEntityColumns receives a POINTER on entity (NOT A VALUE), parses is using reflection and returns
// a slice of columns for db/sql Query() method purpose for retrieving data from result rows.
// https://stackoverflow.com/questions/56525471/how-to-use-rows-scan-of-gos-database-sql
func GetEntityColumns(entity interface{}) []interface{} {
	structure := reflect.ValueOf(entity).Elem()
	numCols := structure.NumField()
	columns := make([]interface{}, numCols)
	for i := range numCols {
		field := structure.Field(i)
		columns[i] = field.Addr().Interface()
	}

	return columns
}

func BuildDsn(config Config) string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DatabaseName,
		config.SSLMode,
	)

	return dsn
}

func CloseConnectionContext(ctx context.Context, connection *sql.Conn, logger *slog.Logger) {
	if err := connection.Close(); err != nil {
		logging.LogErrorContext(ctx, logger, "Failed to close connection", err)
	}
}
