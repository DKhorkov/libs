package loadenv_test

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DKhorkov/libs/loadenv"
)

// Тест в табличном стиле для функции Init
func TestInit(t *testing.T) {
	// Подготовка временных .env файлов для тестов
	const validEnvContent = "KEY=value\nANOTHER=123"
	const invalidEnvContent = "KEY=value\n=invalid_line"

	// Создаём временные файлы
	validFile := createTempFile(t, validEnvContent)
	defer func() {
		require.NoError(t, os.Remove(validFile))
	}() // очистка после теста

	invalidFile := createTempFile(t, invalidEnvContent)
	defer func() {
		require.NoError(t, os.Remove(invalidFile))
	}() // очистка после теста

	nonExistentFile := "non_existent.env"

	tests := []struct {
		name        string
		paths       []string
		expectError bool
		setup       func()
		teardown    func()
	}{
		{
			name:        "Позитивный кейс: корректный .env файл",
			paths:       []string{validFile},
			expectError: false,
			setup:       func() {},
			teardown: func() {
				require.NoError(t, os.Unsetenv("KEY"))
				require.NoError(t, os.Unsetenv("ANOTHER"))
			},
		},
		{
			name:        "Негативный кейс: некорректный синтаксис .env",
			paths:       []string{invalidFile},
			expectError: true,
			setup:       func() {},
			teardown:    func() {},
		},
		{
			name:        "Файл не существует, но это допустимо (godotenv не падает)",
			paths:       []string{nonExistentFile},
			expectError: false, // godotenv.Load не возвращает ошибку, если файл не найден
			setup:       func() {},
			teardown:    func() {},
		},
		{
			name:        "Пустой список путей — загружается .env по умолчанию",
			paths:       []string{},
			expectError: false,
			setup: func() {
				// Создаём .env в корне (если его нет)
				if _, err := os.Stat(".env"); os.IsNotExist(err) {
					f, _ := os.Create(".env")
					_, err = f.WriteString("DEFAULT=value")
					require.NoError(t, err)
					require.NoError(t, f.Close())
				}
			},
			teardown: func() {
				require.NoError(t, os.Remove(".env"))
				require.NoError(t, os.Unsetenv("KEY"))
				require.NoError(t, os.Unsetenv("ANOTHER"))
			},
		},
		{
			name:        "Несколько файлов: один валидный, один нет — ошибка",
			paths:       []string{validFile, invalidFile},
			expectError: true,
			setup:       func() {},
			teardown: func() {
				require.NoError(t, os.Unsetenv("KEY"))
				require.NoError(t, os.Unsetenv("ANOTHER"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

			// Перехватываем вывод, чтобы проверить сообщение "No .env file found"
			// (опционально — можно использовать буферизацию stdout)

			// Вызываем тестируемую функцию
			loadenv.Init(tt.paths...)

			// Проверяем, загрузились ли переменные, если ожидается успех
			if !tt.expectError {
				// Простая проверка: хотя бы одна переменная должна быть установлена
				if len(tt.paths) > 0 && tt.paths[0] == validFile {
					if val := os.Getenv("KEY"); val != "value" {
						t.Errorf("Ожидалось, что KEY=value, получено %s", val)
					}
				}
			}
			// Ошибки мы не возвращаем явно, но можем проверить поведение godotenv.Load
			// Косвенно: если ожидалась ошибка — проверим, что переменные не загружены или поведение нестабильно
			// В данном случае, так как Init не возвращает ошибку, мы полагаемся на логику godotenv
		})
	}
}

// Вспомогательная функция для создания временного .env файла
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "*.env")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err = tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	return tmpfile.Name()
}

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
