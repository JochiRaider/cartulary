package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	commLogViewSchemaID              = "cartulary.view.comm_log.v1"
	findingsViewSchemaID             = "cartulary.view.findings.v1"
	forensicKeywordsViewSchemaID     = "cartulary.view.forensic_keywords.v1"
	handoffViewSchemaID              = "cartulary.view.handoff.v1"
	investigativeQueriesViewSchemaID = "cartulary.view.investigative_queries.v1"
	lessonViewSchemaID               = "cartulary.view.lesson.v1"
	statusReviewViewSchemaID         = "cartulary.view.status_review.v1"
)

func (s *Store) SupportsQuerySurface(viewSchemaID string) bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.querySurfaceForView(viewSchemaID)
	return ok
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	page, err := s.QueryRowsPage(ctx, incidentID, viewSchemaID, query, querypage.Window{Limit: int(^uint(0)>>1) - 1})
	return page.Rows, err
}

func (s *Store) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	definition, ok := s.providerRegistry().querySurfaceForView(viewSchemaID)
	if !ok {
		return querypage.Result{}, fmt.Errorf("projection query surface %q not mapped", viewSchemaID)
	}
	sqlText, args, err := buildGenericQueryPageSQL(incidentID, definition, query, window)
	if err != nil {
		return querypage.Result{}, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return querypage.Result{}, fmt.Errorf("query workbook rows: %w", err)
	}
	defer rows.Close()

	result, err := queryengine.ScanRows(rows, definition.queryEngineSurface())
	if err != nil {
		return querypage.Result{}, err
	}
	return querypage.Finish(result, window.Limit), nil
}

func (s *Store) LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	definition, ok := s.providerRegistry().querySurfaceForView(viewSchemaID)
	if !ok {
		return nil, fmt.Errorf("projection query surface %q not mapped", viewSchemaID)
	}
	return loadRowTx(ctx, tx, definition, recordID)
}

func loadRowTx(ctx context.Context, tx pgx.Tx, definition genericSurface, recordID uuid.UUID) (map[string]any, error) {
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(", r.row_version")
	for _, field := range definition.fields {
		builder.WriteString(", ")
		builder.WriteString(field.expr)
	}
	builder.WriteString(" ")
	builder.WriteString(definition.fromSQL)
	builder.WriteString(" WHERE ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(" = $1 AND r.deleted_at IS NULL")
	if definition.whereSQL != "" {
		builder.WriteString(" AND ")
		builder.WriteString(definition.whereSQL)
	}
	row := tx.QueryRow(ctx, builder.String(), recordID)
	values := make([]any, len(definition.fields)+2)
	scanTargets := make([]any, len(values))
	for index := range values {
		scanTargets[index] = &values[index]
	}
	if err := row.Scan(scanTargets...); err != nil {
		return nil, err
	}
	return buildGenericRow(definition, values)
}
