package db

import "fmt"

// NilDBConnectionError is an error, representing not being able to connect to database and create a connection object.
type NilDBConnectionError struct {
	Message string
	BaseErr error
}

func (e NilDBConnectionError) Error() string {
	template := "DB connection error. Making operation on nil database connection."
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}
