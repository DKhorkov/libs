package mongodb

import (
	"errors"
	"testing"
)

func TestNilClientError(t *testing.T) {
	baseErr := errors.New("underlying error")

	tests := []struct {
		name        string
		err         *NilClientError
		expectedErr string
		expectedMsg string
	}{
		{
			name: "empty error with no base error",
			err: &NilClientError{
				Message: "",
				BaseErr: nil,
			},
			expectedErr: "Mongo client error. Making operation on nil client",
			expectedMsg: "Mongo client error. Making operation on nil client",
		},
		{
			name: "custom message with no base error",
			err: &NilClientError{
				Message: "Custom client error message",
				BaseErr: nil,
			},
			expectedErr: "Custom client error message",
			expectedMsg: "Custom client error message",
		},
		{
			name: "empty message with base error",
			err: &NilClientError{
				Message: "",
				BaseErr: baseErr,
			},
			expectedErr: "Mongo client error. Making operation on nil client. Base error: underlying error",
			expectedMsg: "Mongo client error. Making operation on nil client",
		},
		{
			name: "custom message with base error",
			err: &NilClientError{
				Message: "Failed to execute operation",
				BaseErr: baseErr,
			},
			expectedErr: "Failed to execute operation. Base error: underlying error",
			expectedMsg: "Failed to execute operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Тестируем Error()
			errorString := tt.err.Error()
			if errorString != tt.expectedErr {
				t.Errorf("Error() = %q, want %q", errorString, tt.expectedErr)
			}

			// Тестируем Unwrap()
			unwrapped := tt.err.Unwrap()
			if tt.err.BaseErr != nil && !errors.Is(unwrapped, tt.err.BaseErr) {
				t.Errorf("Unwrap() = %v, want %v", unwrapped, tt.err.BaseErr)
			}
			if tt.err.BaseErr == nil && unwrapped != nil {
				t.Errorf("Unwrap() = %v, want nil", unwrapped)
			}

			// Дополнительная проверка с errors.Is
			if tt.err.BaseErr != nil {
				if !errors.Is(tt.err, baseErr) {
					t.Errorf("errors.Is should return true for base error")
				}
			}
		})
	}
}

func TestNilDatabaseError(t *testing.T) {
	baseErr := errors.New("database connection failed")
	anotherErr := errors.New("another error")

	tests := []struct {
		name        string
		err         *NilDatabaseError
		expectedErr string
		shouldWrap  bool
		wrappedErr  error
	}{
		{
			name: "default error without base error",
			err: &NilDatabaseError{
				Message: "",
				BaseErr: nil,
			},
			expectedErr: "Mongo Database error. Making operation on nil Database",
			shouldWrap:  false,
			wrappedErr:  nil,
		},
		{
			name: "custom error message",
			err: &NilDatabaseError{
				Message: "Database is not initialized",
				BaseErr: nil,
			},
			expectedErr: "Database is not initialized",
			shouldWrap:  false,
			wrappedErr:  nil,
		},
		{
			name: "error with base error",
			err: &NilDatabaseError{
				Message: "Operation failed",
				BaseErr: baseErr,
			},
			expectedErr: "Operation failed. Base error: database connection failed",
			shouldWrap:  true,
			wrappedErr:  baseErr,
		},
		{
			name: "empty message with base error",
			err: &NilDatabaseError{
				Message: "",
				BaseErr: anotherErr,
			},
			expectedErr: "Mongo Database error. Making operation on nil Database. Base error: another error",
			shouldWrap:  true,
			wrappedErr:  anotherErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Тестируем Error() метод
			errorString := tt.err.Error()
			if errorString != tt.expectedErr {
				t.Errorf("Error() = %q, want %q", errorString, tt.expectedErr)
			}

			// Тестируем Unwrap() метод
			unwrapped := tt.err.Unwrap()
			if tt.shouldWrap {
				if !errors.Is(tt.err, tt.wrappedErr) {
					t.Errorf("errors.Is should return true for wrapped error")
				}
			} else {
				if unwrapped != nil {
					t.Errorf("Unwrap() = %v, want nil", unwrapped)
				}
			}
		})
	}
}

func TestErrorInterfaces(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantType string
	}{
		{
			name:     "NilClientError implements error interface",
			err:      &NilClientError{Message: "test"},
			wantType: "*NilClientError",
		},
		{
			name:     "NilDatabaseError implements error interface",
			err:      &NilDatabaseError{Message: "test"},
			wantType: "*NilDatabaseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error should not be nil")
			}

			// Проверяем, что ошибка может быть использована как error
			_ = tt.err.Error()

			// Проверяем тип
			switch tt.err.(type) {
			case *NilClientError:
				if tt.wantType != "*NilClientError" {
					t.Errorf("got type *NilClientError, want %s", tt.wantType)
				}
			case *NilDatabaseError:
				if tt.wantType != "*NilDatabaseError" {
					t.Errorf("got type *NilDatabaseError, want %s", tt.wantType)
				}
			default:
				t.Errorf("unexpected error type: %T", tt.err)
			}
		})
	}
}

func TestErrorsIsChain(t *testing.T) {
	baseErr := errors.New("original error")
	wrappedErr := &NilClientError{
		Message: "client error",
		BaseErr: baseErr,
	}

	doubleWrappedErr := &NilDatabaseError{
		Message: "database error",
		BaseErr: wrappedErr,
	}

	tests := []struct {
		name        string
		err         error
		target      error
		shouldMatch bool
	}{
		{
			name:        "direct match",
			err:         baseErr,
			target:      baseErr,
			shouldMatch: true,
		},
		{
			name:        "wrapped once",
			err:         wrappedErr,
			target:      baseErr,
			shouldMatch: true,
		},
		{
			name:        "wrapped twice",
			err:         doubleWrappedErr,
			target:      baseErr,
			shouldMatch: true,
		},
		{
			name:        "wrapped twice to middle error",
			err:         doubleWrappedErr,
			target:      wrappedErr,
			shouldMatch: true,
		},
		{
			name:        "no match for different error",
			err:         wrappedErr,
			target:      errors.New("different error"),
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := errors.Is(tt.err, tt.target)
			if matches != tt.shouldMatch {
				t.Errorf("errors.Is(%v, %v) = %v, want %v",
					tt.err, tt.target, matches, tt.shouldMatch)
			}
		})
	}
}
