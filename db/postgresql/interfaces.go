package postgresql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

// Connector represents abstraction to work with Database according dependency inversion principal relying on methods.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/connector.go -package=mocks -exclude_interfaces=Transaction,Pool,Connection
type Connector interface {
	Close() error
	Transaction(ctx context.Context, opts ...TransactionOption) (Transaction, error)
	Connection(ctx context.Context) (Connection, error)
	Pool() Pool
}

// Transaction represents abstraction of Database to comply Atomicity principle
// according dependency inversion principal relying on methods.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/transaction.go -package=mocks -exclude_interfaces=Connector,Pool,Connection
type Transaction interface {
	Commit() error
	Rollback() error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
	Stmt(stmt *sql.Stmt) *sql.Stmt
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Exec(query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRow(query string, args ...any) *sql.Row
}

// Connection represents abstraction of Database to execute any operation with Database
//
//go:generate mockgen -source=interfaces.go -destination=mocks/connection.go -package=mocks -exclude_interfaces=Connector,Transaction,Pool
type Connection interface {
	PingContext(ctx context.Context) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Raw(f func(driverConn any) error) (err error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

// Pool represents abstraction of Database to work with connections and transactions
//
//go:generate mockgen -source=interfaces.go -destination=mocks/pool.go -package=mocks -exclude_interfaces=Connector,Transaction,Connection
type Pool interface {
	PingContext(ctx context.Context) error
	Ping() error
	Close() error
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	SetConnMaxLifetime(d time.Duration)
	SetConnMaxIdleTime(d time.Duration)
	Stats() sql.DBStats
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Exec(query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRow(query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Begin() (*sql.Tx, error)
	Driver() driver.Driver
	Conn(ctx context.Context) (*sql.Conn, error)
}
