package logging

import "log/slog"

type Level int

// Levels are a simple abstractions on slog.Level.
var Levels = struct {
	INFO, DEBUG, WARN, ERROR Level
}{
	INFO:  Level(slog.LevelInfo),
	DEBUG: Level(slog.LevelDebug),
	WARN:  Level(slog.LevelWarn),
	ERROR: Level(slog.LevelError),
}
