package queryengine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	entityports "github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type HostReader struct {
	db postgres.DB
}

func NewHostReader(db postgres.DB) *HostReader {
	return &HostReader{db: db}
}

func (reader *HostReader) SelectHostQueryProjections(
	ctx context.Context,
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) ([]entityports.HostQueryProjection, error) {
	if reader == nil || reader.db == nil {
		return nil, fmt.Errorf("query host projections: database is required")
	}
	sqlText, args, err := buildHostQueryPageSQL(incidentID, query, window)
	if err != nil {
		return nil, err
	}
	rows, err := reader.db.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query host projections: %w", err)
	}
	defer rows.Close()

	result := make([]entityports.HostQueryProjection, 0)
	for rows.Next() {
		row, scanErr := scanHostQueryProjection(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate host projections: %w", err)
	}
	return result, nil
}

func (reader *HostReader) CollectHostDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]entityports.DerivedFact, error) {
	rows, err := tx.Query(ctx, `
SELECT h.record_id, to_jsonb(h) - 'incident_id'
  FROM host_grid_projection h
  JOIN records r
    ON r.incident_id = h.incident_id
   AND r.record_id = h.record_id
   AND r.deleted_at IS NULL
 WHERE h.incident_id = $1
 ORDER BY h.record_id
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("collect host projection facts: %w", err)
	}
	defer rows.Close()

	facts := make([]entityports.DerivedFact, 0)
	for rows.Next() {
		var (
			recordID uuid.UUID
			raw      []byte
		)
		if err := rows.Scan(&recordID, &raw); err != nil {
			return nil, fmt.Errorf("scan host projection fact: %w", err)
		}
		value := map[string]any{}
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, fmt.Errorf("decode host projection fact: %w", err)
		}
		facts = append(facts, entityports.DerivedFact{
			RecordID:     recordID,
			RecordType:   "host",
			ContentClass: "derived_analytic",
			Value:        value,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate host projection facts: %w", err)
	}
	return facts, nil
}

var hostSortExpressions = map[string]string{
	"record_id":               "p.record_id",
	"host.display_name":       "p.display_name",
	"host.hostname":           "p.hostname",
	"host.host_state":         "p.host_state",
	"host.linked_event_count": "p.linked_event_count",
	"host.evidence_count":     "p.evidence_count",
	"host.location":           "p.location",
	"host.os_platform":        "p.os_platform",
	"host.business_owner":     "p.business_owner",
	"host.criticality":        "p.criticality",
	"host.containment_status": "p.containment_status",
	"host.edited_at":          "p.edited_at",
}

func buildHostQueryPageSQL(
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    p.record_id,
    p.host_state,
    p.linked_event_count,
    p.evidence_count,
    p.location,
    p.os_platform,
    p.business_owner,
    p.criticality,
    p.containment_status
  FROM host_grid_projection p
  JOIN records r
    ON r.record_id = p.record_id
   AND r.incident_id = p.incident_id
 WHERE p.incident_id = $1
   AND r.deleted_at IS NULL
   AND p.host_state IN ('stub', 'canonical')`)
	args := []any{incidentID}

	for _, filter := range query.Filters {
		var expression string
		switch filter.FieldKey {
		case "host.host_state":
			expression = "p.host_state"
		case "host.location":
			expression = "p.location"
		case "host.os_platform":
			expression = "p.os_platform"
		case "host.business_owner":
			expression = "p.business_owner"
		case "host.criticality":
			expression = "p.criticality"
		case "host.containment_status":
			expression = "p.containment_status"
		default:
			return "", nil, fmt.Errorf("host query filter field %q not mapped", filter.FieldKey)
		}
		if err := appendEntityQueryTextClause(&builder, &args, expression, filter); err != nil {
			return "", nil, err
		}
	}

	if err := querypage.AppendKeyset(
		&builder,
		&args,
		query.Sort,
		entityPageFields(hostSortExpressions),
		window.Position,
	); err != nil {
		return "", nil, err
	}
	if err := appendEntityOrderBy(&builder, query.Sort, hostSortExpressions); err != nil {
		return "", nil, err
	}
	if err := querypage.AppendLimit(&builder, &args, window.Limit); err != nil {
		return "", nil, err
	}
	return builder.String(), args, nil
}

