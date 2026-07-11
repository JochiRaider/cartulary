package indicators

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var indicatorSortExpressions = map[string]string{
	"record_id":                       "i.record_id",
	"indicator.indicator_type":        "i.indicator_type",
	"indicator.value_kind":            "i.value_kind",
	"indicator.display_value":         "i.display_value",
	"indicator.normalized_value":      "i.normalized_value",
	"indicator.defanged_value":        "i.defanged_value",
	"indicator.hash_algorithm":        "i.hash_algorithm",
	"indicator.hash_value":            "i.hash_value",
	"indicator.stix_pattern":          "i.stix_pattern",
	"indicator.first_observed_at":     "i.first_observed_at",
	"indicator.last_observed_at":      "i.last_observed_at",
	"indicator.observation_count":     "i.observation_count",
	"indicator.lifecycle_summary":     "i.lifecycle_summary",
	"indicator.supporting_link_count": "i.supporting_link_count",
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	page, err := s.QueryRowsPage(ctx, incidentID, query, querypage.Window{Limit: int(^uint(0)>>1) - 1})
	return page.Rows, err
}

func (s *Store) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	if s.pool == nil {
		return querypage.Result{}, fmt.Errorf("query indicator rows: store pool is nil")
	}

	sqlText, args, err := buildIndicatorQueryPageSQL(incidentID, query, window)
	if err != nil {
		return querypage.Result{}, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return querypage.Result{}, fmt.Errorf("query indicator rows: %w", err)
	}
	defer rows.Close()

	result := make([]map[string]any, 0)
	for rows.Next() {
		record, err := scanIndicatorProjectionRecord(rows)
		if err != nil {
			return querypage.Result{}, err
		}
		result = append(result, BuildIndicatorRow(record))
	}
	if err := rows.Err(); err != nil {
		return querypage.Result{}, fmt.Errorf("iterate indicator rows: %w", err)
	}
	return querypage.Finish(result, window.Limit), nil
}

