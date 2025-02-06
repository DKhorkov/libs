package security_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DKhorkov/libs/security"
)

func TestGenerateJWT(t *testing.T) {
	testCases := []struct {
		name          string
		secretKey     string
		algorithm     string
		value         any
		ttl           time.Duration
		message       string
		errorExpected bool
	}{
		{
			name:          "should generate valid token",
			secretKey:     "testSecret",
			algorithm:     "HS256",
			ttl:           time.Hour,
			value:         1,
			message:       "should return valid JWT token",
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := security.GenerateJWT(tc.value, tc.secretKey, tc.ttl, tc.algorithm)

			if tc.errorExpected {
				require.Error(t, err, tc.message)
				assert.Equal(
					t,
					"",
					token,
					"\n%s - actual: '%v', expected: '%v'", tc.message, token, "")
			} else {
				require.NoError(t, err, tc.message)
				assert.NotEqual(
					t,
					"",
					token,
					"\n%s - actual: '%v', expected: '%v'", tc.message, token, "SomeJWTValue")
			}
		})
	}
}

func TestParseJWT(t *testing.T) {
	const (
		userID    = 1
		secretKey = "testSecret"
	)

	testCases := []struct {
		name          string
		secretKey     string
		algorithm     string
		ttl           time.Duration
		message       string
		errorExpected bool
		errorType     error
		expected      any
	}{
		{
			name:          "correct JWT",
			secretKey:     secretKey,
			algorithm:     "HS256",
			ttl:           time.Hour,
			message:       "should return valid JWT token",
			errorExpected: false,
			errorType:     nil,
			expected:      userID,
		},
		{
			name:          "invalid secret key",
			secretKey:     "invalidSecret",
			algorithm:     "HS256",
			ttl:           time.Hour,
			message:       "should raise an error due to invalid secret key",
			errorExpected: true,
			errorType:     &security.InvalidJWTError{},
			expected:      nil,
		},
		{
			name:          "expired JWT",
			secretKey:     secretKey,
			algorithm:     "HS256",
			ttl:           time.Duration(0),
			message:       "should raise an error due to expired JWT",
			errorExpected: true,
			errorType:     &security.InvalidJWTError{},
			expected:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := security.GenerateJWT(userID, secretKey, tc.ttl, tc.algorithm)
			require.NoError(t, err, tc.message)

			value, err := security.ParseJWT(token, tc.secretKey)
			if tc.errorExpected {
				require.Error(t, err, tc.message)
				assert.IsType(t, tc.errorType, err)
			} else {
				floatValue, ok := value.(float64)
				require.True(t, ok)
				value = int(floatValue)
			}

			assert.Equal(
				t,
				tc.expected,
				value,
				"\n%s - actual: '%v', expected: '%v'", tc.message, value, tc.expected)
		})
	}
}
