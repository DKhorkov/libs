package validation

import "fmt"

// Error represents, that validation was not passed.
type Error struct {
	Message string
	BaseErr error
}

func (e Error) Error() string {
	template := "validation error"
	if e.Message != "" {
		template = fmt.Sprintf(template+": %s", e.Message)
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e Error) Unwrap() error {
	return e.BaseErr
}
