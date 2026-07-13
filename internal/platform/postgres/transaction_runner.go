package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TransactionRunner centralizes transaction lifecycle while application use
// cases retain ownership of operation order inside the callback.
type TransactionRunner interface {
	WithinTx(context.Context, pgx.TxOptions, func(pgx.Tx) error) error
}

type transactionRunner struct {
	db DB
}

func NewTransactionRunner(db DB) TransactionRunner {
	return &transactionRunner{db: db}
}

func (r *transactionRunner) WithinTx(ctx context.Context, options pgx.TxOptions, operation func(pgx.Tx) error) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("PostgreSQL transaction runner unavailable")
	}
	tx, err := r.db.BeginTx(ctx, options)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := operation(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
