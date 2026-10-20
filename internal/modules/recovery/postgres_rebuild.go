package recovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ErrUnsupportedPostgresRebuild rejects algorithms outside the existing
// transaction-local Recovery rebuild boundary before issuing database work.
var ErrUnsupportedPostgresRebuild = errors.New("unsupported PostgreSQL derived-state rebuild")

// RebuildPostgresDerivedState executes the catalog-owned algorithm using the
// caller's Recovery transaction. It neither commits nor acquires privileges.
func RebuildPostgresDerivedState(ctx context.Context, tx pgx.Tx, algorithmID string) error {
	rebuild, err := PostgresDerivedRebuild(algorithmID)
	if err != nil {
		return err
	}
	return rebuild(ctx, tx)
}

// PostgresDerivedRebuild resolves a supported algorithm before a caller mutates
// source rows. SQL and validation remain owned by this same Recovery boundary.
func PostgresDerivedRebuild(algorithmID string) (func(context.Context, pgx.Tx) error, error) {
	var rebuildSQL, validateSQL string
	switch algorithmID {
	case "entities.restore_active_identifier_claims.v1":
		rebuildSQL = `SELECT public.entities_rebuild_active_identifier_claims_v1()`
		validateSQL = `SELECT public.entities_active_identifier_claims_are_valid_v1()`
	case "parties.restore_active_key_claims.v1":
		rebuildSQL = `SELECT public.parties_rebuild_active_key_claims_v1()`
		validateSQL = `SELECT public.parties_active_key_claims_are_valid_v1()`
	default:
		return nil, ErrUnsupportedPostgresRebuild
	}
	return func(ctx context.Context, tx pgx.Tx) error {
		if tx == nil {
			return errors.New("recovery rebuild transaction is unavailable")
		}
		var rebuilt int64
		if err := tx.QueryRow(ctx, rebuildSQL).Scan(&rebuilt); err != nil {
			return fmt.Errorf("rebuild PostgreSQL derived state: %w", err)
		}
		var valid bool
		if err := tx.QueryRow(ctx, validateSQL).Scan(&valid); err != nil {
			return fmt.Errorf("validate PostgreSQL derived state: %w", err)
		}
		if !valid {
			return fmt.Errorf("%w: rebuilt PostgreSQL derived state is invalid", ErrVNextBackup)
		}
		return nil
	}, nil
}
