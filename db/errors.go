package db

// NilDBConnectionError is an error, representing not being able to connect to database and create a connection object.
type NilDBConnectionError struct {
	Message string
}

func (e NilDBConnectionError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "DB connection error. Making operation on nil database connection."
}
