package referenceassembly

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/jackc/pgx/v5"
)

// RegistryTransactions is the application composition for multi-owner record
// mutations. Its only capability is opening a transaction with registry guards
// already held, before any participant acquires incident or record locks.
// Ordinary readers and transactions that cannot assign registry types keep
// their original database port.
type RegistryTransactions struct {
	database interface {
		BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	}
	assignments reference_data.RegistryAssignments
}

func NewRegistryTransactions(database interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, assignments reference_data.RegistryAssignments) (*RegistryTransactions, error) {
	if database == nil || assignments == nil {
		return nil, errors.New("reference assembly: registry mutation dependencies are required")
	}
	return &RegistryTransactions{database: database, assignments: assignments}, nil
}

func (r *RegistryTransactions) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := r.database.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	if _, err := r.assignments.BeginTx(ctx, tx); err != nil {
		_ = tx.Rollback(context.WithoutCancel(ctx))
		return nil, err
	}
	return tx, nil
}
