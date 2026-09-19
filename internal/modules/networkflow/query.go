package networkflow

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const (
	schemaTableQueryRequest             = "cartulary.network_flow.table_query_request.v1"
	schemaTableQueryContinuation        = "cartulary.network_flow.table_query_continuation.v1"
	schemaRowsQueryRequest              = "cartulary.network_flow.rows_query_request.v1"
	schemaRowsQueryContinuation         = "cartulary.network_flow.rows_query_continuation.v1"
	schemaRejectedRowsQueryRequest      = "cartulary.network_flow.rejected_rows_query_request.v1"
	schemaRejectedRowsQueryContinuation = "cartulary.network_flow.rejected_rows_query_continuation.v1"
)

type queryFilter struct {
	FieldKey string `json:"field_key"`
	Op       string `json:"op"`
	Value    any    `json:"value,omitempty"`
}

type sortSpec struct {
	FieldKey  string `json:"field_key"`
	Direction string `json:"direction"`
}

type rowCursorPosition struct {
	EffectiveSort      []sortSpec `json:"effective_sort"`
	Values             []any      `json:"values"`
	NetworkFlowTableID string     `json:"network_flow_table_id"`
	NetworkFlowRowID   string     `json:"network_flow_row_id"`
}

type diagnosticCursorPosition struct {
	SourceRowNumber     int64   `json:"source_row_number"`
	SourceColumnOrdinal *int64  `json:"source_column_ordinal"`
	FieldKey            *string `json:"field_key"`
	ErrorCode           string  `json:"error_code"`
	ReasonCode          string  `json:"reason_code"`
	DiagnosticID        string  `json:"diagnostic_id"`
}

type contributorCursorPosition struct {
	WorkspaceTableOrder int               `json:"workspace_table_order"`
	Row                 rowCursorPosition `json:"row"`
}

type tableScope struct {
	Mode             string
	ActiveTableID    string
	SelectedTableIDs []string
}

type rowQueryRequest struct {
	SchemaID     string
	Continuation bool
	CursorToken  string
	TableScope   tableScope
	Filters      []queryFilter
	Sort         []sortSpec
	Limit        int
}

type rejectedRowsQueryRequest struct {
	SchemaID     string
	Continuation bool
	CursorToken  string
	ErrorCodes   []string
	FieldKeys    []string
	SourceRowGTE *int64
	SourceRowLTE *int64
	Limit        int
}

