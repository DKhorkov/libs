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
