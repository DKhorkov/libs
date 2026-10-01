package loadenv

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Init подразумевает явный вызов и инициализацию переменных с указанным путем перед
// поулчением переменных окружения.
func Init(paths ...string) {
	// loads values from .env into the system.
	if err := godotenv.Load(paths...); err != nil {
		fmt.Println("No .env file found")
	}
}

// GetEnv is a helper function to read an environment or return a default value.
func GetEnv(key, defaultVal string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	return defaultVal
}

// GetEnvAsInt is a helper function to read an environment variable into integer or return a default value.
func GetEnvAsInt(name string, defaultVal int) int {
	valueStr := GetEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// GetEnvAsBool is a helper to read an environment variable into a bool or return default value.
func GetEnvAsBool(name string, defaultVal bool) bool {
	valStr := GetEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// GetEnvAsSlice is a helper function to read an environment variable into a slice
// or return a default value.
//
// Elements are trimmed and empty ones are dropped, so "a,b" and "a, b" with a ","
// separator give the same result. The default value is returned only when nothing
// is left: the variable is unset, empty, or built of separators alone.
//
// A single element is a valid slice. It used to be replaced by the default value
// without a word, and that is the case every "one origin", "one host", "one
// allowed method" setting falls into — that is, the one people actually write.
func GetEnvAsSlice(name string, defaultVal []string, separator string) []string {
	values := sliceFromString(GetEnv(name, ""), separator)
	if len(values) == 0 {
		return defaultVal
	}

	return values
}

// IsStringIsValidSlice is a helper to check, if string is a valid slice to convert it into []string.
// Can also be used outside loadenv package for other cases.
//
// One element is enough: a list of one is still a list. The former "more than one"
// rule made every single-valued setting silently fall back to its default.
func IsStringIsValidSlice(str, separator string) bool {
	return len(sliceFromString(str, separator)) > 0
}

// sliceFromString splits a string, trimming elements and dropping empty ones.
//
// One function for both exported ones on purpose: they used to answer the same
// question apart — one whether the string is a slice, the other what slice it is —
// and an answer given twice is an answer that drifts. Here the check is literally
// "is the result non-empty".
//
// Not exported: the package API stays what it was, and the two functions above
// are enough for every caller.
func sliceFromString(str, separator string) []string {
	if separator == "" {
		return nil
	}

	parts := strings.Split(str, separator)
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}

	return values
}
