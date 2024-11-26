package loadenv__test

import (
	"testing"

	"github.com/DKhorkov/libs/loadenv"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	testCases := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue string
		expected     string
		message      string
	}{
		{
			name:         "key exists",
			envVar:       "TEST_KEY",
			envValue:     "GRAPHQL_PORT",
			defaultValue: "",
			expected:     "GRAPHQL_PORT",
			message:      "should return value from env",
		},
		{
			name:         "key does not exist",
			envVar:       "NON_EXISTENT_KEY",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
			message:      "should return default value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				t.Setenv(tc.envVar, tc.envValue)
			}
			actual := loadenv.GetEnv(tc.envVar, tc.defaultValue)
			assert.Equal(
				t,
				tc.expected,
				actual,
				"\n%s - actual: '%v', expected: '%v'", tc.message, actual, tc.expected)
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	testCases := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue int
		expected     int
		message      string
	}{
		{
			name:         "env var exists and is valid integer",
			envVar:       "TEST_INT",
			envValue:     "8081",
			defaultValue: 8080,
			expected:     8081,
			message:      "should return int value from env",
		},
		{
			name:         "env var exists but is invalid integer",
			envVar:       "TEST_INVALID_INT",
			envValue:     "abc",
			defaultValue: 8080,
			expected:     8080,
			message:      "should return default value if env is invalid int",
		},
		{
			name:         "env var does not exist",
			envVar:       "NON_EXISTENT_KEY",
			envValue:     "",
			defaultValue: 8080,
			expected:     8080,
			message:      "should return default value if env does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				t.Setenv(tc.envVar, tc.envValue)
			}
			actual := loadenv.GetEnvAsInt(tc.envVar, tc.defaultValue)
			assert.Equal(
				t,
				tc.expected,
				actual,
				"\n%s - actual: '%v', expected: '%v'", tc.message, actual, tc.expected)
		})
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	testCases := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue []string
		separator    string
		expected     []string
		message      string
	}{
		{
			name:         "env var exists but is invalid slice",
			envVar:       "TEST_INVALID_SLICE",
			envValue:     "fs",
			defaultValue: []string{"1", "2"},
			separator:    ",",
			expected:     []string{"1", "2"},
			message:      "should return default value if env is invalid slice",
		},
		{
			name:         "env var exists but is invalid slice with different separator",
			envVar:       "TEST_INVALID_SLICE",
			envValue:     "fs",
			defaultValue: []string{"1", "2"},
			separator:    "|",
			expected:     []string{"1", "2"},
			message:      "should return default value if env is invalid slice with different separator",
		},
		{
			name:         "env var exists and is valid slice",
			envVar:       "TEST_VALID_SLICE",
			envValue:     "fs,a,ass",
			defaultValue: []string{"1", "2"},
			separator:    ",",
			expected:     []string{"fs", "a", "ass"},
			message:      "should return slice from env",
		},
		{
			name:         "env var exists and is valid slice with different separator",
			envVar:       "TEST_VALID_SLICE",
			envValue:     "fs|a|ass",
			defaultValue: []string{"1", "2"},
			separator:    "|",
			expected:     []string{"fs", "a", "ass"},
			message:      "should return slice from env with different separator",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				t.Setenv(tc.envVar, tc.envValue)
			}
			actual := loadenv.GetEnvAsSlice(tc.envVar, tc.defaultValue, tc.separator)
			assert.Equal(
				t,
				tc.expected,
				actual,
				"\n%s - actual: '%v', expected: '%v'", tc.message, actual, tc.expected)
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	testCases := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue bool
		expected     bool
		message      string
	}{
		{
			name:         "env var exists and is true",
			envVar:       "TEST_BOOL_TRUE",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
			message:      "should return true if env is true",
		},
		{
			name:         "env var exists and is false",
			envVar:       "TEST_BOOL_FALSE",
			envValue:     "false",
			defaultValue: false,
			expected:     false,
			message:      "should return false if env is false",
		},
		{
			name:         "env var exists but is invalid boolean",
			envVar:       "TEST_BOOL_INVALID",
			envValue:     "invalid",
			defaultValue: false,
			expected:     false,
			message:      "should return default value if env is invalid bool",
		},
		{
			name:         "env var does not exist",
			envVar:       "TEST_BOOL_NOT_EXIST",
			envValue:     "",
			defaultValue: true,
			expected:     true,
			message:      "should return default value if env does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				t.Setenv(tc.envVar, tc.envValue)
			}
			actual := loadenv.GetEnvAsBool(tc.envVar, tc.defaultValue)
			assert.Equal(
				t,
				tc.expected,
				actual,
				"\n%s - actual: '%v', expected: '%v'", tc.message, actual, tc.expected)
		})
	}
}

func TestIsStringIsValidSlice(t *testing.T) {
	testCases := []struct {
		input     string
		separator string
		expected  bool
		message   string
	}{
		{
			input:     "Uno,Dos,Tres",
			separator: ",",
			expected:  true,
			message:   "should return 'True' for valid slice",
		},
		{
			input:     "Uno, ",
			separator: ",",
			expected:  false,
			message:   "should return 'False' for not valid slice",
		},
		{
			input:     "Uno, Dos, Tres",
			separator: ",",
			expected:  true,
			message:   "should return 'True' for valid slice with whitespaces between separated values",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			actual := loadenv.IsStringIsValidSlice(tc.input, tc.separator)
			assert.Equal(
				t,
				tc.expected,
				actual,
				"\n%s - actual: '%v', expected: '%v'", tc.message, actual, tc.expected)
		})
	}
}
