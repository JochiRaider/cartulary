package runtime

import (
	"context"
	"errors"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectioncontract"
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
		for _, input := range page.Inputs() {
			if err := s.physical.InsertArtifactTx(ctx, tx, input); err != nil {
				return err
			}
		}
		nextRecordID := page.NextRecordID()
		if nextRecordID == nil {
			return nil
		}
		afterRecordID = nextRecordID
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
