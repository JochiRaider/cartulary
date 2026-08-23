package indicators

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool               postgres.DB
	authStore          *authn.Store
	incidentAccess     incidentLifecycleAccess
	recordStore        indicatorRecordStore
	revisionsStore     revisionAppendPort
	projections        indicatorprojection.Rows
	sourceText         SourceTextPort
	now                func() time.Time
	sources            sourceRepository
	observations       observationRepository
	lifecycles         lifecycleRepository
	createService      indicatorCreateService
	observationService indicatorObservationService
	lifecycleService   indicatorLifecycleService
}

type indicatorRecordStore interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	LoadEnvelopesTx(context.Context, pgx.Tx, []uuid.UUID, bool) (map[uuid.UUID]records.Envelope, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
}

type incidentLifecycleAccess interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type StoreDependencies struct {
	Postgres    postgres.DB
	Revisions   *revisions.Appender
	Projections indicatorprojection.Rows
	SourceText  SourceTextPort
	Clock       func() time.Time
}

func NewStore(dependencies StoreDependencies) (*Store, error) {
	if dependencies.Postgres == nil {
		return nil, fmt.Errorf("compose Indicators store: Postgres is required")
	}
	if dependencies.Revisions == nil {
		return nil, fmt.Errorf("compose Indicators store: Revisions is required")
	}
	if dependencies.Projections == nil {
		return nil, fmt.Errorf("compose Indicators store: Projections is required")
	}
	if dependencies.SourceText == nil {
		return nil, fmt.Errorf("compose Indicators store: SourceText is required")
	}
	now := dependencies.Clock
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	store := &Store{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: admission.NewChecker(dependencies.Postgres),
		recordStore:    records.NewStore(),
		revisionsStore: newRevisionAppendAdapter(dependencies.Revisions),
		projections:    dependencies.Projections,
		sourceText:     dependencies.SourceText,
		now:            now,
		sources:        sourceRepository{},
		observations:   observationRepository{},
		lifecycles:     lifecycleRepository{},
	}
	store.createService = indicatorCreateService{owner: store}
	store.observationService = indicatorObservationService{owner: store}
	store.lifecycleService = indicatorLifecycleService{owner: store}
	return store, nil
}

func (s *Store) refreshProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return s.projections.RefreshIndicatorTx(ctx, tx, recordID)
}

func (s *Store) loadProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return s.projections.LoadIndicatorTx(ctx, tx, recordID)
}

func (s *Store) refreshAndLoadProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	if err := s.refreshProjectionRowTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return s.loadProjectionRowTx(ctx, tx, recordID)
}
