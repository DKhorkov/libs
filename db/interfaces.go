package db

import (
	"context"
	"database/sql"
)

// Connector interface is created for usage in external application according to
// "dependency inversion principle" of SOLID due to working via abstractions.
type Connector interface {
	Close() error
	Transaction(ctx context.Context, opts ...TransactionOption) (*sql.Tx, error)
	Connection(ctx context.Context) (*sql.Conn, error)
}