func decodeAcceptedRowQueryRequest(reader io.Reader, expectedSchemaID string, continuationSchemaID string, limits EffectiveLimits) (rowQueryRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return rowQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONString(raw, "schema_id")
	if apiErr != nil {
		return rowQueryRequest{}, apiErr
	}
	switch schemaID {
	case continuationSchemaID:
		if apiErr := ensureAllowedMembers(raw, "schema_id", "cursor_token"); apiErr != nil {
			return rowQueryRequest{}, apiErr
		}
		token, apiErr := requiredJSONString(raw, "cursor_token")
		if apiErr != nil {
			return rowQueryRequest{}, apiErr
		}
		return rowQueryRequest{SchemaID: schemaID, Continuation: true, CursorToken: token}, nil
	case expectedSchemaID:
	default:
		return rowQueryRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	allowed := []string{"schema_id", "filters", "sort", "limit"}
	if expectedSchemaID == schemaRowsQueryRequest {
		allowed = append(allowed, "table_scope")
	}
	if apiErr := ensureAllowedMembers(raw, allowed...); apiErr != nil {
		return rowQueryRequest{}, apiErr
	}
	request := rowQueryRequest{SchemaID: schemaID, Limit: defaultQueryLimit(limits)}
	if expectedSchemaID == schemaRowsQueryRequest {
		scope, apiErr := requiredTableScope(raw["table_scope"], limits)
		if apiErr != nil {
			return rowQueryRequest{}, apiErr
		}
		request.TableScope = scope
	}
	if value, ok := raw["limit"]; ok {
		limit, apiErr := decodePositiveInt(value, "limit")
		if apiErr != nil {
			return rowQueryRequest{}, invalidLimit("limit", "not_integer")
		}
		if limit < 1 {
			return rowQueryRequest{}, invalidLimit("limit", "below_minimum")
		}
		if int64(limit) > limits.MaxQueryLimit {
			return rowQueryRequest{}, invalidLimit("limit", "above_maximum")
		}
		request.Limit = limit
	}
	filters, apiErr := decodeFilters(raw["filters"], limits)
	if apiErr != nil {
		return rowQueryRequest{}, apiErr
	}
	sortSpecs, apiErr := decodeSort(raw["sort"], limits)
	if apiErr != nil {
		return rowQueryRequest{}, apiErr
	}
	request.Filters = filters
	request.Sort = sortSpecs
	return request, nil
}

func decodeRejectedRowsQueryRequest(reader io.Reader, limits EffectiveLimits) (rejectedRowsQueryRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return rejectedRowsQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONString(raw, "schema_id")
	if apiErr != nil {
		return rejectedRowsQueryRequest{}, apiErr
	}
	if schemaID == schemaRejectedRowsQueryContinuation {
		if apiErr := ensureAllowedMembers(raw, "schema_id", "cursor_token"); apiErr != nil {
			return rejectedRowsQueryRequest{}, apiErr
		}
		token, apiErr := requiredJSONString(raw, "cursor_token")
		if apiErr != nil {
			return rejectedRowsQueryRequest{}, apiErr
		}
		return rejectedRowsQueryRequest{SchemaID: schemaID, Continuation: true, CursorToken: token}, nil
	}
	if schemaID != schemaRejectedRowsQueryRequest {
		return rejectedRowsQueryRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "error_codes", "field_keys", "source_row_range", "limit"); apiErr != nil {
		return rejectedRowsQueryRequest{}, apiErr
	}
	request := rejectedRowsQueryRequest{SchemaID: schemaID, Limit: defaultQueryLimit(limits)}
	var err *semanticFailure
	request.ErrorCodes, err = decodeDiagnosticTokens(raw["error_codes"], "error_codes")
	if err != nil {
		return rejectedRowsQueryRequest{}, err
	}
	request.FieldKeys, err = decodeDiagnosticTokens(raw["field_keys"], "field_keys")
	if err != nil {
		return rejectedRowsQueryRequest{}, err
	}
	if value, ok := raw["source_row_range"]; ok {
		gte, lte, apiErr := decodeIntegerRange(value)
		if apiErr != nil {
			return rejectedRowsQueryRequest{}, apiErr
		}
		request.SourceRowGTE = gte
		request.SourceRowLTE = lte
	}
	if value, ok := raw["limit"]; ok {
		limit, apiErr := decodePositiveInt(value, "limit")
		if apiErr != nil {
			return rejectedRowsQueryRequest{}, invalidLimit("limit", "not_integer")
		}
		if limit < 1 {
			return rejectedRowsQueryRequest{}, invalidLimit("limit", "below_minimum")
		}
		if int64(limit) > limits.MaxQueryLimit {
			return rejectedRowsQueryRequest{}, invalidLimit("limit", "above_maximum")
		}
		request.Limit = limit
	}
	return request, nil
}

func decodeNetworkFlowObject(reader io.Reader) (map[string]json.RawMessage, *semanticFailure) {
	raw, err := strictjson.DecodeObject(reader)
	if err == nil {
		return raw, nil
	}
	switch {
	case errors.Is(err, strictjson.ErrDuplicateMember):
		return nil, invalidNetworkFlowRequest("", "duplicate_member")
	case errors.Is(err, strictjson.ErrMalformed), errors.Is(err, strictjson.ErrTrailingData):
		return nil, invalidNetworkFlowRequest("", "malformed_json")
	case errors.Is(err, strictjson.ErrNotObject):
		return nil, invalidNetworkFlowRequest("", "body_not_object")
	default:
		return nil, invalidNetworkFlowRequest("", "malformed_json")
	}
}

func requiredJSONString(raw map[string]json.RawMessage, field string) (string, *semanticFailure) {
	value, ok := raw[field]
	if !ok {
		return "", invalidNetworkFlowRequest(field, "missing_member")
	}
	if bytes.Equal(value, []byte("null")) {
		return "", invalidNetworkFlowRequest(field, "explicit_null")
	}
	var out string
	if err := json.Unmarshal(value, &out); err != nil || out == "" {
		return "", invalidNetworkFlowRequest(field, "type_mismatch")
	}
	return out, nil
}

func ensureAllowedMembers(raw map[string]json.RawMessage, allowed ...string) *semanticFailure {
	allowedSet := map[string]struct{}{}
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range raw {
		if _, ok := allowedSet[key]; !ok {
			return invalidNetworkFlowRequest(key, "unknown_member")
		}
	}
	return nil
}

func requiredTableScope(raw json.RawMessage, limits EffectiveLimits) (tableScope, *semanticFailure) {
	if len(raw) == 0 {
		return tableScope{}, invalidNetworkFlowRequest("table_scope", "missing_member")
	}
	if bytes.Equal(raw, []byte("null")) {
		return tableScope{}, invalidNetworkFlowRequest("table_scope", "explicit_null")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return tableScope{}, invalidTableScope("table_scope", "unknown_mode")
	}
	mode, apiErr := requiredJSONString(object, "mode")
	if apiErr != nil {
		return tableScope{}, invalidTableScope("mode", "unknown_mode")
	}
	switch mode {
	case "active_table":
		if apiErr := ensureAllowedMembers(object, "mode", "active_table_id"); apiErr != nil {
			return tableScope{}, invalidTableScope("table_scope", "variant_member_conflict")
		}
		tableID, apiErr := requiredJSONString(object, "active_table_id")
		if apiErr != nil {
			return tableScope{}, invalidTableScope("active_table_id", "empty_resolved_scope")
		}
		return tableScope{Mode: mode, ActiveTableID: tableID}, nil
	case "selected_tables":
		if apiErr := ensureAllowedMembers(object, "mode", "selected_table_ids"); apiErr != nil {
			return tableScope{}, invalidTableScope("table_scope", "variant_member_conflict")
		}
		tableIDs, apiErr := decodeStringArray(object["selected_table_ids"], "selected_table_ids", int(limits.MaxSelectedTablesPerQuery))
		if apiErr != nil {
			return tableScope{}, invalidTableScope("selected_table_ids", "empty_resolved_scope")
		}
		if len(tableIDs) > int(limits.MaxSelectedTablesPerQuery) {
			return tableScope{}, invalidTableScope("selected_table_ids", "selected_table_limit_exceeded")
		}
		return tableScope{Mode: mode, SelectedTableIDs: tableIDs}, nil
	case "all_active_tables":
		if apiErr := ensureAllowedMembers(object, "mode"); apiErr != nil {
			return tableScope{}, invalidTableScope("table_scope", "variant_member_conflict")
		}
		return tableScope{Mode: mode}, nil
	default:
		return tableScope{}, invalidTableScope("mode", "unknown_mode")
	}
}

func decodeFilters(raw json.RawMessage, limits EffectiveLimits) ([]queryFilter, *semanticFailure) {
	return decodeAndNormalizeFilters(raw, limits)
}

func decodeSort(raw json.RawMessage, limits EffectiveLimits) ([]sortSpec, *semanticFailure) {
	if len(raw) == 0 {
		return nil, nil
	}
	var encodedSpecs []json.RawMessage
	if err := json.Unmarshal(raw, &encodedSpecs); err != nil || encodedSpecs == nil {
		return nil, invalidSort("sort", "invalid_direction")
	}
	if int64(len(encodedSpecs)) > limits.MaxSortsPerQuery {
		return nil, invalidSort("sort", "too_many_sorts")
	}
	specs := make([]sortSpec, 0, len(encodedSpecs))
	seen := map[string]struct{}{}
	for _, encoded := range encodedSpecs {
		object, err := strictjson.DecodeObject(bytes.NewReader(encoded))
		if err != nil || len(object) != 2 {
			return nil, invalidSort("sort", "invalid_direction")
		}
		for member := range object {
			if member != "field_key" && member != "direction" {
				return nil, invalidSort("sort", "invalid_direction")
			}
		}
		fieldKey, fieldErr := strictSortString(object, "field_key")
		if fieldErr != nil {
			return nil, invalidSort("field_key", "unknown_field")
		}
		direction, directionErr := strictSortString(object, "direction")
		if directionErr != nil {
			return nil, invalidSort("direction", "invalid_direction")
		}
		spec := sortSpec{FieldKey: fieldKey, Direction: direction}
		if !isSortField(spec.FieldKey) {
			return nil, invalidSort("field_key", "unknown_field")
		}
		if spec.Direction != "asc" && spec.Direction != "desc" {
			return nil, invalidSort("direction", "invalid_direction")
		}
		if _, exists := seen[spec.FieldKey]; exists {
			return nil, invalidSort("field_key", "duplicate_sort_field")
		}
		seen[spec.FieldKey] = struct{}{}
		specs = append(specs, spec)
	}
	return specs, nil
}

func strictSortString(object map[string]json.RawMessage, member string) (string, error) {
	raw, ok := object[member]
	if !ok || bytes.Equal(raw, []byte("null")) {
		return "", errors.New("missing sort member")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || value == "" {
		return "", errors.New("invalid sort member")
	}
	return value, nil
}

func sortRows(rows []flowRow, specs []sortSpec) []flowRow {
	effective := effectiveSort(specs)
	sorted := append([]flowRow(nil), rows...)
	sort.SliceStable(sorted, func(i, j int) bool {
		for _, spec := range effective {
			cmp := compareRowFieldForSort(sorted[i], sorted[j], spec)
			if cmp == 0 {
				continue
			}
			return cmp < 0
		}
		if sorted[i].NetworkFlowTableID != sorted[j].NetworkFlowTableID {
			return sorted[i].NetworkFlowTableID < sorted[j].NetworkFlowTableID
		}
		return sorted[i].RowID < sorted[j].RowID
	})
	return sorted
}

func effectiveSort(specs []sortSpec) []sortSpec {
	result := append([]sortSpec(nil), specs...)
	seen := make(map[string]struct{}, len(result))
	for _, spec := range result {
		seen[spec.FieldKey] = struct{}{}
	}
	defaults := []sortSpec{
		{FieldKey: fieldFlowStartUTC, Direction: "asc"},
		{FieldKey: fieldFlowEndUTC, Direction: "asc"},
		{FieldKey: "source_row_number", Direction: "asc"},
		{FieldKey: "network_flow_row_id", Direction: "asc"},
	}
	for _, spec := range defaults {
		if _, exists := seen[spec.FieldKey]; exists {
			continue
		}
		result = append(result, spec)
	}
	return result
}

func sameSortSpecs(left, right []sortSpec) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func rowMatchesFilter(row flowRow, filter queryFilter) (bool, *semanticFailure) {
	value := rowPublicFieldValue(row, filter.FieldKey)
	switch filter.Op {
	case "is_null":
		return value == nil, nil
	case "not_null":
		return value != nil, nil
	case "eq":
		return value != nil && compareFilterValues(filter.FieldKey, value, filter.Value) == 0, nil
	case "in":
		values, ok := filter.Value.([]any)
		if !ok || len(values) == 0 {
			return false, invalidFilter("value", "invalid_value")
		}
		seen := map[string]struct{}{}
		for _, candidate := range values {
			key := string(canonicalJSON(candidate))
			if _, exists := seen[key]; exists {
				return false, invalidFilter("value", "duplicate_in_value")
			}
			seen[key] = struct{}{}
			if value != nil && compareFilterValues(filter.FieldKey, value, candidate) == 0 {
				return true, nil
			}
		}
		return false, nil
	case "range":
		object, ok := filter.Value.(map[string]any)
		if !ok {
			return false, invalidFilter("value", "invalid_value")
		}
		if value == nil {
			return false, nil
		}
		gte, hasGTE := object["gte"]
		upperKey := "lte"
		if filter.FieldKey == fieldFlowStartUTC || filter.FieldKey == fieldFlowEndUTC {
			upperKey = "lt"
		}
		upper, hasUpper := object[upperKey]
		hasGTE = hasGTE && gte != nil
		hasUpper = hasUpper && upper != nil
		if !hasGTE && !hasUpper {
			return false, invalidFilter("value", "empty_range")
		}
		if hasGTE && compareFilterValues(filter.FieldKey, value, gte) < 0 {
			return false, nil
		}
		if hasUpper {
			comparison := compareFilterValues(filter.FieldKey, value, upper)
			if comparison > 0 || upperKey == "lt" && comparison == 0 {
				return false, nil
			}
		}
		return true, nil
	case "prefix":
		text, ok := value.(string)
		prefix, ok2 := filter.Value.(string)
		return ok && ok2 && strings.HasPrefix(text, prefix), nil
	case "contains":
		text, ok := value.(string)
		needle, ok2 := filter.Value.(string)
		return ok && ok2 && strings.Contains(text, needle), nil
	case "cidr_contains":
		cidr, ok := filter.Value.(string)
		if !ok {
			return false, invalidFilter("value", "invalid_value")
		}
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return false, invalidFilter("value", "invalid_value")
		}
		if filter.FieldKey == fieldEndpointIP {
			src, srcOK := parseIP(row.SrcIP)
			dst, dstOK := parseIP(row.DstIP)
			return (srcOK && prefix.Contains(src)) || (dstOK && prefix.Contains(dst)), nil
		}
		text, _ := value.(string)
		addr, ok := parseIP(text)
		return ok && prefix.Contains(addr), nil
	default:
		return false, invalidFilter("op", "operator_not_allowed")
	}
}

func rowPublicFieldValue(row flowRow, field string) any {
	switch field {
	case fieldSrcIP:
		return row.SrcIP
	case fieldDstIP:
		return row.DstIP
	case fieldSrcPort:
		return nullableInt32Value(row.SrcPort)
	case fieldDstPort:
		return nullableInt32Value(row.DstPort)
	case fieldIPProtocol:
		return int64(row.IPProtocol)
	case fieldFlowStartUTC:
		return row.FlowStartUTC.UTC().Format(time.RFC3339Nano)
	case fieldFlowEndUTC:
		return row.FlowEndUTC.UTC().Format(time.RFC3339Nano)
	case fieldBytesCount:
		return row.BytesCount
	case fieldPacketsCount:
		return row.PacketsCount
	case fieldExporterID:
		return nullableStringValue(row.ExporterID)
	case fieldInputInterface:
		return nullableStringValue(row.InputInterface)
	case fieldOutputInterface:
		return nullableStringValue(row.OutputInterface)
	case "source_row_number":
		return row.SourceRowNumber
	case "network_flow_row_id":
		return row.RowID
	case "network_flow_table_id":
		return row.NetworkFlowTableID
	default:
		return nil
	}
}

func compareRowField(a, b flowRow, field string) int {
	left := rowPublicFieldValue(a, field)
	right := rowPublicFieldValue(b, field)
	if field == fieldSrcIP || field == fieldDstIP {
		return compareIPValues(left, right)
	}
	return compareFilterValues(field, left, right)
}

func compareRowFieldForSort(a, b flowRow, spec sortSpec) int {
	left := rowPublicFieldValue(a, spec.FieldKey)
	right := rowPublicFieldValue(b, spec.FieldKey)
	if left == nil || right == nil {
		switch {
		case left == nil && right == nil:
			return 0
		case left == nil:
			return 1
		default:
			return -1
		}
	}
	cmp := compareRowField(a, b, spec.FieldKey)
	if spec.Direction == "desc" {
		return -cmp
	}
	return cmp
}

func compareIPValues(left any, right any) int {
	leftText, leftOK := left.(string)
	rightText, rightOK := right.(string)
	if !leftOK || !rightOK {
		return compareScalar(left, right)
	}
	leftIP, leftOK := parseIP(leftText)
	rightIP, rightOK := parseIP(rightText)
	if !leftOK || !rightOK {
		return compareScalar(left, right)
	}
	if leftIP.Is4() != rightIP.Is4() {
		if leftIP.Is4() {
			return -1
		}
		return 1
	}
	return leftIP.Compare(rightIP)
}

func compareScalar(left any, right any) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return -1
	}
	if right == nil {
		return 1
	}
	if l, ok := numericString(left); ok {
		if r, ok := numericString(right); ok {
			return l.Cmp(r)
		}
	}
	ls := fmt.Sprint(left)
	rs := fmt.Sprint(right)
	if t1, ok := parseTimeString(ls); ok {
		if t2, ok := parseTimeString(rs); ok {
			return t1.Compare(t2)
		}
	}
	if ls < rs {
		return -1
	}
	if ls > rs {
		return 1
	}
	return 0
}

