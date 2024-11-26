package security

// InvalidJWTError is an error, which represents, that JWT expired or something else went wrong via parsing it.
type InvalidJWTError struct {
	Message string
}

func (e InvalidJWTError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "JWT token is invalid or has expired"
}

// JWTClaimsError is an error, which represents, that failed to retrieve JWT payload.
type JWTClaimsError struct {
	Message string
}

func (e JWTClaimsError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "JWT claims error"
}
