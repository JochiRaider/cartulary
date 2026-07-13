package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type sqlValueExpression struct {
	expression string
	cast       string
	value      any
	nullable   bool
	direction  string
}

func (s *Store) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, tableIDs []string, filters []Filter, sortSpecs []SortSpec, position *rowCursorPosition, limit int) ([]FlowRow, bool, error) {
	if len(tableIDs) == 0 || limit < 1 {
		return []FlowRow{}, false, nil
	}
	args := []any{incidentID, tableIDs}
	where := []string{"incident_id = $1", "network_flow_table_id = ANY($2::text[])"}
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
	order := make([]string, 0, len(effective)+2)
	for _, spec := range effective {
		expression, ok := rowSQLExpression(spec.FieldKey)
		if !ok {
			return nil, false, fmt.Errorf("unsupported Network Flow sort field %q", spec.FieldKey)
		}
		order = append(order, expression+" "+strings.ToUpper(spec.Direction)+" NULLS LAST")
	}
	order = append(order, `network_flow_table_id COLLATE "C" ASC`, `network_flow_row_id COLLATE "C" ASC`)
	args = append(args, limit+1)
	query := `SELECT ` + flowRowColumnList() + `
  FROM network_flow_rows
 WHERE ` + strings.Join(where, "\n   AND ") + `
 ORDER BY ` + strings.Join(order, ", ") + `
 LIMIT $` + fmt.Sprint(len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query Network Flow row page: %w", err)
	}
	defer rows.Close()
	result := make([]FlowRow, 0, limit+1)
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
	return result, hasMore, nil
}

func appendRowKeysetSQL(effective []SortSpec, position rowCursorPosition, args *[]any) (string, error) {
	if !sameSortSpecs(effective, position.EffectiveSort) || len(position.Values) != len(effective) {
		return "", fmt.Errorf("network flow cursor sort tuple mismatch")
	}
	components := make([]sqlValueExpression, 0, len(effective)+2)
	for index, spec := range effective {
		expression, ok := rowSQLExpression(spec.FieldKey)
		if !ok {
			return "", fmt.Errorf("unsupported Network Flow cursor field %q", spec.FieldKey)
		}
		value, cast, err := rowSQLCursorValue(spec.FieldKey, position.Values[index])
		if err != nil {
			return "", err
		}
		components = append(components, sqlValueExpression{expression: expression, cast: cast, value: value, nullable: rowSQLFieldNullable(spec.FieldKey), direction: spec.Direction})
	}
	components = append(components,
		sqlValueExpression{expression: `network_flow_table_id COLLATE "C"`, cast: "text", value: position.NetworkFlowTableID, direction: "asc"},
		sqlValueExpression{expression: `network_flow_row_id COLLATE "C"`, cast: "text", value: position.NetworkFlowRowID, direction: "asc"},
	)
	prefix := make([]string, 0, len(components))
	disjunction := make([]string, 0, len(components))
	for _, component := range components {
		placeholder := "$" + fmt.Sprint(len(*args)+1) + "::" + component.cast
		*args = append(*args, component.value)
		equal := component.expression + " IS NOT DISTINCT FROM " + placeholder
		if component.value != nil {
			operator := ">"
			if component.direction == "desc" {
				operator = "<"
			}
			after := component.expression + " " + operator + " " + placeholder
			if component.nullable {
				after = "(" + after + " OR " + component.expression + " IS NULL)"
			}
			term := after
			if len(prefix) > 0 {
				term = "(" + strings.Join(prefix, " AND ") + " AND " + after + ")"
			}
			disjunction = append(disjunction, term)
		}
		prefix = append(prefix, equal)
	}
	if len(disjunction) == 0 {
		return "FALSE", nil
	}
	return "(" + strings.Join(disjunction, " OR ") + ")", nil
}

