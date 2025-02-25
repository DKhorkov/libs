package logging

// Config is a logging config, on base of which logger instance is created.
type Config struct {
	Level       Level
	LogFilePath string
}
