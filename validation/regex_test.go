package validation_test

import (
	"testing"

	"github.com/DKhorkov/libs/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidateValueByRule(t *testing.T) {
	t.Parallel()

	const commonRule = "^[a-z0-9._%+\\-]+@[a-z0-9.\\-]+\\.[a-z]{2,4}$"

	testCases := []struct {
		name     string
		email    string
		rule     string
		expected bool
	}{
		{
			name:     "email validation success empty rule",
			email:    "alexqwerty228@yandex.ru",
			rule:     "",
			expected: true,
		},
		{
			name:     "email validation success common rule",
			email:    "alexqwerty228@yandex.ru",
			rule:     commonRule,
			expected: true,
		},
		{
			name:     "email validation failure no domain",
			email:    "alexqwerty228@",
			rule:     commonRule,
			expected: false,
		},
		{
			name:     "email validation failure invalid domain end",
			email:    "alexqwerty228@gmail.commsdf",
			rule:     commonRule,
			expected: false,
		},
		{
			name:     "email validation failure no @",
			email:    "alexqwerty228gmail.com",
			rule:     commonRule,
			expected: false,
		},
		{
			name:     "email validation failure invalid symbol",
			email:    "alexqwerty#228@gmail.com",
			rule:     commonRule,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			valid := validation.ValidateValueByRule(tc.email, tc.rule)
			assert.Equal(t, tc.expected, valid)
		})
	}
}

func TestValidateValueByRules(t *testing.T) {
	t.Parallel()

	commonRules := []string{
		".{8,}",
		"[a-z]",
		"[A-Z]",
		"[0-9]",
		"[^\\d\\w]",
	}

	testCases := []struct {
		name     string
		password string
		rules    []string
		expected bool
	}{
		{
			name:     "password validation success empty rule",
			password: "SomePass",
			rules:    []string{},
			expected: true,
		},
		{
			name:     "password validation success common rules",
			password: "Some@StrongPass1",
			rules:    commonRules,
			expected: true,
		},
		{
			name:     "password validation failure too short",
			password: "@Strng1",
			rules:    commonRules,
			expected: false,
		},
		{
			name:     "password validation failure no special characters",
			password: "StrongPass1",
			rules:    commonRules,
			expected: false,
		},
		{
			name:     "password validation failure no numbers",
			password: "Strong@Pass",
			rules:    commonRules,
			expected: false,
		},
		{
			name:     "password validation failure no uppercase characters",
			password: "strong@pass1",
			rules:    commonRules,
			expected: false,
		},
		{
			name:     "password validation failure no lowercase characters",
			password: "STRONG@PASS1",
			rules:    commonRules,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			valid := validation.ValidateValueByRules(tc.password, tc.rules)
			assert.Equal(t, tc.expected, valid)
		})
	}
}
