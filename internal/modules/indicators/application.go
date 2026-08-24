package indicators

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Application struct {
	pool            postgres.DB
	idempotency     IdempotencyPort
	incidentState   IncidentStatePort
	recordEnvelopes RecordEnvelopePort
	revisions       RevisionPort
	projections     indicatorprojection.Rows
	sourceText      SourceTextPort
	now             func() time.Time
}

type IdempotencyPort interface {
	GetRouteIdempotency(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error)
	InsertRouteIdempotencyPayload(context.Context, pgx.Tx, authn.RouteIdempotencyKey, []byte, int, any) error
}

type RecordEnvelopePort interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	LoadEnvelopesTx(context.Context, pgx.Tx, []uuid.UUID, bool) (map[uuid.UUID]records.Envelope, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
}

type IncidentStatePort interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type ApplicationDependencies struct {
	Postgres        postgres.DB
	Idempotency     IdempotencyPort
	IncidentState   IncidentStatePort
	Revisions       RevisionPort
	RecordEnvelopes RecordEnvelopePort
	Projections     indicatorprojection.Rows
	SourceText      SourceTextPort
	Clock           func() time.Time
}

func NewApplication(dependencies ApplicationDependencies) (*Application, error) {
	checks := []struct {
		name  string
		value any
	}{
		{name: "Postgres", value: dependencies.Postgres},
		{name: "Idempotency", value: dependencies.Idempotency},
		{name: "IncidentState", value: dependencies.IncidentState},
		{name: "Revisions", value: dependencies.Revisions},
		{name: "RecordEnvelopes", value: dependencies.RecordEnvelopes},
		{name: "Projections", value: dependencies.Projections},
		{name: "SourceText", value: dependencies.SourceText},
		{name: "Clock", value: dependencies.Clock},
	}
	for _, check := range checks {
		if nilApplicationDependency(check.value) {
			return nil, fmt.Errorf("compose Indicators application: %s is required", check.name)
		}
	}
	return &Application{
		pool:            dependencies.Postgres,
		idempotency:     dependencies.Idempotency,
		incidentState:   dependencies.IncidentState,
		recordEnvelopes: dependencies.RecordEnvelopes,
		revisions:       dependencies.Revisions,
		projections:     dependencies.Projections,
		sourceText:      dependencies.SourceText,
		now:             dependencies.Clock,
	}, nil
}

func nilApplicationDependency(dependency any) bool {
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

func (s *Application) refreshProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return s.projections.RefreshIndicatorTx(ctx, tx, recordID)
}

func (s *Application) loadProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return s.projections.LoadIndicatorTx(ctx, tx, recordID)
}

func (s *Application) refreshAndLoadProjectionRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	if err := s.refreshProjectionRowTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return s.loadProjectionRowTx(ctx, tx, recordID)
}
