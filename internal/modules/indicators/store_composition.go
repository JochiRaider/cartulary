package indicators

import (
	"context"
	"fmt"
	"reflect"
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
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidentLifecycleAccess
	recordStore    RecordEnvelopePort
	revisionsStore revisionAppendPort
	projections    indicatorprojection.Rows
	sourceText     SourceTextPort
	now            func() time.Time
}

type RecordEnvelopePort interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	LoadEnvelopesTx(context.Context, pgx.Tx, []uuid.UUID, bool) (map[uuid.UUID]records.Envelope, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
}

type incidentLifecycleAccess interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type StoreDependencies struct {
	Postgres        postgres.DB
	Revisions       *revisions.Appender
	RecordEnvelopes RecordEnvelopePort
	Projections     indicatorprojection.Rows
	SourceText      SourceTextPort
	Clock           func() time.Time
}

func NewStore(dependencies StoreDependencies) (*Store, error) {
	checks := []struct {
		name  string
		value any
	}{
		{name: "Postgres", value: dependencies.Postgres},
		{name: "Revisions", value: dependencies.Revisions},
		{name: "RecordEnvelopes", value: dependencies.RecordEnvelopes},
		{name: "Projections", value: dependencies.Projections},
		{name: "SourceText", value: dependencies.SourceText},
		{name: "Clock", value: dependencies.Clock},
	}
	for _, check := range checks {
		if nilStoreDependency(check.value) {
			return nil, fmt.Errorf("compose Indicators store: %s is required", check.name)
		}
	}
	return &Store{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: admission.NewChecker(dependencies.Postgres),
		recordStore:    dependencies.RecordEnvelopes,
		revisionsStore: dependencies.Revisions,
		projections:    dependencies.Projections,
		sourceText:     dependencies.SourceText,
		now:            dependencies.Clock,
	}, nil
}

func nilStoreDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
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
