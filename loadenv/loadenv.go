package loadenv

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

/*
init is invoked before main()

https://pkg.go.dev/github.com/joho/godotenv#section-readme
https://habr.com/ru/articles/446468/
*/
func init() {
	// loads values from .env into the system.
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

// GetEnv is a helper function to read an environment or return a default value.
func GetEnv(key string, defaultVal string) string {
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

// GetEnvAsSlice is a helper to read an environment variable into a string slice or return default value.
func GetEnvAsSlice(name string, defaultVal []string, separator string) []string {
	valStr := GetEnv(name, "")
	if valStr == "" {
		return defaultVal
	}

	if !IsStringIsValidSlice(valStr, separator) {
		return defaultVal
	}

	return strings.Split(valStr, separator)
}

// IsStringIsValidSlice is a helper to check, if string is a valid slice to convert it into []string.
// Can also be used outside loadenv package for other cases.
func IsStringIsValidSlice(str string, separator string) bool {
	str = strings.TrimSpace(str)
	slice := strings.Split(str, separator)
	for i := range slice {
		if slice[i] == "" {
			slice = append(slice[:i], slice[i+1:]...)
		}
	}

	return len(slice) > 1
}
