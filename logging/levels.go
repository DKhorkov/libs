package logging

import "log/slog"

// Levels are a simple abstractions on slog.Level.
var Levels = struct {
	INFO, DEBUG, WARN, ERROR slog.Level
}{
	INFO:  slog.LevelInfo,
	DEBUG: slog.LevelDebug,
	WARN:  slog.LevelWarn,
	ERROR: slog.LevelError,
}
