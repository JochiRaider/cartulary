package projections

import (
	"context"
	"fmt"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/projectionprovider"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) RefreshArtifactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return s.refreshProjectionRowTx(ctx, tx, notesViewSchemaID, recordID)
}

func (s *Store) refreshArtifactTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return artifactprojection.RefreshArtifactTx(ctx, tx, recordID)
}

func (s *Store) RebuildIncidentArtifacts(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, notesViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentArtifactsTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentArtifactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, notesViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentArtifactsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return artifactprojection.RebuildIncidentArtifactsTx(ctx, tx, incidentID)
}

func (s *Store) RefreshEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return s.refreshProjectionRowTx(ctx, tx, evidenceViewSchemaID, recordID)
}

func (s *Store) refreshEvidenceTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return evidenceprojection.RefreshEvidenceTx(ctx, tx, recordID)
}

func (s *Store) RebuildIncidentEvidence(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, evidenceViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentEvidenceTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentEvidenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, evidenceViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentEvidenceTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return evidenceprojection.RebuildIncidentEvidenceTx(ctx, tx, incidentID)
}

func (s *Store) RefreshPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return s.refreshProjectionRowTx(ctx, tx, partiesViewSchemaID, recordID)
}

func (s *Store) refreshPartyTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return partyprojection.RefreshPartyTx(ctx, tx, recordID)
}

func (s *Store) RebuildIncidentParties(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, partiesViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentPartiesTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentPartiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, partiesViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentPartiesTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return partyprojection.RebuildIncidentPartiesTx(ctx, tx, incidentID)
}

func (s *Store) rebuildIncidentHotProjection(ctx context.Context, viewSchemaID string, rebuild func(context.Context, pgx.Tx) error) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, viewSchemaID)
	defer func() { finishTelemetry(err) }()

	if s == nil || s.pool == nil {
		return fmt.Errorf("rebuild hot projection: projection store is required")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin hot projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := rebuild(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit hot projection rebuild: %w", err)
	}
	return nil
}
