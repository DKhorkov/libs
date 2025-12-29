package logging_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/requestid"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	// Создаём временный файл для логов
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	logger := logging.New(logging.Levels.INFO, logFilePath)
	require.NotNil(t, logger)

	// Проверяем, что логгер — это *slog.Logger
	_, ok := logger.(*slog.Logger)
	require.True(t, ok)

	// Проверяем singleton: повторный вызов должен вернуть тот же экземпляр
	logger2 := logging.New(logging.Levels.DEBUG, logFilePath)
	require.Equal(t, logger, logger2)

	// Проверяем, что файл создан
	_, err := os.Stat(logFilePath)
	require.NoError(t, err)
}

func TestNewInvalidLogFile(t *testing.T) {
	t.Parallel()

	// Путь к невалидной директории
	logFilePath := "/invalid/path/test.log"

	logger := logging.New(logging.Levels.INFO, logFilePath)
	require.NotNil(t, logger)

	// Проверяем, что логгер всё равно создан (используется stdout)
	_, ok := logger.(*slog.Logger)
	require.True(t, ok)
}

func TestGetLogTraceback(t *testing.T) {
	t.Parallel()

	traceback := logging.GetLogTraceback(1)
	require.NotEmpty(t, traceback)
	require.Contains(t, traceback, "logging_test.TestGetLogTraceback")
	require.Contains(t, traceback, ".go")
	require.Contains(t, traceback, "line")
}

func TestGetLogTracebackInvalidSkipLevel(t *testing.T) {
	t.Parallel()

	traceback := logging.GetLogTraceback(1000)
	require.Contains(t, traceback, "Unknown")
	require.Contains(t, traceback, "line 0")
}

func TestLogErrorContext(t *testing.T) {
	t.Parallel()

	// Настраиваем мок для contextlib.ValueFromContext
	ctx := contextlib.WithValue(context.Background(), requestid.Key, "test-request-id")

	// Создаём буфер для захвата логов
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	err := errors.New("test error")
	logging.LogErrorContext(ctx, logger, "test error message", err)

	// Проверяем содержимое логов
	logOutput := buf.String()
	require.Contains(t, logOutput, `"level":"ERROR"`)
	require.Contains(t, logOutput, `"msg":"test error message"`)
	require.Contains(t, logOutput, `"Request ID":"test-request-id"`)
	require.Contains(t, logOutput, `"Traceback"`)
	require.Contains(t, logOutput, `"Error":"test error"`)
}

func TestLogErrorContextNoRequestID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Создаём буфер для захвата логов
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	err := errors.New("test error")
	logging.LogErrorContext(ctx, logger, "test error message", err)

	// Проверяем содержимое логов
	logOutput := buf.String()
	require.Contains(t, logOutput, `"level":"ERROR"`)
	require.Contains(t, logOutput, `"msg":"test error message"`)
	require.Contains(t, logOutput, `"Request ID":""`)
	require.Contains(t, logOutput, `"Traceback"`)
	require.Contains(t, logOutput, `"Error":"test error"`)
}

func TestLogInfoContext(t *testing.T) {
	t.Parallel()

	// Настраиваем мок для contextlib.ValueFromContext
	ctx := contextlib.WithValue(context.Background(), requestid.Key, "test-request-id")

	// Создаём буфер для захвата логов
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logging.LogInfoContext(ctx, logger, "test info message")

	// Проверяем содержимое логов
	logOutput := buf.String()
	require.Contains(t, logOutput, `"level":"INFO"`)
	require.Contains(t, logOutput, `"msg":"test info message"`)
	require.Contains(t, logOutput, `"Request ID":"test-request-id"`)
	require.Contains(t, logOutput, `"Traceback"`)
}

func TestLogInfoContextNoRequestID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Создаём буфер для захвата логов
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logging.LogInfoContext(ctx, logger, "test info message")

	// Проверяем содержимое логов
	logOutput := buf.String()
	require.Contains(t, logOutput, `"level":"INFO"`)
	require.Contains(t, logOutput, `"msg":"test info message"`)
	require.Contains(t, logOutput, `"Request ID":""`)
	require.Contains(t, logOutput, `"Traceback"`)
}

func TestLogError(t *testing.T) {
	t.Parallel()

	// Создаём буфер для захвата логов
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	err := errors.New("test error")
	logging.LogError(logger, "test error message", err)

	// Проверяем содержимое логов
	logOutput := buf.String()
	require.Contains(t, logOutput, `"level":"ERROR"`)
	require.Contains(t, logOutput, `"msg":"test error message"`)
	require.Contains(t, logOutput, `"Traceback"`)
	require.Contains(t, logOutput, `"Error":"test error"`)
}

func TestLogInfo(t *testing.T) {
	t.Parallel()

	// Создаём буфер для захвата логов
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logging.LogInfo(logger, "test info message")

	// Проверяем содержимое логов
	logOutput := buf.String()
	require.Contains(t, logOutput, `"level":"INFO"`)
	require.Contains(t, logOutput, `"msg":"test info message"`)
	require.Contains(t, logOutput, `"Traceback"`)
}
