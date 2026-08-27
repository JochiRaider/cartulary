package fixturewriter

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	projectionstorage "github.com/JochiRaider/cartulary/internal/modules/projections/internal/storage"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// NewIndicatorRows creates an explicit test-only fixture writer. It uses the
// canonical source, query plan, and Projections-owned physical storage without
// constructing a parallel runtime graph.
func NewIndicatorFixtureWriter(
	t testing.TB,
	db postgres.DB,
	source indicatorprojection.SourceReader,
) indicatorprojection.Rows {
	t.Helper()
	if source == nil {
		t.Fatal("Indicator projection fixture source is required")
	}
	storage, err := projectionstorage.New(db)
	if err != nil {
		t.Fatalf("construct Indicator projection fixture storage: %v", err)
	}
	return &indicatorRows{source: source, storage: storage}
}

type indicatorRows struct {
	source  indicatorprojection.SourceReader
	storage *projectionstorage.Store
}

func (fixture *indicatorRows) RefreshIndicatorTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	input, found, err := fixture.source.LoadProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if !found {
		return fixture.storage.DeleteIndicatorRowTx(ctx, tx, recordID)
	}
	return fixture.storage.UpsertIndicatorTx(ctx, tx, input)
}

func (fixture *indicatorRows) LoadIndicatorTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	plan := queryengine.IndicatorPlans()[0]
	var query strings.Builder
	query.WriteString("SELECT ")
	query.WriteString(plan.RecordExpr)
	query.WriteString(", r.row_version")
	for _, field := range plan.Fields {
		query.WriteString(", ")
		query.WriteString(field.Expr)
	}
	query.WriteString(" ")
	query.WriteString(plan.FromSQL)
	query.WriteString(" WHERE ")
	query.WriteString(plan.RecordExpr)
	query.WriteString(" = $1 AND r.deleted_at IS NULL")
	values := make([]any, len(plan.Fields)+2)
	destinations := make([]any, len(values))
	for index := range values {
		destinations[index] = &values[index]
	}
	if err := tx.QueryRow(ctx, query.String(), recordID).Scan(destinations...); err != nil {
		return nil, err
	}
	return queryengine.BuildRow(plan, values)
}
