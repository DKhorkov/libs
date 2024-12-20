package db

import "database/sql"

// transactionOptions represents options for *sql.Tx configuration.
type transactionOptions struct {
	isolationLevel sql.IsolationLevel
	readOnly       bool
}

// TransactionOption represents golang functional option pattern func for transaction settings.
type TransactionOption func(options *transactionOptions) error

// WithTransactionIsolationLevel sets transaction isolation level for database transaction.
func WithTransactionIsolationLevel(isolationLevel sql.IsolationLevel) TransactionOption {
	return func(options *transactionOptions) error {
		options.isolationLevel = isolationLevel
		return nil
	}
}

// WithTransactionReadOnly sets readOnly attribute for database transaction.
func WithTransactionReadOnly(readOnly bool) TransactionOption {
	return func(options *transactionOptions) error {
		options.readOnly = readOnly
		return nil
	}
}