func appendRowFilterSQL(filter Filter, args *[]any) (string, error) {
	if filter.FieldKey == FieldEndpointIP {
		return appendEndpointFilterSQL(filter, args)
	}
	expression, ok := rowSQLExpression(filter.FieldKey)
	if !ok {
		return "", fmt.Errorf("unsupported Network Flow filter field %q", filter.FieldKey)
	}
	cast := rowSQLCast(filter.FieldKey)
	appendValue := func(value any) string {
		*args = append(*args, rowSQLArgument(filter.FieldKey, value))
		return "$" + fmt.Sprint(len(*args)) + "::" + cast
	}
	switch filter.Op {
	case "is_null":
		return expression + " IS NULL", nil
	case "not_null":
		return expression + " IS NOT NULL", nil
	case "eq":
		return expression + " = " + appendValue(filter.Value), nil
	case "in":
		values, ok := filter.Value.([]any)
		if !ok || len(values) == 0 {
			return "", fmt.Errorf("invalid Network Flow in filter")
		}
		clauses := make([]string, 0, len(values))
		for _, value := range values {
			clauses = append(clauses, expression+" = "+appendValue(value))
		}
		return "(" + strings.Join(clauses, " OR ") + ")", nil
	case "range":
		object, ok := filter.Value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid Network Flow range filter")
		}
		clauses := []string{}
		if value, exists := object["gte"]; exists && value != nil {
			clauses = append(clauses, expression+" >= "+appendValue(value))
		}
		if value, exists := object["lte"]; exists && value != nil {
			clauses = append(clauses, expression+" <= "+appendValue(value))
		}
		if value, exists := object["lt"]; exists && value != nil {
			clauses = append(clauses, expression+" < "+appendValue(value))
		}
		if len(clauses) == 0 {
			return "", fmt.Errorf("empty Network Flow range filter")
		}
		return "(" + strings.Join(clauses, " AND ") + ")", nil
	case "prefix":
		return "left(" + expression + ", char_length(" + appendValue(filter.Value) + ")) = " + "$" + fmt.Sprint(len(*args)) + "::" + cast, nil
	case "contains":
		return "strpos(" + expression + ", " + appendValue(filter.Value) + ") > 0", nil
	case "cidr_contains":
		*args = append(*args, filter.Value)
		return expression + " <<= $" + fmt.Sprint(len(*args)) + "::cidr", nil
	default:
		return "", fmt.Errorf("unsupported Network Flow filter operator %q", filter.Op)
	}
}

func appendEndpointFilterSQL(filter Filter, args *[]any) (string, error) {
	if filter.Op == "in" {
		values, ok := filter.Value.([]any)
		if !ok || len(values) == 0 {
			return "", fmt.Errorf("invalid Network Flow endpoint in filter")
		}
		clauses := make([]string, 0, len(values))
		for _, value := range values {
			*args = append(*args, value)
			placeholder := "$" + fmt.Sprint(len(*args)) + "::inet"
			clauses = append(clauses, "(src_ip::inet = "+placeholder+" OR dst_ip::inet = "+placeholder+")")
		}
		return "(" + strings.Join(clauses, " OR ") + ")", nil
	}
	*args = append(*args, filter.Value)
	placeholder := "$" + fmt.Sprint(len(*args))
	switch filter.Op {
	case "eq":
		return "(src_ip::inet = " + placeholder + "::inet OR dst_ip::inet = " + placeholder + "::inet)", nil
	case "cidr_contains":
		return "(src_ip::inet <<= " + placeholder + "::cidr OR dst_ip::inet <<= " + placeholder + "::cidr)", nil
	default:
		return "", fmt.Errorf("unsupported Network Flow endpoint operator %q", filter.Op)
	}
}

func rowSQLExpression(field string) (string, bool) {
	switch field {
	case FieldSrcIP:
		return "src_ip::inet", true
	case FieldDstIP:
		return "dst_ip::inet", true
	case FieldSrcPort:
		return "src_port", true
	case FieldDstPort:
		return "dst_port", true
	case FieldIPProtocol:
		return "ip_protocol", true
	case FieldFlowStartUTC:
		return "flow_start_utc", true
	case FieldFlowEndUTC:
		return "flow_end_utc", true
	case FieldBytesCount:
		return "bytes_count::numeric", true
	case FieldPacketsCount:
		return "packets_count::numeric", true
	case FieldExporterID:
		return `exporter_id COLLATE "C"`, true
	case FieldInputInterface:
		return `input_interface COLLATE "C"`, true
	case FieldOutputInterface:
		return `output_interface COLLATE "C"`, true
	case "source_row_number":
		return "source_row_number", true
	case "network_flow_row_id":
		return `network_flow_row_id COLLATE "C"`, true
	case "network_flow_table_id":
		return `network_flow_table_id COLLATE "C"`, true
	default:
		return "", false
	}
}

