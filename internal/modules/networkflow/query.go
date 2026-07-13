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

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	schemaTableQueryRequest             = "cartulary.network_flow.table_query_request.v1"
	schemaTableQueryContinuation        = "cartulary.network_flow.table_query_continuation.v1"
	schemaRowsQueryRequest              = "cartulary.network_flow.rows_query_request.v1"
	schemaRowsQueryContinuation         = "cartulary.network_flow.rows_query_continuation.v1"
	schemaRejectedRowsQueryRequest      = "cartulary.network_flow.rejected_rows_query_request.v1"
	schemaRejectedRowsQueryContinuation = "cartulary.network_flow.rejected_rows_query_continuation.v1"
)

type Filter struct {
	FieldKey string `json:"field_key"`
	Op       string `json:"op"`
	Value    any    `json:"value,omitempty"`
}

type SortSpec struct {
	FieldKey  string `json:"field_key"`
	Direction string `json:"direction"`
}

type rowCursorPosition struct {
	EffectiveSort      []SortSpec `json:"effective_sort"`
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

type TableScope struct {
	Mode             string
	ActiveTableID    string
	SelectedTableIDs []string
}

type RowQueryRequest struct {
	SchemaID     string
	Continuation bool
	CursorToken  string
	TableScope   TableScope
	Filters      []Filter
	Sort         []SortSpec
	Limit        int
}

type RejectedRowsQueryRequest struct {
	SchemaID     string
	Continuation bool
	CursorToken  string
	ErrorCodes   []string
	FieldKeys    []string
	SourceRowGTE *int64
	SourceRowLTE *int64
	Limit        int
}

func decodeAcceptedRowQueryRequest(reader io.Reader, expectedSchemaID string, continuationSchemaID string, limits Limits) (RowQueryRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return RowQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONString(raw, "schema_id")
	if apiErr != nil {
		return RowQueryRequest{}, apiErr
	}
	switch schemaID {
	case continuationSchemaID:
		if apiErr := ensureAllowedMembers(raw, "schema_id", "cursor_token"); apiErr != nil {
			return RowQueryRequest{}, apiErr
		}
		token, apiErr := requiredJSONString(raw, "cursor_token")
		if apiErr != nil {
			return RowQueryRequest{}, apiErr
		}
		return RowQueryRequest{SchemaID: schemaID, Continuation: true, CursorToken: token}, nil
	case expectedSchemaID:
	default:
		return RowQueryRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	allowed := []string{"schema_id", "filters", "sort", "limit"}
	if expectedSchemaID == schemaRowsQueryRequest {
		allowed = append(allowed, "table_scope")
	}
	if apiErr := ensureAllowedMembers(raw, allowed...); apiErr != nil {
		return RowQueryRequest{}, apiErr
	}
	request := RowQueryRequest{SchemaID: schemaID, Limit: defaultQueryLimit(limits)}
	if expectedSchemaID == schemaRowsQueryRequest {
		scope, apiErr := requiredTableScope(raw["table_scope"], limits)
		if apiErr != nil {
			return RowQueryRequest{}, apiErr
		}
		request.TableScope = scope
	}
	if value, ok := raw["limit"]; ok {
		limit, apiErr := decodePositiveInt(value, "limit")
		if apiErr != nil {
			return RowQueryRequest{}, invalidLimit("limit", "not_integer")
		}
		if limit < 1 {
			return RowQueryRequest{}, invalidLimit("limit", "below_minimum")
		}
		if int64(limit) > limits.MaxQueryLimit {
			return RowQueryRequest{}, invalidLimit("limit", "above_maximum")
		}
		request.Limit = limit
	}
	filters, apiErr := decodeFilters(raw["filters"], limits)
	if apiErr != nil {
		return RowQueryRequest{}, apiErr
	}
	sortSpecs, apiErr := decodeSort(raw["sort"], limits)
	if apiErr != nil {
		return RowQueryRequest{}, apiErr
	}
	request.Filters = filters
	request.Sort = sortSpecs
	return request, nil
}

func decodeRejectedRowsQueryRequest(reader io.Reader, limits Limits) (RejectedRowsQueryRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return RejectedRowsQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONString(raw, "schema_id")
	if apiErr != nil {
		return RejectedRowsQueryRequest{}, apiErr
	}
	if schemaID == schemaRejectedRowsQueryContinuation {
		if apiErr := ensureAllowedMembers(raw, "schema_id", "cursor_token"); apiErr != nil {
			return RejectedRowsQueryRequest{}, apiErr
		}
		token, apiErr := requiredJSONString(raw, "cursor_token")
		if apiErr != nil {
			return RejectedRowsQueryRequest{}, apiErr
		}
		return RejectedRowsQueryRequest{SchemaID: schemaID, Continuation: true, CursorToken: token}, nil
	}
	if schemaID != schemaRejectedRowsQueryRequest {
		return RejectedRowsQueryRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "error_codes", "field_keys", "source_row_range", "limit"); apiErr != nil {
		return RejectedRowsQueryRequest{}, apiErr
	}
	request := RejectedRowsQueryRequest{SchemaID: schemaID, Limit: defaultQueryLimit(limits)}
	var err *httpapi.APIError
	request.ErrorCodes, err = decodeStringArray(raw["error_codes"], "error_codes", 64)
	if err != nil {
		return RejectedRowsQueryRequest{}, err
	}
	request.FieldKeys, err = decodeStringArray(raw["field_keys"], "field_keys", 64)
	if err != nil {
		return RejectedRowsQueryRequest{}, err
	}
	if value, ok := raw["source_row_range"]; ok {
		gte, lte, apiErr := decodeIntegerRange(value)
		if apiErr != nil {
			return RejectedRowsQueryRequest{}, apiErr
		}
		request.SourceRowGTE = gte
		request.SourceRowLTE = lte
	}
	if value, ok := raw["limit"]; ok {
		limit, apiErr := decodePositiveInt(value, "limit")
		if apiErr != nil {
			return RejectedRowsQueryRequest{}, invalidLimit("limit", "not_integer")
		}
		if limit < 1 {
			return RejectedRowsQueryRequest{}, invalidLimit("limit", "below_minimum")
		}
		if int64(limit) > limits.MaxQueryLimit {
			return RejectedRowsQueryRequest{}, invalidLimit("limit", "above_maximum")
		}
		request.Limit = limit
	}
	return request, nil
}

func decodeNetworkFlowObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err == nil {
		return raw, nil
	}
	switch {
	case errors.Is(err, httpapi.ErrStrictJSONDuplicateMember):
		return nil, invalidNetworkFlowRequest("", "duplicate_member")
	case errors.Is(err, httpapi.ErrStrictJSONMalformed), errors.Is(err, httpapi.ErrStrictJSONTrailingData):
		return nil, invalidNetworkFlowRequest("", "malformed_json")
	case errors.Is(err, httpapi.ErrStrictJSONNotObject):
		return nil, invalidNetworkFlowRequest("", "body_not_object")
	default:
		return nil, invalidNetworkFlowRequest("", "malformed_json")
	}
}

func requiredJSONString(raw map[string]json.RawMessage, field string) (string, *httpapi.APIError) {
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

func ensureAllowedMembers(raw map[string]json.RawMessage, allowed ...string) *httpapi.APIError {
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

func requiredTableScope(raw json.RawMessage, limits Limits) (TableScope, *httpapi.APIError) {
	if len(raw) == 0 {
		return TableScope{}, invalidNetworkFlowRequest("table_scope", "missing_member")
	}
	if bytes.Equal(raw, []byte("null")) {
		return TableScope{}, invalidNetworkFlowRequest("table_scope", "explicit_null")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return TableScope{}, invalidTableScope("table_scope", "unknown_mode")
	}
	mode, apiErr := requiredJSONString(object, "mode")
	if apiErr != nil {
		return TableScope{}, invalidTableScope("mode", "unknown_mode")
	}
	switch mode {
	case "active_table":
		if apiErr := ensureAllowedMembers(object, "mode", "active_table_id"); apiErr != nil {
			return TableScope{}, invalidTableScope("table_scope", "variant_member_conflict")
		}
		tableID, apiErr := requiredJSONString(object, "active_table_id")
		if apiErr != nil {
			return TableScope{}, invalidTableScope("active_table_id", "empty_resolved_scope")
		}
		return TableScope{Mode: mode, ActiveTableID: tableID}, nil
	case "selected_tables":
		if apiErr := ensureAllowedMembers(object, "mode", "selected_table_ids"); apiErr != nil {
			return TableScope{}, invalidTableScope("table_scope", "variant_member_conflict")
		}
		tableIDs, apiErr := decodeStringArray(object["selected_table_ids"], "selected_table_ids", int(limits.MaxSelectedTablesPerQuery))
		if apiErr != nil {
			return TableScope{}, invalidTableScope("selected_table_ids", "empty_resolved_scope")
		}
		if len(tableIDs) > int(limits.MaxSelectedTablesPerQuery) {
			return TableScope{}, invalidTableScope("selected_table_ids", "selected_table_limit_exceeded")
		}
		return TableScope{Mode: mode, SelectedTableIDs: tableIDs}, nil
	case "all_active_tables":
		if apiErr := ensureAllowedMembers(object, "mode"); apiErr != nil {
			return TableScope{}, invalidTableScope("table_scope", "variant_member_conflict")
		}
		return TableScope{Mode: mode}, nil
	default:
		return TableScope{}, invalidTableScope("mode", "unknown_mode")
	}
}

func decodeFilters(raw json.RawMessage, limits Limits) ([]Filter, *httpapi.APIError) {
	return decodeAndNormalizeFilters(raw, limits)
}

func decodeSort(raw json.RawMessage, limits Limits) ([]SortSpec, *httpapi.APIError) {
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
	specs := make([]SortSpec, 0, len(encodedSpecs))
	seen := map[string]struct{}{}
	for _, encoded := range encodedSpecs {
		object, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(encoded))
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
		spec := SortSpec{FieldKey: fieldKey, Direction: direction}
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

func filterRows(rows []FlowRow, filters []Filter) ([]FlowRow, *httpapi.APIError) {
	if len(filters) == 0 {
		return rows, nil
	}
	out := make([]FlowRow, 0, len(rows))
	for _, row := range rows {
		matched := true
		for _, filter := range filters {
			ok, apiErr := rowMatchesFilter(row, filter)
			if apiErr != nil {
				return nil, apiErr
			}
			if !ok {
				matched = false
				break
			}
		}
		if matched {
			out = append(out, row)
		}
	}
	return out, nil
}

func sortRows(rows []FlowRow, specs []SortSpec) []FlowRow {
	effective := effectiveSort(specs)
	sorted := append([]FlowRow(nil), rows...)
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

func effectiveSort(specs []SortSpec) []SortSpec {
	result := append([]SortSpec(nil), specs...)
	seen := make(map[string]struct{}, len(result))
	for _, spec := range result {
		seen[spec.FieldKey] = struct{}{}
	}
	defaults := []SortSpec{
		{FieldKey: FieldFlowStartUTC, Direction: "asc"},
		{FieldKey: FieldFlowEndUTC, Direction: "asc"},
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

func sameSortSpecs(left, right []SortSpec) bool {
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

func rowMatchesFilter(row FlowRow, filter Filter) (bool, *httpapi.APIError) {
	value := rowPublicFieldValue(row, filter.FieldKey)
	switch filter.Op {
	case "is_null":
		return value == nil, nil
	case "not_null":
		return value != nil, nil
	case "eq":
		return compareScalar(value, filter.Value) == 0, nil
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
			if compareScalar(value, candidate) == 0 {
				return true, nil
			}
		}
		return false, nil
	case "range":
		object, ok := filter.Value.(map[string]any)
		if !ok {
			return false, invalidFilter("value", "invalid_value")
		}
		gte, hasGTE := object["gte"]
		lte, hasLTE := object["lte"]
		if !hasGTE && !hasLTE {
			return false, invalidFilter("value", "empty_range")
		}
		if hasGTE && compareScalar(value, gte) < 0 {
			return false, nil
		}
		if hasLTE && compareScalar(value, lte) > 0 {
			return false, nil
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
		if filter.FieldKey == FieldEndpointIP {
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

func rowPublicFieldValue(row FlowRow, field string) any {
	switch field {
	case FieldSrcIP:
		return row.SrcIP
	case FieldDstIP:
		return row.DstIP
	case FieldSrcPort:
		return nullableInt32Value(row.SrcPort)
	case FieldDstPort:
		return nullableInt32Value(row.DstPort)
	case FieldIPProtocol:
		return int64(row.IPProtocol)
	case FieldFlowStartUTC:
		return row.FlowStartUTC.UTC().Format(time.RFC3339Nano)
	case FieldFlowEndUTC:
		return row.FlowEndUTC.UTC().Format(time.RFC3339Nano)
	case FieldBytesCount:
		return row.BytesCount
	case FieldPacketsCount:
		return row.PacketsCount
	case FieldExporterID:
		return nullableStringValue(row.ExporterID)
	case FieldInputInterface:
		return nullableStringValue(row.InputInterface)
	case FieldOutputInterface:
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

func compareRowField(a, b FlowRow, field string) int {
	left := rowPublicFieldValue(a, field)
	right := rowPublicFieldValue(b, field)
	if field == FieldSrcIP || field == FieldDstIP {
		return compareIPValues(left, right)
	}
	return compareScalar(left, right)
}

func compareRowFieldForSort(a, b FlowRow, spec SortSpec) int {
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

func filterOpAllowed(filter Filter) bool {
	switch filter.FieldKey {
	case FieldSrcIP, FieldDstIP, FieldEndpointIP:
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "cidr_contains"
	case FieldSrcPort, FieldDstPort:
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "range" || filter.Op == "is_null" || filter.Op == "not_null"
	case FieldIPProtocol, "source_row_number":
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "range"
	case FieldFlowStartUTC, FieldFlowEndUTC:
		return filter.Op == "range"
	case FieldBytesCount, FieldPacketsCount:
		return filter.Op == "eq" || filter.Op == "range"
	case FieldExporterID, FieldInputInterface, FieldOutputInterface:
		return filter.Op == "eq" || filter.Op == "in" || filter.Op == "prefix" || filter.Op == "contains" || filter.Op == "is_null" || filter.Op == "not_null"
	default:
		return false
	}
}

const FieldEndpointIP = "network_flow.endpoint_ip"

func isFilterField(field string) bool {
	switch field {
	case FieldSrcIP, FieldDstIP, FieldEndpointIP, FieldSrcPort, FieldDstPort, FieldIPProtocol, FieldFlowStartUTC, FieldFlowEndUTC, FieldBytesCount, FieldPacketsCount, FieldExporterID, FieldInputInterface, FieldOutputInterface, "source_row_number":
		return true
	default:
		return false
	}
}

func isSortField(field string) bool {
	switch field {
	case FieldSrcIP, FieldDstIP, FieldSrcPort, FieldDstPort, FieldIPProtocol, FieldFlowStartUTC, FieldFlowEndUTC, FieldBytesCount, FieldPacketsCount, FieldExporterID, FieldInputInterface, FieldOutputInterface, "source_row_number", "network_flow_row_id", "network_flow_table_id":
		return true
	default:
		return false
	}
}

func newRowCursorPosition(row FlowRow, specs []SortSpec) rowCursorPosition {
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

func compareRowToPosition(row FlowRow, position rowCursorPosition) int {
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
		cmp := compareScalar(left, right)
		if spec.FieldKey == FieldSrcIP || spec.FieldKey == FieldDstIP {
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

func pageFlowRowsAfter(rows []FlowRow, position *rowCursorPosition, limit int) ([]FlowRow, bool) {
	start := 0
	if position != nil {
		start = sort.Search(len(rows), func(index int) bool {
			return compareRowToPosition(rows[index], *position) > 0
		})
	}
	if start >= len(rows) {
		return []FlowRow{}, false
	}
	end := start + limit
	if end >= len(rows) {
		return rows[start:], false
	}
	return rows[start:end], true
}

func newContributorCursorPosition(row FlowRow, tableRanks map[string]int) contributorCursorPosition {
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

func pageContributorRowsAfter(rows []FlowRow, tableRanks map[string]int, position *contributorCursorPosition, limit int) ([]FlowRow, bool) {
	start := 0
	if position != nil {
		start = sort.Search(len(rows), func(index int) bool {
			rank := tableRanks[rows[index].NetworkFlowTableID]
			if rank != position.WorkspaceTableOrder {
				return rank > position.WorkspaceTableOrder
			}
			return compareRowToPosition(rows[index], position.Row) > 0
		})
	}
	if start >= len(rows) {
		return []FlowRow{}, false
	}
	end := start + limit
	if end >= len(rows) {
		return rows[start:], false
	}
	return rows[start:end], true
}

func mustMarshalJSON(value any) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}

func newDiagnosticCursorPosition(value RejectedRowDiagnostic) diagnosticCursorPosition {
	return diagnosticCursorPosition{
		SourceRowNumber: value.SourceRowNumber, SourceColumnOrdinal: value.SourceColumnOrdinal,
		FieldKey: value.FieldKey, ErrorCode: value.ErrorCode, ReasonCode: value.ReasonCode,
		DiagnosticID: value.DiagnosticID,
	}
}

func pageDiagnosticsAfter(rows []RejectedRowDiagnostic, position *diagnosticCursorPosition, limit int) ([]RejectedRowDiagnostic, bool) {
	start := 0
	if position != nil {
		needle := RejectedRowDiagnostic{
			SourceRowNumber: position.SourceRowNumber, SourceColumnOrdinal: position.SourceColumnOrdinal,
			FieldKey: position.FieldKey, ErrorCode: position.ErrorCode, ReasonCode: position.ReasonCode,
			DiagnosticID: position.DiagnosticID,
		}
		start = sort.Search(len(rows), func(index int) bool { return compareDiagnostics(rows[index], needle) > 0 })
	}
	if start >= len(rows) {
		return []RejectedRowDiagnostic{}, false
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

func decodePositiveInt(raw json.RawMessage, field string) (int, *httpapi.APIError) {
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

func decodeStringArray(raw json.RawMessage, field string, max int) ([]string, *httpapi.APIError) {
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

func decodeIntegerRange(raw json.RawMessage) (*int64, *int64, *httpapi.APIError) {
	if bytes.Equal(raw, []byte("null")) {
		return nil, nil, nil
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
	if gte != nil && lte != nil && *gte > *lte {
		return nil, nil, invalidFilter("source_row_range", "empty_range")
	}
	return gte, lte, nil
}

func defaultQueryLimit(limits Limits) int {
	if limits.MaxQueryLimit <= 0 || limits.MaxQueryLimit > DefaultMaxQueryLimit {
		return 100
	}
	if limits.MaxQueryLimit < 100 {
		return int(limits.MaxQueryLimit)
	}
	return 100
}

func invalidNetworkFlowRequest(field string, reason string) *httpapi.APIError {
	return networkFlowAPIError(400, "network_flow_invalid_request", field, reason)
}

func invalidFilter(field string, reason string) *httpapi.APIError {
	return networkFlowAPIError(400, "network_flow_invalid_filter", field, reason)
}

func invalidSort(field string, reason string) *httpapi.APIError {
	return networkFlowAPIError(400, "network_flow_invalid_sort", field, reason)
}

func invalidTableScope(field string, reason string) *httpapi.APIError {
	return networkFlowAPIError(400, "network_flow_invalid_table_scope", field, reason)
}

func invalidLimit(field string, reason string) *httpapi.APIError {
	return networkFlowAPIError(400, "network_flow_invalid_limit", field, reason)
}

func cursorInvalid(reason string) *httpapi.APIError {
	return networkFlowAPIError(400, "network_flow_cursor_invalid", "", reason)
}

func networkFlowAPIError(status int, code string, field string, reason string) *httpapi.APIError {
	details := map[string]any{"reason_code": reason}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: status, Code: code, Message: code, Details: details}
}
