package security_test

import (
	"testing"

	"github.com/DKhorkov/libs/security"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	testCases := []struct {
		name    string
		data    []byte
		message string
	}{
		{
			name:    "successfully encoded data",
			data:    []byte("someDataToEncode"),
			message: "provided data should be successfully encoded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hashedRefreshToken := security.Encode(tc.data)
			assert.NotEqual(
				t,
				"",
				hashedRefreshToken,
				tc.message)
		})
	}
}

func TestDecode(t *testing.T) {
	testCases := []struct {
		name          string
		encoded       string
		message       string
		errorExpected bool
	}{
		{
			name:          "successfully decoded data",
			encoded:       security.Encode([]byte("someDataToEncode")),
			errorExpected: false,
			message:       "should correctly decode data without error",
		},
		{
			name:          "failed to decode data due to incorrect base64 format",
			encoded:       "invalid base64 data",
			errorExpected: true,
			message:       "should raise an error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := security.Decode(tc.encoded)
			if tc.errorExpected {
				require.Error(t, err, tc.message)
				assert.Nil(
					t,
					value,
					tc.message)
			} else {
				require.NoError(t, err, tc.message)
				assert.NotEmpty(
					t,
					value,
					tc.message)
			}
		})
	}
}

func TestRawEncode(t *testing.T) {
	testCases := []struct {
		name    string
		data    []byte
		message string
	}{
		{
			name:    "successfully encoded data",
			data:    []byte("someDataToEncode"),
			message: "provided data should be successfully encoded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hashedRefreshToken := security.RawEncode(tc.data)
			assert.NotEqual(
				t,
				"",
				hashedRefreshToken,
				tc.message)
		})
	}
}

func TestRawDecode(t *testing.T) {
	testCases := []struct {
		name          string
		encoded       string
		message       string
		errorExpected bool
	}{
		{
			name:          "successfully decoded data",
			encoded:       security.RawEncode([]byte("someDataToEncode")),
			errorExpected: false,
			message:       "should correctly decode data without error",
		},
		{
			name:          "failed to decode data due to incorrect base64 format",
			encoded:       "invalid base64 data",
			errorExpected: true,
			message:       "should raise an error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := security.RawDecode(tc.encoded)
			if tc.errorExpected {
				require.Error(t, err, tc.message)
				assert.Nil(
					t,
					value,
					tc.message)
			} else {
				require.NoError(t, err, tc.message)
				assert.NotEmpty(
					t,
					value,
					tc.message)
			}
		})
	}
}