func buildIndicatorQueryPageSQL(incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    i.record_id::text,
    i.incident_id::text,
    r.row_version,
    i.indicator_type,
    i.value_kind,
    i.display_value,
    i.normalized_value,
    i.dedupe_key,
    i.defanged_value,
    i.hash_algorithm,
    i.hash_value,
    i.stix_pattern,
    i.first_observed_at,
    i.last_observed_at,
    i.observation_count,
    i.lifecycle_summary,
    i.supporting_link_count,
    i.edited_at
  FROM indicator_grid_projection i
  JOIN records r
    ON r.record_id = i.record_id
 WHERE i.incident_id = $1`)
	builder.WriteString(`
   AND r.deleted_at IS NULL`)
	args := []any{incidentID}

	for _, filter := range query.Filters {
		switch filter.FieldKey {
		case "indicator.indicator_type":
			if err := appendQueryTextClause(&builder, &args, "i.indicator_type", filter); err != nil {
				return "", nil, err
			}
		case "indicator.value_kind":
			if err := appendQueryTextClause(&builder, &args, "i.value_kind", filter); err != nil {
				return "", nil, err
			}
		case "indicator.hash_algorithm":
			if err := appendQueryTextClause(&builder, &args, "i.hash_algorithm", filter); err != nil {
				return "", nil, err
			}
		case "indicator.lifecycle_summary":
			if err := appendQueryTextClause(&builder, &args, "i.lifecycle_summary", filter); err != nil {
				return "", nil, err
			}
		case "indicator.first_observed_at":
			if err := appendQueryDateTimeClause(&builder, &args, "i.first_observed_at", filter); err != nil {
				return "", nil, err
			}
		case "indicator.last_observed_at":
			if err := appendQueryDateTimeClause(&builder, &args, "i.last_observed_at", filter); err != nil {
				return "", nil, err
			}
		default:
			return "", nil, fmt.Errorf("indicator query filter field %q not mapped", filter.FieldKey)
		}
	}

	pageFields := make(map[string]querypage.Field, len(indicatorSortExpressions))
	for key, expression := range indicatorSortExpressions {
		cast := ""
		switch {
		case key == "record_id":
			cast = "uuid"
		case strings.HasSuffix(key, "_count"):
			cast = "bigint"
		case strings.HasSuffix(key, "_at"):
			cast = "timestamptz"
		}
		pageFields[key] = querypage.Field{Expression: expression, Cast: cast}
	}
	if err := querypage.AppendKeyset(&builder, &args, query.Sort, pageFields, window.Position); err != nil {
		return "", nil, err
	}
	if err := appendOrderBy(&builder, query.Sort, indicatorSortExpressions); err != nil {
		return "", nil, err
	}
	if err := querypage.AppendLimit(&builder, &args, window.Limit); err != nil {
		return "", nil, err
	}
	return builder.String(), args, nil
}

func appendOrderBy(builder *strings.Builder, sort []viewschema.SortEntry, expressions map[string]string) error {
	builder.WriteString(" ORDER BY ")
	for index, sortEntry := range sort {
		if index > 0 {
			builder.WriteString(", ")
		}
		expr, ok := expressions[sortEntry.FieldKey]
		if !ok {
			return fmt.Errorf("query sort field %q not mapped", sortEntry.FieldKey)
		}
		builder.WriteString(expr)
		if sortEntry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
		builder.WriteString(" NULLS LAST")
	}
	return nil
}

func appendQueryTextClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendQueryCaseFoldedEqualityClause(builder, args, expr, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		builder.WriteString("\n   AND left(lower(coalesce((")
		builder.WriteString(expr)
		builder.WriteString(")::text, '')), char_length(")
		builder.WriteString(bindQueryValue(args, value, ""))
		builder.WriteString(")) = ")
		builder.WriteString(bindQueryValue(args, value, ""))
		return nil
	default:
		return fmt.Errorf("text filter operator %q not mapped", filter.Op)
	}
}

func appendQueryCaseFoldedEqualityClause(builder *strings.Builder, args *[]any, expr string, arg map[string]any) error {
	if value, ok := arg["value"]; ok {
		if value == nil {
			builder.WriteString("\n   AND ")
			builder.WriteString(expr)
			builder.WriteString(" IS NULL")
			return nil
		}
		builder.WriteString("\n   AND lower(coalesce((")
		builder.WriteString(expr)
		builder.WriteString(")::text, '')) = ")
		builder.WriteString(bindQueryValue(args, value, ""))
		return nil
	}
	values, ok := arg["values"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("missing equality values for %s", expr)
	}
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString("lower(coalesce((")
		builder.WriteString(expr)
		builder.WriteString(")::text, '')) = ")
		builder.WriteString(bindQueryValue(args, value, ""))
	}
	builder.WriteString(")")
	return nil
}

func appendQueryDateTimeClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendQueryEqualityClause(builder, args, expr, filter.Arg, "timestamptz")
	case "range":
		for _, bound := range []struct {
			Key string
			Op  string
		}{
			{Key: "gt", Op: ">"},
			{Key: "gte", Op: ">="},
			{Key: "lt", Op: "<"},
			{Key: "lte", Op: "<="},
		} {
			value, ok := filter.Arg[bound.Key]
			if !ok {
				continue
			}
			builder.WriteString("\n   AND ")
			builder.WriteString(expr)
			builder.WriteByte(' ')
			builder.WriteString(bound.Op)
			builder.WriteByte(' ')
			builder.WriteString(bindQueryValue(args, value, "timestamptz"))
		}
		return nil
	default:
		return fmt.Errorf("timestamp filter operator %q not mapped", filter.Op)
	}
}

func appendQueryEqualityClause(builder *strings.Builder, args *[]any, expr string, arg map[string]any, cast string) error {
	if value, ok := arg["value"]; ok {
		if value == nil {
			builder.WriteString("\n   AND ")
			builder.WriteString(expr)
			builder.WriteString(" IS NULL")
			return nil
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindQueryValue(args, value, cast))
		return nil
	}
	values, ok := arg["values"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("missing equality values for %s", expr)
	}
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindQueryValue(args, value, cast))
	}
	builder.WriteString(")")
	return nil
}

func bindQueryValue(args *[]any, value any, cast string) string {
	*args = append(*args, value)
	placeholder := fmt.Sprintf("$%d", len(*args))
	if cast == "" {
		return placeholder
	}
	return placeholder + "::" + cast
}

func scanIndicatorProjectionRecord(scanner interface {
	Scan(dest ...any) error
}) (IndicatorProjectionRecord, error) {
	var (
		record          IndicatorProjectionRecord
		recordIDText    string
		incidentIDText  string
		rawNormalized   pgtype.Text
		rawDefanged     pgtype.Text
		rawHashAlg      pgtype.Text
		rawHashValue    pgtype.Text
		rawSTIX         pgtype.Text
		rawFirst        pgtype.Timestamptz
		rawLast         pgtype.Timestamptz
		rawLifecycle    pgtype.Text
		observationCnt  int32
		supportingLinks int32
	)
	if err := scanner.Scan(
		&recordIDText,
		&incidentIDText,
		&record.RowVersion,
		&record.IndicatorType,
		&record.ValueKind,
		&record.DisplayValue,
		&rawNormalized,
		&record.DedupeKey,
		&rawDefanged,
		&rawHashAlg,
		&rawHashValue,
		&rawSTIX,
		&rawFirst,
		&rawLast,
		&observationCnt,
		&rawLifecycle,
		&supportingLinks,
		&record.UpdatedAt,
	); err != nil {
		return IndicatorProjectionRecord{}, fmt.Errorf("scan indicator projection record: %w", err)
	}
	record.RecordID = uuid.MustParse(recordIDText)
	record.IncidentID = uuid.MustParse(incidentIDText)
	record.NormalizedValue = textPointer(rawNormalized)
	record.DefangedValue = textPointer(rawDefanged)
	record.HashAlgorithm = textPointer(rawHashAlg)
	record.HashValue = textPointer(rawHashValue)
	record.STIXPattern = textPointer(rawSTIX)
	record.FirstObservedAt = timePointerFromPG(rawFirst)
	record.LastObservedAt = timePointerFromPG(rawLast)
	record.ObservationCount = int(observationCnt)
	record.LifecycleSummary = textPointer(rawLifecycle)
	record.SupportingLinkCnt = int(supportingLinks)
	return record, nil
}
