package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
)

func (s *Store) RebuildIncidentHosts(ctx context.Context, incidentID uuid.UUID) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, hostsViewSchemaID)
	defer func() { finishTelemetry(err) }()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin host projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentHostsTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit host projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, hostsViewSchemaID, incidentID)
}

func (s *Store) refreshHostTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source entityprojection.SourceReader,
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
	source entityprojection.SourceReader,
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

func (s *Store) RebuildIncidentIdentities(ctx context.Context, incidentID uuid.UUID) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, identitiesViewSchemaID)
	defer func() { finishTelemetry(err) }()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin identity projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentIdentitiesTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit identity projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, identitiesViewSchemaID, incidentID)
}

func (s *Store) refreshIdentityTxCore(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	source entityprojection.SourceReader,
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
	source entityprojection.SourceReader,
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

func (s *Store) RebuildIncidentIndicators(ctx context.Context, incidentID uuid.UUID) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, indicatorsViewSchemaID)
	defer func() { finishTelemetry(err) }()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin indicator projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentIndicatorsTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit indicator projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentIndicatorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, indicatorsViewSchemaID, incidentID)
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
