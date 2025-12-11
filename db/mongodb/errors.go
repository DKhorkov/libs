package mongodb

import "fmt"

// NilClientError is an error, representing no connection established with mongoDB.
type NilClientError struct {
	Message string
	BaseErr error
}

func (e NilClientError) Error() string {
	template := "Mongo client error. Making operation on nil client"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e NilClientError) Unwrap() error {
	return e.BaseErr
}

// NilDatabaseError is an error, representing no mongo.Database provided.
type NilDatabaseError struct {
	Message string
	BaseErr error
}

func (e NilDatabaseError) Error() string {
	template := "Mongo Database error. Making operation on nil Database"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e NilDatabaseError) Unwrap() error {
	return e.BaseErr
}
