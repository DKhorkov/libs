package logging

import (
	"context"
	"fmt"
	"github.com/DKhorkov/libs/contextlib"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"sync"
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

func LogRequest(logger *slog.Logger, ctx context.Context, request *http.Request) {
	requestID, err := contextlib.GetValueFromContext[string](ctx, "requestID")
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
		GetLogTraceback(2), // 2 because 1 - is this func and 2 - it's caller
	)
}

func LogErrorContext(logger *slog.Logger, ctx context.Context, msg string, err error) {
	requestID, err := contextlib.GetValueFromContext[string](ctx, "requestID")
	if err != nil {
		requestID = ""
	}

	logger.ErrorContext(
		ctx,
		"Request ID",
		requestID,
		msg,
		"Traceback",
		GetLogTraceback(2),
		"Error",
		err,
	)
}

func LogError(logger *slog.Logger, msg string, err error) {
	logger.Error(
		msg,
		"Traceback",
		GetLogTraceback(2),
		"Error",
		err,
	)
}
