package projections

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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

func SupportsQuerySurface(viewSchemaID string) bool {
	_, ok := defaultProviderRegistry().querySurfaceForView(viewSchemaID)
	return ok
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	definition, ok := s.providerRegistry().querySurfaceForView(viewSchemaID)
	if !ok {
		return nil, fmt.Errorf("projection query surface %q not mapped", viewSchemaID)
	}
	sqlText, args, err := buildGenericQuerySQL(incidentID, definition, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query workbook rows: %w", err)
	}
	defer rows.Close()

	result := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("read workbook row values: %w", err)
		}
		if len(values) != len(definition.fields)+2 {
			return nil, fmt.Errorf("query workbook rows: unexpected value count %d", len(values))
		}
		row, err := buildGenericRow(definition, query.GroupBy, values)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workbook rows: %w", err)
	}
	return result, nil
}

func (s *Store) LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	definition, ok := s.providerRegistry().querySurfaceForView(viewSchemaID)
	if !ok {
		return nil, fmt.Errorf("projection query surface %q not mapped", viewSchemaID)
	}
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
	return buildGenericRow(definition, nil, values)
}
