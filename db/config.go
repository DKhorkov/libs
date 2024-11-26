package db

// Config is a database config, on base of which new connector is created.
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSLMode      string
	Driver       string
}

type TestConfig struct {
	Driver        string
	DSN           string
	MigrationsDir string
}

func NewTestConfig() *TestConfig {
	return &TestConfig{
		Driver:        "sqlite3",
		DSN:           "file::memory:?cache=shared", // "test.db" can be also used
		MigrationsDir: "/internal/database/migrations",
	}
}
