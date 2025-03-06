package graphql

import "fmt"

// ParseError is an error, representing not being able to parse GraphQL request.
type ParseError struct {
	Message string
	BaseErr error
}

func (e ParseError) Error() string {
	template := "failed to parse GraphQL query"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}

func (e ParseError) Unwrap() error {
	return e.BaseErr
}
