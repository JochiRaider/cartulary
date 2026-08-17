package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	projectionstorage "github.com/JochiRaider/cartulary/internal/modules/projections/internal/storage"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool     postgres.DB
	registry *providerRegistry
	physical *projectionstorage.Store
}

type TimelineSource interface {
	BuildProjectionMutationTx(context.Context, pgx.Tx, uuid.UUID) (timelineprojection.ProjectionMutation, error)
	ListProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (timelineprojection.ProjectionInputPage, error)
}

func NewStore(
	pool postgres.DB,
	catalog *Catalog,
	physical *projectionstorage.Store,
) (*Store, error) {
	if pool == nil {
		return nil, errors.New("projection database is required")
	}
	if catalog == nil || catalog.registry == nil {
		return nil, errors.New("projection catalog is required")
	}
	if physical == nil {
		return nil, errors.New("projection physical storage is required")
	}
	return &Store{pool: pool, registry: catalog.registry, physical: physical}, nil
}

func (s *Store) UpsertTimelineRowTx(ctx context.Context, tx pgx.Tx, input timelineprojection.ProjectionInput) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	return s.physical.UpsertTimelineRowTx(ctx, tx, input)
}

func (s *Store) ApplyTimelineMutationTx(ctx context.Context, tx pgx.Tx, mutation timelineprojection.ProjectionMutation) error {
	if err := mutation.Validate(); err != nil {
		return err
	}
	switch mutation.Kind {
	case timelineprojection.ProjectionMutationUpsert:
		return s.UpsertTimelineRowTx(ctx, tx, mutation.Input)
	case timelineprojection.ProjectionMutationDelete:
		if s == nil || s.physical == nil {
			return errors.New("projection storage is required")
		}
		return s.physical.DeleteTimelineRowTx(ctx, tx, mutation.RecordID)
	default:
		return fmt.Errorf("unsupported timeline projection mutation kind %q", mutation.Kind)
	}
}

func (s *Store) ApplyTimelineFixtureBatchTx(ctx context.Context, tx pgx.Tx, inputs []timelineprojection.ProjectionInput) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	for index, input := range inputs {
		mutation := timelineprojection.ProjectionMutation{
			Kind: timelineprojection.ProjectionMutationUpsert, RecordID: input.RecordID, Input: input,
		}
		if err := mutation.Validate(); err != nil {
			return fmt.Errorf("validate Timeline fixture projection input %d: %w", index+1, err)
		}
	}
	return s.physical.InsertTimelineFixtureBatchTx(ctx, tx, inputs)
}

func (s *Store) CountTimelineFixtureRows(ctx context.Context, incidentID uuid.UUID) (int, error) {
	if s == nil || s.physical == nil {
		return 0, errors.New("projection storage is required")
	}
	return s.physical.CountTimelineFixtureRows(ctx, incidentID)
}

func (s *Store) CountTimelineFixtureRowsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (int, error) {
	if s == nil || s.physical == nil {
		return 0, errors.New("projection storage is required")
	}
	return s.physical.CountTimelineFixtureRowsTx(ctx, tx, incidentID)
}

func (s *Store) refreshTimelineTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, source TimelineSource) error {
	if s == nil || source == nil {
		return errors.New("timeline projection source is required")
	}
	mutation, err := source.BuildProjectionMutationTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	return s.ApplyTimelineMutationTx(ctx, tx, mutation)
}

func (s *Store) RebuildTimeline(ctx context.Context, incidentID uuid.UUID) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, timelineViewSchemaID)
	defer func() { finishTelemetry(err) }()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin timeline projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentTimelineTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit timeline projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentTimelineTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, timelineViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentTimelineTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, source TimelineSource) error {
	if s == nil || source == nil {
		return errors.New("timeline projection source is required")
	}
	if s.physical == nil {
		return errors.New("projection storage is required")
	}
	if err := s.physical.DeleteTimelineIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.ApplyTimelineMutationTx(ctx, tx, timelineprojection.ProjectionMutation{
				Kind:     timelineprojection.ProjectionMutationUpsert,
				RecordID: input.RecordID,
				Input:    input,
			}); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			break
		}
		afterRecordID = page.NextRecordID
	}
	return nil
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}
