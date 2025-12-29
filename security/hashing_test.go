package security_test

import (
	"testing"

	"github.com/DKhorkov/libs/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		hashCost      int
		value         string
		message       string
		errorExpected bool
	}{
		{
			name:          "value successfully hashed",
			hashCost:      14,
			value:         "value",
			message:       "should return hash for value",
			errorExpected: false,
		},
		{
			name:          "too long value > 72 bytes",
			hashCost:      14,
			value:         "tooLongValueThatCanNotBeLessThanSeventyTwoBytesForSureAndThereCouldAlsoBeSomeStory",
			message:       "should return error",
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hashedValue, err := security.Hash(tc.value, tc.hashCost)

			if tc.errorExpected {
				require.Error(t, err, tc.message)
				assert.Empty(
					t,
					hashedValue,
					"\n%s - actual: '%v', expected: '%v'", tc.message, hashedValue, "")
			} else {
				require.NoError(t, err, tc.message)
				assert.NotEmpty(
					t,
					hashedValue,
					"\n%s - actual: '%v', expected: '%v'", tc.message, hashedValue, "SomeHashedValue")
			}
		})
	}
}

func TestValidateHash(t *testing.T) {
	t.Parallel()

	valueToHash := "value"
	testCases := []struct {
		name     string
		expected bool
		value    string
		message  string
	}{
		{
			name:     "hashed value was created based on provided value",
			value:    valueToHash,
			expected: true,
			message:  "should return true",
		},
		{
			name:     "hash value was not created based on provided value\"",
			value:    "Incorrectvalue",
			expected: false,
			message:  "should return false",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hashedValue, _ := security.Hash(valueToHash, 0)
			isValid := security.ValidateHash(tc.value, hashedValue)

			assert.Equal(
				t,
				tc.expected,
				isValid,
				"\n%s - actual: '%v', expected: '%v'", tc.message, isValid, tc.expected)
		})
	}
}
