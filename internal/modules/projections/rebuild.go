package projections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Store) RebuildRestoreProjections(ctx context.Context) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, "unknown")
	defer func() { finishTelemetry(err) }()

	if s == nil || s.pool == nil {
		return fmt.Errorf("rebuild restore projections: projection store is required")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin restore projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	incidentIDs, err := listRestoreProjectionIncidentIDs(ctx, tx)
	if err != nil {
		return err
	}
	for _, incidentID := range incidentIDs {
		if err := s.RebuildIncidentTimelineTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild timeline projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentHostsTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild host projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentIdentitiesTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild identity projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentIndicatorsTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild indicator projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentAssessmentsTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild assessment projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentArtifactsTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild artifact projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentEvidenceTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild evidence projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentPartiesTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild party projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentTaskRequestsTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild task request projection for incident %s: %w", incidentID, err)
		}
		if err := s.RebuildIncidentDecisionsTx(ctx, tx, incidentID); err != nil {
			return fmt.Errorf("rebuild decision projection for incident %s: %w", incidentID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit restore projection rebuild: %w", err)
	}
	return nil
}

func listRestoreProjectionIncidentIDs(ctx context.Context, tx pgx.Tx) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `SELECT id FROM incidents ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list restore projection incidents: %w", err)
	}
	defer rows.Close()

	incidentIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var rawIncidentID pgtype.UUID
		if err := rows.Scan(&rawIncidentID); err != nil {
			return nil, fmt.Errorf("scan restore projection incident: %w", err)
		}
		incidentID, err := uuidFromPG(rawIncidentID)
		if err != nil {
			return nil, fmt.Errorf("decode restore projection incident: %w", err)
		}
		incidentIDs = append(incidentIDs, incidentID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate restore projection incidents: %w", err)
	}
	return incidentIDs, nil
}