func rowSQLCast(field string) string {
	switch field {
	case FieldSrcIP, FieldDstIP:
		return "inet"
	case FieldSrcPort, FieldDstPort, FieldIPProtocol:
		return "integer"
	case FieldFlowStartUTC, FieldFlowEndUTC:
		return "timestamptz"
	case FieldBytesCount, FieldPacketsCount:
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
	if field == FieldFlowStartUTC || field == FieldFlowEndUTC {
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
	switch field {
	case FieldSrcPort, FieldDstPort, FieldExporterID, FieldInputInterface, FieldOutputInterface:
		return true
	default:
		return false
	}
}

func flowRowColumnList() string {
	return `network_flow_row_id, network_flow_table_id, incident_id, source_row_number,
       source_row_digest_sha256, normalized_row_digest_sha256, mapping_fingerprint,
       flow_start_utc, flow_end_utc, src_ip, dst_ip, src_port, dst_port, ip_protocol,
       bytes_count, packets_count, exporter_id, input_interface, output_interface,
       tcp_flags, application_label, unmapped_raw, observation_source_ref, created_at, created_by_user_id`
}

func (s *Store) QueryRejectedDiagnosticsPage(ctx context.Context, incidentID uuid.UUID, tableID string, request RejectedRowsQueryRequest, position *diagnosticCursorPosition, limit int) ([]RejectedRowDiagnostic, bool, error) {
	args := []any{incidentID, tableID}
	where := []string{"incident_id = $1", "network_flow_table_id = $2"}
	if len(request.ErrorCodes) > 0 {
		args = append(args, request.ErrorCodes)
		where = append(where, "error_code = ANY($"+fmt.Sprint(len(args))+"::text[])")
	}
	if len(request.FieldKeys) > 0 {
		args = append(args, request.FieldKeys)
		where = append(where, "field_key = ANY($"+fmt.Sprint(len(args))+"::text[])")
	}
	if request.SourceRowGTE != nil {
		args = append(args, *request.SourceRowGTE)
		where = append(where, "source_row_number >= $"+fmt.Sprint(len(args))+"::bigint")
	}
	if request.SourceRowLTE != nil {
		args = append(args, *request.SourceRowLTE)
		where = append(where, "source_row_number <= $"+fmt.Sprint(len(args))+"::bigint")
	}
	if position != nil {
		components := []sqlValueExpression{
			{expression: "source_row_number", cast: "bigint", value: position.SourceRowNumber},
			{expression: "source_column_ordinal", cast: "bigint", value: pointerValue(position.SourceColumnOrdinal), nullable: true},
			{expression: `field_key COLLATE "C"`, cast: "text", value: pointerValue(position.FieldKey), nullable: true},
			{expression: `error_code COLLATE "C"`, cast: "text", value: position.ErrorCode},
			{expression: `reason_code COLLATE "C"`, cast: "text", value: position.ReasonCode},
			{expression: `diagnostic_id COLLATE "C"`, cast: "text", value: position.DiagnosticID},
		}
		clause, err := appendSQLKeysetComponents(components, &args)
		if err != nil {
			return nil, false, err
		}
		where = append(where, clause)
	}
	args = append(args, limit+1)
	query := `SELECT diagnostic_id, source_row_number, source_column_ordinal, raw_header_sha256,
       field_key, error_code, reason_code, safe_sample, raw_value_sha256, message_key,
       message_args, message, limit_name, limit_value, actual_value
  FROM network_flow_rejected_row_diagnostics
 WHERE ` + strings.Join(where, "\n   AND ") + `
 ORDER BY source_row_number ASC, source_column_ordinal ASC NULLS LAST,
          field_key COLLATE "C" ASC NULLS LAST, error_code COLLATE "C" ASC,
          reason_code COLLATE "C" ASC, diagnostic_id COLLATE "C" ASC
 LIMIT $` + fmt.Sprint(len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query Network Flow diagnostic page: %w", err)
	}
	defer rows.Close()
	result := make([]RejectedRowDiagnostic, 0, limit+1)
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

func appendSQLKeysetComponents(components []sqlValueExpression, args *[]any) (string, error) {
	prefix := []string{}
	disjunction := []string{}
	for _, component := range components {
		placeholder := "$" + fmt.Sprint(len(*args)+1) + "::" + component.cast
		*args = append(*args, component.value)
		equal := component.expression + " IS NOT DISTINCT FROM " + placeholder
		if component.value != nil {
			after := component.expression + " > " + placeholder
			if component.nullable {
				after = "(" + after + " OR " + component.expression + " IS NULL)"
			}
			if len(prefix) > 0 {
				after = "(" + strings.Join(prefix, " AND ") + " AND " + after + ")"
			}
			disjunction = append(disjunction, after)
		}
		prefix = append(prefix, equal)
	}
	if len(disjunction) == 0 {
		return "FALSE", nil
	}
	return "(" + strings.Join(disjunction, " OR ") + ")", nil
}

func pointerValue[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
