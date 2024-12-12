package security

import "fmt"

// InvalidJWTError is an error, which represents, that JWT expired or something else went wrong via parsing it.
type InvalidJWTError struct {
	Message string
	BaseErr error
}

func (e InvalidJWTError) Error() string {
	template := "JWT token is invalid or has expired"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}

// JWTClaimsError is an error, which represents, that failed to retrieve JWT payload.
type JWTClaimsError struct {
	Message string
	BaseErr error
}

func (e JWTClaimsError) Error() string {
	template := "JWT claims error"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}
