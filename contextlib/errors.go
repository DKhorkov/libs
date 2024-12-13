package contextlib

import "fmt"

type ValueNotFoundError struct {
	Message string
	BaseErr error
}

func (e ValueNotFoundError) Error() string {
	template := "context with value %s not found"
	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}
