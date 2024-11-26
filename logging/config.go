package logging

import "log/slog"

// Config is a logging config, on base of which logger instance is created.
type Config struct {
	Level       slog.Level
	LogFilePath string
}