func numericString(value any) (*big.Int, bool) {
	switch v := value.(type) {
	case int:
		return big.NewInt(int64(v)), true
	case int32:
		return big.NewInt(int64(v)), true
	case int64:
		return big.NewInt(v), true
	case float64:
		if v != float64(int64(v)) {
			return nil, false
		}
		return big.NewInt(int64(v)), true
	case json.Number:
		i, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return nil, false
		}
		return big.NewInt(i), true
	case string:
		if v == "" {
			return nil, false
		}
		n := new(big.Int)
		if _, ok := n.SetString(v, 10); ok {
			return n, true
		}
	}
	return nil, false
}

func parseTimeString(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func parseIP(value string) (netip.Addr, bool) {
	addr, err := netip.ParseAddr(value)
	return addr, err == nil
}

func nullableInt32Value(value *int32) any {
	if value == nil {
		return nil
	}
	return int64(*value)
}

func nullableStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func filterOpAllowed(filter queryFilter) bool {
	switch filter.FieldKey {
	case fieldSrcIP, fieldDstIP, fieldEndpointIP:
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "cidr_contains"
	case fieldSrcPort, fieldDstPort:
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "range" || filter.Op == "is_null" || filter.Op == "not_null"
	case fieldIPProtocol, "source_row_number":
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "range"
	case fieldFlowStartUTC, fieldFlowEndUTC:
		return filter.Op == "range"
	case fieldBytesCount, fieldPacketsCount:
		return filter.Op == "eq" || filter.Op == "range"
	case fieldExporterID, fieldInputInterface, fieldOutputInterface:
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "prefix" || filter.Op == "contains" || filter.Op == "is_null" || filter.Op == "not_null"
	default:
		return false
	}
}

const fieldEndpointIP = "network_flow.endpoint_ip"

func isFilterField(field string) bool {
	switch field {
	case fieldSrcIP, fieldDstIP, fieldEndpointIP, fieldSrcPort, fieldDstPort, fieldIPProtocol, fieldFlowStartUTC, fieldFlowEndUTC, fieldBytesCount, fieldPacketsCount, fieldExporterID, fieldInputInterface, fieldOutputInterface, "source_row_number":
		return true
	default:
		return false
	}
}

func isSortField(field string) bool {
	switch field {
	case fieldSrcIP, fieldDstIP, fieldSrcPort, fieldDstPort, fieldIPProtocol, fieldFlowStartUTC, fieldFlowEndUTC, fieldBytesCount, fieldPacketsCount, fieldExporterID, fieldInputInterface, fieldOutputInterface, "source_row_number", "network_flow_row_id", "network_flow_table_id":
		return true
	default:
		return false
	}
}

func newRowCursorPosition(row flowRow, specs []sortSpec) rowCursorPosition {
	effective := effectiveSort(specs)
	values := make([]any, 0, len(effective))
	for _, spec := range effective {
		values = append(values, rowPublicFieldValue(row, spec.FieldKey))
	}
	return rowCursorPosition{
		EffectiveSort:      effective,
		Values:             values,
		NetworkFlowTableID: row.NetworkFlowTableID,
		NetworkFlowRowID:   row.RowID,
	}
}

func decodeRowCursorPosition(raw json.RawMessage) (rowCursorPosition, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var position rowCursorPosition
	if err := decoder.Decode(&position); err != nil {
		return rowCursorPosition{}, err
	}
	if len(position.EffectiveSort) == 0 || len(position.Values) != len(position.EffectiveSort) || position.NetworkFlowTableID == "" || position.NetworkFlowRowID == "" {
		return rowCursorPosition{}, errors.New("invalid row cursor position")
	}
	return position, nil
}

func compareRowToPosition(row flowRow, position rowCursorPosition) int {
	for index, spec := range position.EffectiveSort {
		left := rowPublicFieldValue(row, spec.FieldKey)
		right := position.Values[index]
		if left == nil || right == nil {
			var cmp int
			switch {
			case left == nil && right == nil:
				cmp = 0
			case left == nil:
				cmp = 1
			default:
				cmp = -1
			}
			if cmp != 0 {
				return cmp
			}
			continue
		}
		cmp := compareFilterValues(spec.FieldKey, left, right)
		if spec.FieldKey == fieldSrcIP || spec.FieldKey == fieldDstIP {
			cmp = compareIPValues(left, right)
		}
		if spec.Direction == "desc" {
			cmp = -cmp
		}
		if cmp != 0 {
			return cmp
		}
	}
	if row.NetworkFlowTableID < position.NetworkFlowTableID {
		return -1
	}
	if row.NetworkFlowTableID > position.NetworkFlowTableID {
		return 1
	}
	if row.RowID < position.NetworkFlowRowID {
		return -1
	}
	if row.RowID > position.NetworkFlowRowID {
		return 1
	}
	return 0
}

func pageFlowRowsAfter(rows []flowRow, position *rowCursorPosition, limit int) ([]flowRow, bool) {
	start := 0
	if position != nil {
		start = sort.Search(len(rows), func(index int) bool {
			return compareRowToPosition(rows[index], *position) > 0
		})
	}
	if start >= len(rows) {
		return []flowRow{}, false
	}
	end := start + limit
	if end >= len(rows) {
		return rows[start:], false
	}
	return rows[start:end], true
}

func newContributorCursorPosition(row flowRow, tableRanks map[string]int) contributorCursorPosition {
	return contributorCursorPosition{WorkspaceTableOrder: tableRanks[row.NetworkFlowTableID], Row: newRowCursorPosition(row, nil)}
}

func decodeContributorCursorPosition(raw json.RawMessage) (contributorCursorPosition, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var position contributorCursorPosition
	if err := decoder.Decode(&position); err != nil {
		return contributorCursorPosition{}, err
	}
	decodedRow, err := decodeRowCursorPosition(mustMarshalJSON(position.Row))
	if err != nil || position.WorkspaceTableOrder < 0 {
		return contributorCursorPosition{}, errors.New("invalid contributor cursor position")
	}
	position.Row = decodedRow
	return position, nil
}

func mustMarshalJSON(value any) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}

