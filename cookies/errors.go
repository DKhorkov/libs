package cookies

import "fmt"

// NotFoundError represents, that there is no http.Cookie with provided name in http.ResponseWriter.
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

func (e NotFoundError) Unwrap() error {
	return e.BaseErr
}
