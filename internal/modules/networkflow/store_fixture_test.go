package networkflow

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Fixture-only convenience operations wrap the actual transaction-bound owner.
func (s *store) CreateTable(ctx context.Context, params createTableParams) (tableRecord, error) {
	var table tableRecord
	err := withinTransaction(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.CreateTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return tableRecord{}, err
	}
	return table, nil
}

func (s *store) RenameTable(ctx context.Context, params renameTableParams) (tableRecord, error) {
	var table tableRecord
	err := withinTransaction(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.renameTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return tableRecord{}, err
	}
	return table, nil
}

func (s *store) SoftDeleteTable(ctx context.Context, params softDeleteTableParams) (tableRecord, error) {
	var table tableRecord
	err := withinTransaction(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		table, err = s.softDeleteTableTx(ctx, tx, params)
		return err
	})
	if err != nil {
		return tableRecord{}, err
	}
	return table, nil
}

func (s *store) RetainedCounts(ctx context.Context, incidentID uuid.UUID) (retainedCounts, error) {
	return retainedCountsTx(ctx, s.pool, incidentID)
}
