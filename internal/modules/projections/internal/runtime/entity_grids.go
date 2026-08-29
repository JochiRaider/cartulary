package runtime

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entitycontract "github.com/JochiRaider/cartulary/internal/modules/entities/projectioncontract"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
)

func (s *Store) refreshHostTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source entitycontract.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("host projection source is required")
	}
	input, found, err := source.LoadHostProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if !found {
		return s.physical.DeleteHostRowTx(ctx, tx, recordID)
	}
	return s.physical.UpsertHostTx(ctx, tx, input)
}

func (s *Store) rebuildIncidentHostsTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source entitycontract.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("host projection source is required")
	}
	if err := s.physical.DeleteHostIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListHostProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.UpsertHostTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
}

func (s *Store) refreshIdentityTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source entitycontract.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("identity projection source is required")
	}
	input, found, err := source.LoadIdentityProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if !found {
		return s.physical.DeleteIdentityRowTx(ctx, tx, recordID)
	}
	return s.physical.UpsertIdentityTx(ctx, tx, input)
}

func (s *Store) rebuildIncidentIdentitiesTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source entitycontract.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("identity projection source is required")
	}
	if err := s.physical.DeleteIdentityIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListIdentityProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.UpsertIdentityTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
}

func (s *Store) refreshIndicatorTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source indicatorprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("indicator projection source is required")
	}
	input, found, err := source.LoadProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if !found {
		return s.physical.DeleteIndicatorRowTx(ctx, tx, recordID)
	}
	return s.physical.UpsertIndicatorTx(ctx, tx, input)
}

func (s *Store) rebuildIncidentIndicatorsTxCore(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	source indicatorprojection.SourceReader,
) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("indicator projection source is required")
	}
	if err := s.physical.DeleteIndicatorIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.UpsertIndicatorTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
}
