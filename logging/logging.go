package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/DKhorkov/libs/requestid"
)

const (
	skipLevel = 2
)

var (
	instance *slog.Logger
	once     sync.Once
)

// GetInstance implemented as singleton pattern to get Logger instance, created once for whole app:.
func GetInstance(logLevel slog.Level, logFilePath string) *slog.Logger {
	var logWriter io.Writer

	if logFile, err := os.OpenFile(
		logFilePath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666,
	); err != nil {
		fmt.Printf("Failed to open log file %s: %s\n", logFilePath, err)
		logWriter = os.Stdout
	} else {
		logWriter = io.MultiWriter(os.Stdout, logFile)
	}

	once.Do(func() {
		instance = slog.New(
			slog.NewJSONHandler(
				logWriter,
				&slog.HandlerOptions{
					Level: logLevel,
				},
			),
		)
	})

	return instance
}

// GetLogTraceback return a string with info about filename, function name and line
// https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
func GetLogTraceback(skipLevel int) string {
	pc, file, line, ok := runtime.Caller(skipLevel)
	if !ok {
		return fmt.Sprintf("%s on line %d: %s", "Unknown", 0, "Unknown")
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return fmt.Sprintf("%s on line %d: %s", file, line, "Unknown")
	}

	return fmt.Sprintf("%s on line %d: %s", file, line, fn.Name())
}

// LogRequest uses provided logger to save request info and connect it with request ID from the context.
func LogRequest(ctx context.Context, logger *slog.Logger, request any) {
	requestID, err := contextlib.GetValue[string](ctx, requestid.Key)
	if err != nil {
		requestID = ""
	}

	logger.InfoContext(
		ctx,
		"Received new request",
		"Request ID",
		requestID,
		"Request",
		request,
		"Traceback",
		GetLogTraceback(skipLevel), // 2 because 1 - is this func and 2 - it's caller
	)
}

// LogErrorContext uses provided logger to save error with message info and context.
// Context is used to get request ID and connect it with error.
func LogErrorContext(ctx context.Context, logger *slog.Logger, msg string, err error) {
	requestID, contextErr := contextlib.GetValue[string](ctx, requestid.Key)
	if contextErr != nil {
		requestID = ""
	}

	logger.ErrorContext(
		ctx,
		msg,
		"Request ID",
		requestID,
		"Traceback",
		GetLogTraceback(skipLevel),
		"Error",
		err,
	)
}

// LogError logs error with message info, using provided logger.
func LogError(logger *slog.Logger, msg string, err error) {
	logger.Error(
		msg,
		"Traceback",
		GetLogTraceback(skipLevel),
		"Error",
		err,
	)
}

// LogInfo logs message, using provided logger.
func LogInfo(logger *slog.Logger, msg string) {
	logger.Error(
		msg,
		"Traceback",
		GetLogTraceback(skipLevel),
	)
}
