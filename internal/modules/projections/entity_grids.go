package projections

import (
	"context"
	"fmt"

	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/projectionprovider"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/projectionprovider"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (s *Store) refreshHostTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return entityprojection.RefreshHostTx(ctx, tx, recordID)
}

func (s *Store) rebuildIncidentHostsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return entityprojection.RebuildIncidentHostsTx(ctx, tx, incidentID)
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

func (s *Store) refreshIdentityTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return entityprojection.RefreshIdentityTx(ctx, tx, recordID)
}

func (s *Store) rebuildIncidentIdentitiesTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return entityprojection.RebuildIncidentIdentitiesTx(ctx, tx, incidentID)
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

func (s *Store) refreshIndicatorTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return indicatorprojection.RefreshIndicatorTx(ctx, tx, recordID)
}

func (s *Store) rebuildIncidentIndicatorsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return indicatorprojection.RebuildIncidentIndicatorsTx(ctx, tx, incidentID)
}
