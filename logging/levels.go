package logging

import "log/slog"

// LogLevels are a simple abstractions on slog.Level.
var LogLevels = struct {
	INFO, DEBUG, WARN, ERROR slog.Level
}{
	INFO:  slog.LevelInfo,
	DEBUG: slog.LevelDebug,
	WARN:  slog.LevelWarn,
	ERROR: slog.LevelError,
}