func newDiagnosticCursorPosition(value rejectedRowDiagnostic) diagnosticCursorPosition {
	return diagnosticCursorPosition{
		SourceRowNumber: value.SourceRowNumber, SourceColumnOrdinal: value.SourceColumnOrdinal,
		FieldKey: value.FieldKey, ErrorCode: value.ErrorCode, ReasonCode: value.ReasonCode,
		DiagnosticID: value.DiagnosticID,
	}
}

func pageDiagnosticsAfter(rows []rejectedRowDiagnostic, position *diagnosticCursorPosition, limit int) ([]rejectedRowDiagnostic, bool) {
	start := 0
	if position != nil {
		needle := rejectedRowDiagnostic{
			SourceRowNumber: position.SourceRowNumber, SourceColumnOrdinal: position.SourceColumnOrdinal,
			FieldKey: position.FieldKey, ErrorCode: position.ErrorCode, ReasonCode: position.ReasonCode,
			DiagnosticID: position.DiagnosticID,
		}
		start = sort.Search(len(rows), func(index int) bool { return compareDiagnostics(rows[index], needle) > 0 })
	}
	if start >= len(rows) {
		return []rejectedRowDiagnostic{}, false
	}
	end := start + limit
	if end >= len(rows) {
		return rows[start:], false
	}
	return rows[start:end], true
}

