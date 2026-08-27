package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RebuildIncident atomically rebuilds the complete active provider catalog.
// It is the sole non-restore maintenance entry point.
func (s *Store) RebuildIncident(ctx context.Context, incidentID uuid.UUID) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx)
	defer func() { finishTelemetry(err) }()

	if s == nil || s.pool == nil {
		return fmt.Errorf("rebuild incident projections: projection store is required")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin incident projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := s.RebuildIncidentTx(ctx, tx, incidentID); err != nil {
		return fmt.Errorf("rebuild incident projections: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit incident projection rebuild: %w", err)
	}
	return nil
}
