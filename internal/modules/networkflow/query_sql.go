package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

type sqlValueExpression struct {
	expression sqlshape.Fragment
	cast       sqlshape.Fragment
	value      any
	nullable   bool
	direction  string
}

func (s *store) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, tableIDs []string, filters []queryFilter, sortSpecs []sortSpec, position *rowCursorPosition, limit int) ([]flowRow, bool, error) {
	if len(tableIDs) == 0 || limit < 1 {
		return []flowRow{}, false, nil
	}
	args := []any{incidentID, tableIDs}
	where := []sqlshape.Fragment{sqlshape.Literal("incident_id = $1"), sqlshape.Literal("network_flow_table_id = ANY($2::text[])")}
	for _, filter := range filters {
		clause, err := appendRowFilterSQL(filter, &args)
		if err != nil {
			return nil, false, err
		}
		where = append(where, clause)
	}
	effective := effectiveSort(sortSpecs)
	if position != nil {
		clause, err := appendRowKeysetSQL(effective, *position, &args)
		if err != nil {
			return nil, false, err
		}
		where = append(where, clause)
	}
	order := make([]sqlshape.Fragment, 0, len(effective)+2)
	for _, spec := range effective {
		expression, ok := rowSQLExpression(spec.FieldKey)
		if !ok {
			return nil, false, fmt.Errorf("unsupported Network Flow sort field %q", spec.FieldKey)
		}
		direction := sqlshape.Literal("ASC")
		if spec.Direction == "desc" {
			direction = sqlshape.Literal("DESC")
		}
		order = append(order, sqlshape.Format("%s %s NULLS LAST", expression, direction))
	}
	order = append(order, sqlshape.Literal(`network_flow_table_id COLLATE "C" ASC`), sqlshape.Literal(`network_flow_row_id COLLATE "C" ASC`))
	args = append(args, limit+1)
	query := sqlshape.Format(`SELECT %s FROM network_flow_rows WHERE %s ORDER BY %s LIMIT %s`, sqlshape.Literal(flowRowColumnList()), sqlshape.Join(where, sqlshape.Literal("AND")), sqlshape.Join(order, sqlshape.Literal(",")), sqlshape.Parameter(len(args)))
	rows, err := s.pool.Query(ctx, sqlshape.Text(query), args...)
	if err != nil {
		return nil, false, fmt.Errorf("query Network Flow row page: %w", err)
	}
	defer rows.Close()
	result := make([]flowRow, 0, limit+1)
	for rows.Next() {
		row, err := scanFlowRow(rows)
		if err != nil {
			return nil, false, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("scan Network Flow row page: %w", err)
	}
	hasMore := len(result) > limit
	if hasMore {
		result = result[:limit]
	}
	rows.Close()
	if err := s.resolveRowAuthors(ctx, incidentID, result); err != nil {
		return nil, false, err
	}
	return result, hasMore, nil
}

func appendRowKeysetSQL(effective []sortSpec, position rowCursorPosition, args *[]any) (sqlshape.Fragment, error) {
	if !sameSortSpecs(effective, position.EffectiveSort) || len(position.Values) != len(effective) {
		return sqlshape.Fragment{}, fmt.Errorf("network flow cursor sort tuple mismatch")
	}
	components := make([]sqlValueExpression, 0, len(effective)+2)
	for index, spec := range effective {
		expression, ok := rowSQLExpression(spec.FieldKey)
		if !ok {
			return sqlshape.Fragment{}, fmt.Errorf("unsupported Network Flow cursor field %q", spec.FieldKey)
		}
		value, _, err := rowSQLCursorValue(spec.FieldKey, position.Values[index])
		if err != nil {
			return sqlshape.Fragment{}, err
		}
		components = append(components, sqlValueExpression{expression: expression, cast: sqlshape.Literal(rowSQLCast(spec.FieldKey)), value: value, nullable: rowSQLFieldNullable(spec.FieldKey), direction: spec.Direction})
	}
	components = append(components,
		sqlValueExpression{expression: sqlshape.Literal(`network_flow_table_id COLLATE "C"`), cast: sqlshape.Literal("text"), value: position.NetworkFlowTableID, direction: "asc"},
		sqlValueExpression{expression: sqlshape.Literal(`network_flow_row_id COLLATE "C"`), cast: sqlshape.Literal("text"), value: position.NetworkFlowRowID, direction: "asc"},
	)
	return appendSQLKeysetComponents(components, args)
}

func appendRowFilterSQL(filter queryFilter, args *[]any) (clause sqlshape.Fragment, err error) {
	defer func() {
		if err == nil && directionalQueryField(filter.FieldKey) {
			clause = sqlshape.Format("(direction_kind <> 'unknown' AND (%s))", clause)
		}
	}()

	if filter.FieldKey == fieldEndpointIP {
		return appendEndpointFilterSQL(filter, args)
	}
	expression, ok := rowSQLExpression(filter.FieldKey)
	if !ok {
		return sqlshape.Fragment{}, fmt.Errorf("unsupported Network Flow filter field %q", filter.FieldKey)
	}
	cast := sqlshape.Literal(rowSQLCast(filter.FieldKey))
	expr := expression
	appendValue := func(value any) sqlshape.Fragment {
		*args = append(*args, rowSQLArgument(filter.FieldKey, value))
		return sqlshape.Format("%s::%s", sqlshape.Parameter(len(*args)), cast)
	}
	switch filter.Op {
	case "is_null":
		return sqlshape.Format("%s IS NULL", expr), nil
	case "not_null":
		return sqlshape.Format("%s IS NOT NULL", expr), nil
	case "eq":
		return sqlshape.Format("%s = %s", expr, appendValue(filter.Value)), nil
	case "in":
		values, ok := filter.Value.([]any)
		if !ok || len(values) == 0 {
			return sqlshape.Fragment{}, fmt.Errorf("invalid Network Flow in filter")
		}
		clauses := make([]sqlshape.Fragment, 0, len(values))
		for _, value := range values {
			clauses = append(clauses, sqlshape.Format("%s = %s", expr, appendValue(value)))
		}
		return sqlshape.Format("(%s)", sqlshape.Join(clauses, sqlshape.Literal("OR"))), nil
	case "range":
		object, ok := filter.Value.(map[string]any)
		if !ok {
			return sqlshape.Fragment{}, fmt.Errorf("invalid Network Flow range filter")
		}
		clauses := []sqlshape.Fragment{}
		if value, exists := object["gte"]; exists && value != nil {
			clauses = append(clauses, sqlshape.Format("%s >= %s", expr, appendValue(value)))
		}
		if value, exists := object["lte"]; exists && value != nil {
			clauses = append(clauses, sqlshape.Format("%s <= %s", expr, appendValue(value)))
		}
		if value, exists := object["lt"]; exists && value != nil {
			clauses = append(clauses, sqlshape.Format("%s < %s", expr, appendValue(value)))
		}
		if len(clauses) == 0 {
			return sqlshape.Fragment{}, fmt.Errorf("empty Network Flow range filter")
		}
		return sqlshape.Format("(%s)", sqlshape.Join(clauses, sqlshape.Literal("AND"))), nil
	case "prefix":
		return sqlshape.Format("left(%s,char_length(%s)) = %s::%s", expr, appendValue(filter.Value), sqlshape.Parameter(len(*args)), cast), nil
	case "contains":
		return sqlshape.Format("strpos(%s,%s) > 0", expr, appendValue(filter.Value)), nil
	case "cidr_contains":
		*args = append(*args, filter.Value)
		return sqlshape.Format("%s <<= %s::cidr", expr, sqlshape.Parameter(len(*args))), nil
	default:
		return sqlshape.Fragment{}, fmt.Errorf("unsupported Network Flow filter operator %q", filter.Op)
	}
}

func appendEndpointFilterSQL(filter queryFilter, args *[]any) (sqlshape.Fragment, error) {
	if filter.Op == "in" {
		values, ok := filter.Value.([]any)
		if !ok || len(values) == 0 {
			return sqlshape.Fragment{}, fmt.Errorf("invalid Network Flow endpoint in filter")
		}
		clauses := make([]sqlshape.Fragment, 0, len(values))
		for _, value := range values {
			*args = append(*args, value)
			placeholder := sqlshape.Format("%s::inet", sqlshape.Parameter(len(*args)))
			clauses = append(clauses, sqlshape.Format("(endpoint_a_ip::inet = %s OR endpoint_b_ip::inet = %s)", placeholder, placeholder))
		}
		return sqlshape.Format("(%s)", sqlshape.Join(clauses, sqlshape.Literal("OR"))), nil
	}
	*args = append(*args, filter.Value)
	placeholder := sqlshape.Parameter(len(*args))
	switch filter.Op {
	case "eq":
		return sqlshape.Format("(endpoint_a_ip::inet = %s::inet OR endpoint_b_ip::inet = %s::inet)", placeholder, placeholder), nil
	case "cidr_contains":
		return sqlshape.Format("(endpoint_a_ip::inet <<= %s::cidr OR endpoint_b_ip::inet <<= %s::cidr)", placeholder, placeholder), nil
	default:
		return sqlshape.Fragment{}, fmt.Errorf("unsupported Network Flow endpoint operator %q", filter.Op)
	}
}

func rowSQLExpression(field string) (sqlshape.Fragment, bool) {
	if directionalQueryField(field) {
		first, second := sqlshape.Literal("endpoint_a_ip::inet"), sqlshape.Literal("endpoint_b_ip::inet")
		if field == fieldFromPort || field == fieldToPort {
			first, second = sqlshape.Literal("endpoint_a_port"), sqlshape.Literal("endpoint_b_port")
		}
		if field == fieldToIP || field == fieldToPort {
			first, second = second, first
		}
		return sqlshape.Format("(CASE from_endpoint WHEN 'a' THEN %s WHEN 'b' THEN %s ELSE NULL END)", first, second), true
	}
	if field == fieldDirectionKind {
		return sqlshape.Literal(`direction_kind COLLATE "C"`), true
	}
	if field == fieldDirectionReason {
		return sqlshape.Literal(`NULLIF(direction_reason,'') COLLATE "C"`), true
	}

	switch field {
	case fieldEndpointAIP:
		return sqlshape.Literal("endpoint_a_ip::inet"), true
	case fieldEndpointBIP:
		return sqlshape.Literal("endpoint_b_ip::inet"), true
	case fieldEndpointAPort:
		return sqlshape.Literal("endpoint_a_port"), true
	case fieldEndpointBPort:
		return sqlshape.Literal("endpoint_b_port"), true
	case fieldIPProtocol:
		return sqlshape.Literal("ip_protocol"), true
	case fieldFlowStartUTC:
		return sqlshape.Literal("flow_start_utc"), true
	case fieldFlowEndUTC:
		return sqlshape.Literal("flow_end_utc"), true
	case fieldBytesCount:
		return sqlshape.Literal("bytes_count::numeric"), true
	case fieldPacketsCount:
		return sqlshape.Literal("packets_count::numeric"), true
	case fieldExporterID:
		return sqlshape.Literal(`exporter_id COLLATE "C"`), true
	case fieldInputInterface:
		return sqlshape.Literal(`input_interface COLLATE "C"`), true
	case fieldOutputInterface:
		return sqlshape.Literal(`output_interface COLLATE "C"`), true
	case "source_row_number":
		return sqlshape.Literal("source_row_number"), true
	case "network_flow_row_id":
		return sqlshape.Literal(`network_flow_row_id COLLATE "C"`), true
	case "network_flow_table_id":
		return sqlshape.Literal(`network_flow_table_id COLLATE "C"`), true
	default:
		return sqlshape.Fragment{}, false
	}
}

func rowSQLCast(field string) string {
	switch field {
	case fieldFromIP, fieldToIP, fieldEndpointAIP, fieldEndpointBIP:
		return "inet"
	case fieldFromPort, fieldToPort, fieldEndpointAPort, fieldEndpointBPort, fieldIPProtocol:
		return "integer"
	case fieldFlowStartUTC, fieldFlowEndUTC:
		return "timestamptz"
	case fieldBytesCount, fieldPacketsCount:
		return "numeric"
	case "source_row_number":
		return "bigint"
	default:
		return "text"
	}
}

func rowSQLCursorValue(field string, value any) (any, string, error) {
	if value == nil {
		return nil, rowSQLCast(field), nil
	}
	if field == fieldFlowStartUTC || field == fieldFlowEndUTC {
		text, ok := value.(string)
		if !ok {
			return nil, "", fmt.Errorf("invalid Network Flow timestamp cursor value")
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return nil, "", fmt.Errorf("invalid Network Flow timestamp cursor value: %w", err)
		}
		return parsed.UTC(), "timestamptz", nil
	}
	return rowSQLArgument(field, value), rowSQLCast(field), nil
}

func rowSQLArgument(field string, value any) any {
	if number, ok := value.(json.Number); ok {
		return number.String()
	}
	return value
}

func rowSQLFieldNullable(field string) bool {
	if directionalQueryField(field) || field == fieldDirectionReason {
		return true
	}
	switch field {
	case fieldFromPort, fieldToPort, fieldEndpointAPort, fieldEndpointBPort, fieldExporterID, fieldInputInterface, fieldOutputInterface:
		return true
	default:
		return false
	}
}

func flowRowColumnList() string {
	return `network_flow_row_id, network_flow_table_id, incident_id, source_row_number,
       source_row_digest_sha256, normalized_row_digest_sha256, mapping_fingerprint,
       flow_start_utc, flow_end_utc, endpoint_a_ip, endpoint_b_ip, endpoint_a_port, endpoint_b_port, ip_protocol,
       bytes_count, packets_count, exporter_id, input_interface, output_interface,
       tcp_flags, application_label, unmapped_raw, observation_source_ref, created_at, created_by_user_id, direction_kind, from_endpoint, direction_reason, direction_evidence`
}

func (s *store) QueryRejectedDiagnosticsPage(ctx context.Context, incidentID uuid.UUID, tableID string, request rejectedRowsQueryRequest, position *diagnosticCursorPosition, limit int) ([]rejectedRowDiagnostic, bool, error) {
	args := []any{incidentID, tableID}
	where := []sqlshape.Fragment{sqlshape.Literal("incident_id = $1"), sqlshape.Literal("network_flow_table_id = $2")}
	if len(request.ErrorCodes) > 0 {
		args = append(args, request.ErrorCodes)
		where = append(where, sqlshape.Format("error_code = ANY(%s::text[])", sqlshape.Parameter(len(args))))
	}
	if len(request.FieldKeys) > 0 {
		args = append(args, request.FieldKeys)
		where = append(where, sqlshape.Format("field_key = ANY(%s::text[])", sqlshape.Parameter(len(args))))
	}
	if request.SourceRowGTE != nil {
		args = append(args, *request.SourceRowGTE)
		where = append(where, sqlshape.Format("source_row_number >= %s::bigint", sqlshape.Parameter(len(args))))
	}
	if request.SourceRowLTE != nil {
		args = append(args, *request.SourceRowLTE)
		where = append(where, sqlshape.Format("source_row_number <= %s::bigint", sqlshape.Parameter(len(args))))
	}
	if position != nil {
		components := []sqlValueExpression{
			{expression: sqlshape.Literal("source_row_number"), cast: sqlshape.Literal("bigint"), value: position.SourceRowNumber},
			{expression: sqlshape.Literal("source_column_ordinal"), cast: sqlshape.Literal("bigint"), value: pointerValue(position.SourceColumnOrdinal), nullable: true},
			{expression: sqlshape.Literal(`field_key COLLATE "C"`), cast: sqlshape.Literal("text"), value: pointerValue(position.FieldKey), nullable: true},
			{expression: sqlshape.Literal(`error_code COLLATE "C"`), cast: sqlshape.Literal("text"), value: position.ErrorCode},
			{expression: sqlshape.Literal(`reason_code COLLATE "C"`), cast: sqlshape.Literal("text"), value: position.ReasonCode},
			{expression: sqlshape.Literal(`diagnostic_id COLLATE "C"`), cast: sqlshape.Literal("text"), value: position.DiagnosticID},
		}
		clause, err := appendSQLKeysetComponents(components, &args)
		if err != nil {
			return nil, false, err
		}
		where = append(where, clause)
	}
	args = append(args, limit+1)
	query := sqlshape.Format(`SELECT diagnostic_id, source_row_number, source_column_ordinal, raw_header_sha256,
       field_key, error_code, reason_code, safe_sample, raw_value_sha256, message_key,
       message_args, message, limit_name, limit_value, actual_value
  FROM network_flow_rejected_row_diagnostics
 WHERE %s
 ORDER BY source_row_number ASC, source_column_ordinal ASC NULLS LAST,
          field_key COLLATE "C" ASC NULLS LAST, error_code COLLATE "C" ASC,
          reason_code COLLATE "C" ASC, diagnostic_id COLLATE "C" ASC
 LIMIT %s`, sqlshape.Join(where, sqlshape.Literal("AND")), sqlshape.Parameter(len(args)))
	rows, err := s.pool.Query(ctx, sqlshape.Text(query), args...)
	if err != nil {
		return nil, false, fmt.Errorf("query Network Flow diagnostic page: %w", err)
	}
	defer rows.Close()
	result := make([]rejectedRowDiagnostic, 0, limit+1)
	for rows.Next() {
		diagnostic, err := scanRejectedRowDiagnostic(rows)
		if err != nil {
			return nil, false, err
		}
		result = append(result, diagnostic)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("scan Network Flow diagnostic page: %w", err)
	}
	hasMore := len(result) > limit
	if hasMore {
		result = result[:limit]
	}
	return result, hasMore, nil
}

func appendSQLKeysetComponents(components []sqlValueExpression, args *[]any) (sqlshape.Fragment, error) {
	prefix := []sqlshape.Fragment{}
	disjunction := []sqlshape.Fragment{}
	for _, component := range components {
		placeholder := sqlshape.Format("%s::%s", sqlshape.Parameter(len(*args)+1), component.cast)
		*args = append(*args, component.value)
		equal := sqlshape.Format("%s IS NOT DISTINCT FROM %s", component.expression, placeholder)
		if component.value != nil {
			operator := sqlshape.Literal(">")
			if component.direction == "desc" {
				operator = sqlshape.Literal("<")
			}
			after := sqlshape.Format("%s %s %s", component.expression, operator, placeholder)
			if component.nullable {
				after = sqlshape.Format("(%s OR %s IS NULL)", after, component.expression)
			}
			if len(prefix) > 0 {
				after = sqlshape.Format("(%s AND %s)", sqlshape.Join(prefix, sqlshape.Literal("AND")), after)
			}
			disjunction = append(disjunction, after)
		}
		prefix = append(prefix, equal)
	}
	if len(disjunction) == 0 {
		return sqlshape.Literal("FALSE"), nil
	}
	return sqlshape.Format("(%s)", sqlshape.Join(disjunction, sqlshape.Literal("OR"))), nil
}

func pointerValue[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
