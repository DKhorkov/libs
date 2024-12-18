package cookies

import "fmt"

type NotFoundError struct {
	Message string
	BaseErr error
}

func (e NotFoundError) Error() string {
	template := "%s cookie not found"
	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}
