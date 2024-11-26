package db

import "database/sql"

// Connector interface is created for usage in external application according to
// "dependency inversion principle" of SOLID due to working via abstractions.
type Connector interface {
	Connect() error
	CloseConnection()
	GetTransaction() (*sql.Tx, error)
	GetConnection() *sql.DB
}
