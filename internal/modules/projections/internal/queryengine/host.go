package queryengine

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	entityports "github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type HostReader struct {
	db dbport.DB
}

func NewHostReader(db dbport.DB) *HostReader {
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
	rows, err := reader.db.Query(ctx, sqlshape.Text(sqlText), args...)
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

func (reader *HostReader) SelectHostQueryProjectionsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) ([]entityports.HostQueryProjection, error) {
	if reader == nil || tx == nil {
		return nil, fmt.Errorf("query host projections: database is required")
	}
	sqlText, args, err := buildHostQueryPageSQL(incidentID, query, window)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, sqlshape.Text(sqlText), args...)
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
SELECT h.record_id, jsonb_build_object(
       'record_id', h.record_id,
       'row_version', h.row_version,
       'display_name', h.display_name,
       'hostname', h.hostname,
       'host_state', h.host_state,
       'linked_event_count', h.linked_event_count,
       'evidence_count', h.evidence_count,
       'location', h.location,
       'os_platform', h.os_platform,
       'business_owner', h.business_owner,
       'criticality', h.criticality,
       'containment_status', h.containment_status,
       'edited_at', h.edited_at
   )
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
		value, err := decodeReportingFactObject(raw)
		if err != nil {
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

var hostSortExpressions = map[string]sqlshape.Fragment{
	"record_id":               sqlshape.Literal("p.record_id"),
	"host.display_name":       sqlshape.Literal("p.display_name"),
	"host.hostname":           sqlshape.Literal("p.hostname"),
	"host.host_state":         sqlshape.Literal("p.host_state"),
	"host.linked_event_count": sqlshape.Literal("p.linked_event_count"),
	"host.evidence_count":     sqlshape.Literal("p.evidence_count"),
	"host.location":           sqlshape.Literal("p.location"),
	"host.os_platform":        sqlshape.Literal("p.os_platform"),
	"host.business_owner":     sqlshape.Literal("p.business_owner"),
	"host.criticality":        sqlshape.Literal("p.criticality"),
	"host.containment_status": sqlshape.Literal("p.containment_status"),
	"host.edited_at":          sqlshape.Literal("p.edited_at"),
}

func buildHostQueryPageSQL(
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) (sqlshape.Fragment, []any, error) {
	var builder sqlshape.Builder
	builder.Write(sqlshape.Literal(`
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
   AND p.host_state IN ('stub', 'canonical')`))
	args := []any{incidentID}
	if window.TargetRecordID != "" {
		args = append(args, window.TargetRecordID)
		builder.Write(sqlshape.Literal(" AND p.record_id = $2::uuid"))
	}

	for _, filter := range query.Filters {
		var expression sqlshape.Fragment
		switch filter.FieldKey {
		case "host.host_state":
			expression = sqlshape.Literal("p.host_state")
		case "host.location":
			expression = sqlshape.Literal("p.location")
		case "host.os_platform":
			expression = sqlshape.Literal("p.os_platform")
		case "host.business_owner":
			expression = sqlshape.Literal("p.business_owner")
		case "host.criticality":
			expression = sqlshape.Literal("p.criticality")
		case "host.containment_status":
			expression = sqlshape.Literal("p.containment_status")
		default:
			return sqlshape.Fragment{}, nil, fmt.Errorf("host query filter field %q not mapped", filter.FieldKey)
		}
		if err := appendEntityQueryTextClause(&builder, &args, expression, filter); err != nil {
			return sqlshape.Fragment{}, nil, err
		}
	}

	appendPosition := querypage.AppendKeyset
	if window.Before {
		appendPosition = querypage.AppendBefore
	}
	if err := appendPosition(
		&builder,
		&args,
		query.Sort,
		entityPageFields(hostSortExpressions),
		window.Position,
	); err != nil {
		return sqlshape.Fragment{}, nil, err
	}
	if err := appendEntityOrderBy(&builder, query.Sort, hostSortExpressions, window.Before); err != nil {
		return sqlshape.Fragment{}, nil, err
	}
	if err := querypage.AppendLimit(&builder, &args, window.Limit); err != nil {
		return sqlshape.Fragment{}, nil, err
	}
	return builder.Fragment(), args, nil
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

func entityPageFields(expressions map[string]sqlshape.Fragment) map[string]querypage.Field {
	fields := make(map[string]querypage.Field, len(expressions))
	for key, expression := range expressions {
		var cast sqlshape.Fragment
		switch {
		case key == "record_id":
			cast = sqlshape.Literal("uuid")
		case strings.HasSuffix(key, "_count"):
			cast = sqlshape.Literal("bigint")
		case strings.HasSuffix(key, ".edited_at"):
			cast = sqlshape.Literal("timestamptz")
		}
		fields[key] = querypage.Field{Expression: expression, Cast: cast}
	}
	return fields
}

func appendEntityOrderBy(
	builder *sqlshape.Builder,
	sortEntries []viewschema.SortEntry,
	expressions map[string]sqlshape.Fragment,
	before bool,
) error {
	builder.Write(sqlshape.Literal(" ORDER BY "))
	for index, sortEntry := range sortEntries {
		if index > 0 {
			builder.Write(sqlshape.Literal(", "))
		}
		expression, ok := expressions[sortEntry.FieldKey]
		if !ok {
			return fmt.Errorf("query sort field %q not mapped", sortEntry.FieldKey)
		}
		builder.Write(expression)
		if (sortEntry.Direction == "desc") != before {
			builder.Write(sqlshape.Literal(" DESC"))
		} else {
			builder.Write(sqlshape.Literal(" ASC"))
		}
		if before {
			builder.Write(sqlshape.Literal(" NULLS FIRST"))
		} else {
			builder.Write(sqlshape.Literal(" NULLS LAST"))
		}
	}
	return nil
}

func appendEntityQueryTextClause(
	builder *sqlshape.Builder,
	args *[]any,
	expression sqlshape.Fragment,
	filter viewschema.Filter,
) error {
	switch filter.Op {
	case "eq":
		return appendEntityCaseFoldedEqualityClause(builder, args, expression, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		builder.Write(sqlshape.Literal("\n   AND left(lower(coalesce(("))
		builder.Write(expression)
		builder.Write(sqlshape.Literal(")::text, '')), char_length("))
		builder.Write(bindEntityQueryValue(args, value))
		builder.Write(sqlshape.Literal(")) = "))
		builder.Write(bindEntityQueryValue(args, value))
		return nil
	default:
		return fmt.Errorf("text filter operator %q not mapped", filter.Op)
	}
}

func appendEntityCaseFoldedEqualityClause(
	builder *sqlshape.Builder,
	args *[]any,
	expression sqlshape.Fragment,
	arg map[string]any,
) error {
	if value, ok := arg["value"]; ok {
		if value == nil {
			builder.Write(sqlshape.Literal("\n   AND "))
			builder.Write(expression)
			builder.Write(sqlshape.Literal(" IS NULL"))
			return nil
		}
		builder.Write(sqlshape.Literal("\n   AND lower(coalesce(("))
		builder.Write(expression)
		builder.Write(sqlshape.Literal(")::text, '')) = "))
		builder.Write(bindEntityQueryValue(args, value))
		return nil
	}
	values, ok := arg["values"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("missing equality values for %s", sqlshape.Text(expression))
	}
	builder.Write(sqlshape.Literal("\n   AND ("))
	for index, value := range values {
		if index > 0 {
			builder.Write(sqlshape.Literal(" OR "))
		}
		builder.Write(sqlshape.Literal("lower(coalesce(("))
		builder.Write(expression)
		builder.Write(sqlshape.Literal(")::text, '')) = "))
		builder.Write(bindEntityQueryValue(args, value))
	}
	builder.Write(sqlshape.Literal(")"))
	return nil
}

func bindEntityQueryValue(args *[]any, value any) sqlshape.Fragment {
	*args = append(*args, value)
	return sqlshape.Parameter(len(*args))
}

func queryTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