func scanHostQueryProjection(scanner interface{ Scan(...any) error }) (entityports.HostQueryProjection, error) {
	var (
		row               entityports.HostQueryProjection
		location          pgtype.Text
		osPlatform        pgtype.Text
		businessOwner     pgtype.Text
		criticality       pgtype.Text
		containmentStatus pgtype.Text
		linkedEventCount  int32
		evidenceCount     int32
	)
	if err := scanner.Scan(
		&row.RecordID,
		&row.HostState,
		&linkedEventCount,
		&evidenceCount,
		&location,
		&osPlatform,
		&businessOwner,
		&criticality,
		&containmentStatus,
	); err != nil {
		return entityports.HostQueryProjection{}, fmt.Errorf("scan host projection query row: %w", err)
	}
	row.LinkedEventCount = int(linkedEventCount)
	row.EvidenceCount = int(evidenceCount)
	row.Location = queryTextPointer(location)
	row.OSPlatform = queryTextPointer(osPlatform)
	row.BusinessOwner = queryTextPointer(businessOwner)
	row.Criticality = queryTextPointer(criticality)
	row.ContainmentStatus = queryTextPointer(containmentStatus)
	return row, nil
}

func entityPageFields(expressions map[string]string) map[string]querypage.Field {
	fields := make(map[string]querypage.Field, len(expressions))
	for key, expression := range expressions {
		cast := ""
		switch {
		case key == "record_id":
			cast = "uuid"
		case strings.HasSuffix(key, "_count"):
			cast = "bigint"
		case strings.HasSuffix(key, ".edited_at"):
			cast = "timestamptz"
		}
		fields[key] = querypage.Field{Expression: expression, Cast: cast}
	}
	return fields
}

func appendEntityOrderBy(
	builder *strings.Builder,
	sortEntries []viewschema.SortEntry,
	expressions map[string]string,
) error {
	builder.WriteString(" ORDER BY ")
	for index, sortEntry := range sortEntries {
		if index > 0 {
			builder.WriteString(", ")
		}
		expression, ok := expressions[sortEntry.FieldKey]
		if !ok {
			return fmt.Errorf("query sort field %q not mapped", sortEntry.FieldKey)
		}
		builder.WriteString(expression)
		if sortEntry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
		builder.WriteString(" NULLS LAST")
	}
	return nil
}

func appendEntityQueryTextClause(
	builder *strings.Builder,
	args *[]any,
	expression string,
	filter viewschema.Filter,
) error {
	switch filter.Op {
	case "eq":
		return appendEntityCaseFoldedEqualityClause(builder, args, expression, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		builder.WriteString("\n   AND left(lower(coalesce((")
		builder.WriteString(expression)
		builder.WriteString(")::text, '')), char_length(")
		builder.WriteString(bindEntityQueryValue(args, value))
		builder.WriteString(")) = ")
		builder.WriteString(bindEntityQueryValue(args, value))
		return nil
	default:
		return fmt.Errorf("text filter operator %q not mapped", filter.Op)
	}
}

func appendEntityCaseFoldedEqualityClause(
	builder *strings.Builder,
	args *[]any,
	expression string,
	arg map[string]any,
) error {
	if value, ok := arg["value"]; ok {
		if value == nil {
			builder.WriteString("\n   AND ")
			builder.WriteString(expression)
			builder.WriteString(" IS NULL")
			return nil
		}
		builder.WriteString("\n   AND lower(coalesce((")
		builder.WriteString(expression)
		builder.WriteString(")::text, '')) = ")
		builder.WriteString(bindEntityQueryValue(args, value))
		return nil
	}
	values, ok := arg["values"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("missing equality values for %s", expression)
	}
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString("lower(coalesce((")
		builder.WriteString(expression)
		builder.WriteString(")::text, '')) = ")
		builder.WriteString(bindEntityQueryValue(args, value))
	}
	builder.WriteString(")")
	return nil
}

func bindEntityQueryValue(args *[]any, value any) string {
	*args = append(*args, value)
	return fmt.Sprintf("$%d", len(*args))
}

func queryTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