func queryHash(value any) string {
	sum := sha256.Sum256(canonicalJSON(value))
	return hex.EncodeToString(sum[:])
}

func decodePositiveInt(raw json.RawMessage, field string) (int, *semanticFailure) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return 0, invalidNetworkFlowRequest(field, "type_mismatch")
	}
	number, ok := value.(json.Number)
	if !ok {
		return 0, invalidNetworkFlowRequest(field, "type_mismatch")
	}
	parsed, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil {
		return 0, invalidNetworkFlowRequest(field, "type_mismatch")
	}
	return int(parsed), nil
}

func decodeStringArray(raw json.RawMessage, field string, max int) ([]string, *semanticFailure) {
	if len(raw) == 0 {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, invalidNetworkFlowRequest(field, "type_mismatch")
	}
	if len(values) > max {
		return nil, invalidNetworkFlowRequest(field, "invalid_schema_id")
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			return nil, invalidNetworkFlowRequest(field, "type_mismatch")
		}
		if _, exists := seen[value]; exists {
			return nil, invalidFilter(field, "duplicate_in_value")
		}
		seen[value] = struct{}{}
	}
	sort.Strings(values)
	return values, nil
}

func decodeIntegerRange(raw json.RawMessage) (*int64, *int64, *semanticFailure) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil, invalidFilter("source_row_range", "invalid_value")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, nil, invalidFilter("source_row_range", "invalid_value")
	}
	if apiErr := ensureAllowedMembers(object, "gte", "lte"); apiErr != nil {
		return nil, nil, invalidFilter("source_row_range", "invalid_value")
	}
	var gte *int64
	var lte *int64
	if value, ok := object["gte"]; ok && !bytes.Equal(value, []byte("null")) {
		parsed, apiErr := decodePositiveInt(value, "gte")
		if apiErr != nil {
			return nil, nil, invalidFilter("source_row_range", "invalid_value")
		}
		v := int64(parsed)
		gte = &v
	}
	if value, ok := object["lte"]; ok && !bytes.Equal(value, []byte("null")) {
		parsed, apiErr := decodePositiveInt(value, "lte")
		if apiErr != nil {
			return nil, nil, invalidFilter("source_row_range", "invalid_value")
		}
		v := int64(parsed)
		lte = &v
	}
	if (gte == nil && lte == nil) || (gte != nil && lte != nil && *gte > *lte) {
		return nil, nil, invalidFilter("source_row_range", "empty_range")
	}
	return gte, lte, nil
}

func defaultQueryLimit(limits EffectiveLimits) int {
	if limits.MaxQueryLimit <= 0 || limits.MaxQueryLimit > defaultMaxQueryLimit {
		return 200
	}
	if limits.MaxQueryLimit < 200 {
		return int(limits.MaxQueryLimit)
	}
	return 200
}

func invalidNetworkFlowRequest(field string, reason string) *semanticFailure {
	return newSemanticFailure(failureInvalidRequest, field, reason)
}

func invalidFilter(field string, reason string) *semanticFailure {
	if reason == "value_forbidden" || reason == "unknown_member" {
		reason = "invalid_value"
	}
	return newSemanticFailure(failureInvalidFilter, "", reason)
}

func invalidSort(field string, reason string) *semanticFailure {
	return newSemanticFailure(failureInvalidSort, field, reason)
}

func invalidTableScope(field string, reason string) *semanticFailure {
	return newSemanticFailure(failureInvalidTableScope, field, reason)
}

func invalidLimit(field string, reason string) *semanticFailure {
	return newSemanticFailure(failureInvalidLimit, field, reason)
}

func cursorInvalid(reason string) *semanticFailure {
	return newSemanticFailure(failureCursorInvalid, "", reason)
}
