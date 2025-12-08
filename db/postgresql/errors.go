package postgresql

import "fmt"

// NilDBConnectionError is an error, representing not being able to connect to database and create a connection pool.
type NilDBConnectionError struct {
	Message string
	BaseErr error
}

func (e NilDBConnectionError) Error() string {
	template := "DB connections pool error. Making operation on nil database connections pool"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e NilDBConnectionError) Unwrap() error {
	return e.BaseErr
}
