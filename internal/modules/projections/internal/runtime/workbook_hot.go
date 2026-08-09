package runtime

import (
	"context"
	"errors"
	"fmt"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) refreshArtifactTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source artifactprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("artifact projection source is required")
	}
	input, found, err := source.LoadProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if err := s.physical.DeleteArtifactRowTx(ctx, tx, recordID); err != nil {
		return err
	}
	if !found {
		return nil
	}
	return s.physical.InsertArtifactTx(ctx, tx, input)
}

func (s *Store) RebuildIncidentArtifacts(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, notesViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentArtifactsTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentArtifactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, notesViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentArtifactsTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source artifactprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("artifact projection source is required")
	}
	if err := s.physical.DeleteArtifactIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.InsertArtifactTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
}

func (s *Store) refreshEvidenceTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source evidenceprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("evidence projection source is required")
	}
	input, found, err := source.LoadProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if err := s.physical.DeleteEvidenceRowTx(ctx, tx, recordID); err != nil {
		return err
	}
	if !found {
		return nil
	}
	return s.physical.InsertEvidenceTx(ctx, tx, input)
}

func (s *Store) RebuildIncidentEvidence(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, evidenceViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentEvidenceTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentEvidenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, evidenceViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentEvidenceTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source evidenceprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("evidence projection source is required")
	}
	if err := s.physical.DeleteEvidenceIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.InsertEvidenceTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
}

func (s *Store) refreshPartyTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source partyprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("party projection source is required")
	}
	input, found, err := source.LoadProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if err := s.physical.DeletePartyRowTx(ctx, tx, recordID); err != nil {
		return err
	}
	if !found {
		return nil
	}
	return s.physical.InsertPartyTx(ctx, tx, input)
}

func (s *Store) RebuildIncidentParties(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildIncidentHotProjection(ctx, partiesViewSchemaID, func(ctx context.Context, tx pgx.Tx) error {
		return s.RebuildIncidentPartiesTx(ctx, tx, incidentID)
	})
}

func (s *Store) RebuildIncidentPartiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, partiesViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentPartiesTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source partyprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("party projection source is required")
	}
	if err := s.physical.DeletePartyIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.InsertPartyTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
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
