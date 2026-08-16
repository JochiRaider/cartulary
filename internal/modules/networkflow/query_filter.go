package networkflow

import (
	"bytes"
	"encoding/json"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func decodeAndNormalizeFilters(raw json.RawMessage, limits EffectiveLimits) ([]Filter, *httpapi.APIError) {
	if len(raw) == 0 {
		return nil, nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, invalidFilter("filters", "invalid_value")
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, invalidFilter("filters", "invalid_value")
	}
	if int64(len(entries)) > limits.MaxFiltersPerQuery {
		return nil, invalidFilter("filters", "too_many_filters")
	}
	filters := make([]Filter, 0, len(entries))
	seen := map[string]struct{}{}
	for _, entry := range entries {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(entry, &object); err != nil || object == nil {
			return nil, invalidFilter("filters", "invalid_value")
		}
		if apiErr := ensureAllowedMembers(object, "field_key", "op", "value"); apiErr != nil {
			return nil, invalidFilter("filters", "unknown_member")
		}
		field, apiErr := requiredJSONString(object, "field_key")
		if apiErr != nil || !isFilterField(field) {
			return nil, invalidFilter("field_key", "unknown_field")
		}
		op, apiErr := requiredJSONString(object, "op")
		if apiErr != nil {
			return nil, invalidFilter("op", "operator_not_allowed")
		}
		filter := Filter{FieldKey: field, Op: op}
		if !filterOpAllowed(filter) {
			return nil, invalidFilter("op", "operator_not_allowed")
		}
		valueRaw, hasValue := object["value"]
		if op == "is_null" || op == "not_null" {
			if hasValue {
				return nil, invalidFilter("value", "value_forbidden")
			}
		} else {
			if !hasValue || bytes.Equal(bytes.TrimSpace(valueRaw), []byte("null")) {
				return nil, invalidFilter("value", "invalid_value")
			}
			value, apiErr := normalizeFilterValue(field, op, valueRaw)
			if apiErr != nil {
				return nil, apiErr
			}
			filter.Value = value
		}
		key := string(canonicalJSON(filter))
		if _, exists := seen[key]; exists {
			return nil, invalidFilter("filters", "duplicate_filter")
		}
		seen[key] = struct{}{}
		filters = append(filters, filter)
	}
	sort.SliceStable(filters, func(i, j int) bool {
		return string(canonicalJSON(filters[i])) < string(canonicalJSON(filters[j]))
	})
	return filters, nil
}

func normalizeFilterValue(field, op string, raw json.RawMessage) (any, *httpapi.APIError) {
	switch op {
	case "in":
		var entries []json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil || len(entries) == 0 || len(entries) > 256 {
			return nil, invalidFilter("value", "invalid_value")
		}
		values := make([]any, 0, len(entries))
		for _, entry := range entries {
			value, apiErr := normalizeFilterScalar(field, entry)
			if apiErr != nil {
				return nil, apiErr
			}
			values = append(values, value)
		}
		sort.SliceStable(values, func(i, j int) bool { return compareFilterValues(field, values[i], values[j]) < 0 })
		for index := 1; index < len(values); index++ {
			if compareFilterValues(field, values[index-1], values[index]) == 0 {
				return nil, invalidFilter("value", "duplicate_in_value")
			}
		}
		return values, nil
	case "range":
		return normalizeFilterRange(field, raw)
	case "cidr_contains":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || strings.Contains(value, "%") {
			return nil, invalidFilter("value", "invalid_value")
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		return prefix.Masked().String(), nil
	case "prefix", "contains":
		value, apiErr := normalizeFilterScalar(field, raw)
		if apiErr != nil || value == "" {
			return nil, invalidFilter("value", "invalid_value")
		}
		return value, nil
	default:
		return normalizeFilterScalar(field, raw)
	}
}

func normalizeFilterScalar(field string, raw json.RawMessage) (any, *httpapi.APIError) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, invalidFilter("value", "invalid_value")
	}
	switch field {
	case FieldSrcIP, FieldDstIP, FieldEndpointIP:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		canonical, err := parseIPLiteral(value)
		if err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		return canonical, nil
	case FieldSrcPort, FieldDstPort:
		return normalizeFilterInteger(raw, 0, 65535)
	case FieldIPProtocol:
		return normalizeFilterInteger(raw, 0, 255)
	case "source_row_number":
		return normalizeFilterInteger(raw, 1, uint64(^uint64(0)>>1))
	case FieldBytesCount, FieldPacketsCount:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		canonical, err := parseUint64Decimal(value)
		if err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		return canonical, nil
	case FieldFlowStartUTC, FieldFlowEndUTC:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		profile := materializeTimestampProfile(TimestampProfile{SchemaID: timestampProfileSchemaID, Mode: "rfc3339", Precision: "microseconds"})
		parsed, err := parseExactRFC3339(value, profile)
		if err != nil || !strings.ContainsAny(value[len("0001-01-01T00:00:00"):], "Z+-") {
			return nil, invalidFilter("value", "invalid_value")
		}
		return formatTimestamp(parsed), nil
	case FieldExporterID, FieldInputInterface, FieldOutputInterface:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		if _, err := parseBoundedText256(value); err != nil {
			return nil, invalidFilter("value", "invalid_value")
		}
		return value, nil
	default:
		return nil, invalidFilter("field_key", "unknown_field")
	}
}

func normalizeFilterInteger(raw json.RawMessage, minimum, maximum uint64) (any, *httpapi.APIError) {
	text := string(raw)
	if !unsignedDecimalRE.MatchString(text) {
		return nil, invalidFilter("value", "invalid_value")
	}
	value, err := strconv.ParseUint(text, 10, 64)
	if err != nil || value < minimum || value > maximum {
		return nil, invalidFilter("value", "invalid_value")
	}
	return json.Number(text), nil
}

func normalizeFilterRange(field string, raw json.RawMessage) (any, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, invalidFilter("value", "invalid_value")
	}
	upperName := "lte"
	if field == FieldFlowStartUTC || field == FieldFlowEndUTC {
		upperName = "lt"
	}
	if apiErr := ensureAllowedMembers(object, "gte", upperName); apiErr != nil {
		return nil, invalidFilter("value", "invalid_value")
	}
	result := map[string]any{}
	for _, name := range []string{"gte", upperName} {
		valueRaw, exists := object[name]
		if !exists || bytes.Equal(bytes.TrimSpace(valueRaw), []byte("null")) {
			continue
		}
		value, apiErr := normalizeFilterScalar(field, valueRaw)
		if apiErr != nil {
			return nil, apiErr
		}
		result[name] = value
	}
	if len(result) == 0 {
		return nil, invalidFilter("value", "empty_range")
	}
	if lower, ok := result["gte"]; ok {
		if upper, ok := result[upperName]; ok {
			cmp := compareFilterValues(field, lower, upper)
			if cmp > 0 || upperName == "lt" && cmp == 0 {
				return nil, invalidFilter("value", "empty_range")
			}
		}
	}
	return result, nil
}

func compareFilterValues(field string, left, right any) int {
	if field == FieldSrcIP || field == FieldDstIP || field == FieldEndpointIP {
		return compareIPValues(left, right)
	}
	return compareScalar(left, right)
}
