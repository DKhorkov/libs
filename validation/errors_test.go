package validation_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DKhorkov/libs/validation"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *validation.Error
		expected string
	}{
		{
			name:     "empty error",
			err:      &validation.Error{},
			expected: "validation error",
		},
		{
			name: "with message only",
			err: &validation.Error{
				Message: "invalid input",
			},
			expected: "validation error: invalid input",
		},
		{
			name: "with base error only",
			err: &validation.Error{
				BaseErr: errors.New("io error"),
			},
			expected: "validation error. Base error: io error",
		},
		{
			name: "with both message and base error",
			err: &validation.Error{
				Message: "invalid input",
				BaseErr: errors.New("io error"),
			},
			expected: "validation error: invalid input. Base error: io error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.err.Error()
			if actual != tc.expected {
				t.Errorf("Error() = %v, want %v", actual, tc.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	tests := []struct {
		name     string
		err      *validation.Error
		expected error
	}{
		{
			name:     "no base error",
			err:      &validation.Error{},
			expected: nil,
		},
		{
			name: "with simple base error",
			err: &validation.Error{
				BaseErr: baseErr,
			},
			expected: baseErr,
		},
		{
			name: "with wrapped error",
			err: &validation.Error{
				BaseErr: wrappedErr,
			},
			expected: wrappedErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.err.Unwrap()
			if !errors.Is(actual, tc.expected) {
				t.Errorf("Unwrap() = %v, want %v", actual, tc.expected)
			}
		})
	}
}
