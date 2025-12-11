package mongodb

// Config is a database config, on base of which new Connector is created.
type Config struct {
	Host string
	Port int
}
