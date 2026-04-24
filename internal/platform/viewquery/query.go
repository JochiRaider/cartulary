package viewquery

import (
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/cases"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type FieldKind string

const (
	FieldKindString           FieldKind = "string"
	FieldKindCaseFoldedString FieldKind = "case_folded_string"
	FieldKindBool             FieldKind = "bool"
	FieldKindDate             FieldKind = "date"
	FieldKindTimestamp        FieldKind = "timestamp"
	FieldKindStringCollection FieldKind = "string_collection"
	FieldKindFullText         FieldKind = "full_text"
)

type ValidationError struct {
	Field          string
	FieldKey       string
	FilterIndex    *int
	ReasonCode     string
	RequestedCount *int
	MaxCount       *int
}

type FieldSpec struct {
	Kind       FieldKind
	AllowedOps map[string]struct{}
}

type Query struct {
	Meta viewschema.QueryMeta
}

var queryCaseFolder = cases.Fold()

func Decode(reader io.Reader, viewSchemaID string) (Query, *ValidationError) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return Query{}, &ValidationError{
			Field:      "view_schema_id",
			ReasonCode: "unknown_view_schema",
		}
	}
	spec, ok := lookupSpec(viewSchemaID)
	if !ok {
		return Query{}, &ValidationError{
			Field:      "view_schema_id",
			ReasonCode: "unknown_view_schema",
		}
	}

	raw, err := decodeObject(reader)
	if err != nil {
		return Query{}, err
	}
	if err := validateTopLevelMembers(raw); err != nil {
		return Query{}, err
	}

	sortEntries, err := normalizeSort(raw["sort"], schema)
	if err != nil {
		return Query{}, err
	}
	filters, err := normalizeFilters(raw["filters"], spec)
	if err != nil {
		return Query{}, err
	}
	groupBy, err := normalizeGroupBy(raw["group_by"], schema)
	if err != nil {
		return Query{}, err
	}
	if err := rejectPaginationMembers(raw); err != nil {
		return Query{}, err
	}

	return Query{
		Meta: viewschema.QueryMeta{
			Filters: filters,
			Sort:    effectiveSort(sortEntries, schema.DefaultSort()),
			GroupBy: groupBy,
		},
	}, nil
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *ValidationError) {
	if reader == nil {
		return map[string]json.RawMessage{}, nil
	}
	decoder := json.NewDecoder(reader)
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		if err == io.EOF {
			return map[string]json.RawMessage{}, nil
		}
		return nil, &ValidationError{ReasonCode: "request_not_object"}
	}
	if raw == nil {
		return nil, &ValidationError{ReasonCode: "request_not_object"}
	}
	return raw, nil
}

func validateTopLevelMembers(raw map[string]json.RawMessage) *ValidationError {
	for key := range raw {
		switch key {
		case "sort", "filters", "group_by", "limit", "cursor_token", "page", "offset", "block_size", "page_size":
		default:
			return &ValidationError{
				Field:      key,
				ReasonCode: "unknown_field",
			}
		}
	}
	return nil
}

func rejectPaginationMembers(raw map[string]json.RawMessage) *ValidationError {
	for _, key := range []string{"limit", "cursor_token", "page", "offset", "block_size", "page_size"} {
		if _, ok := raw[key]; ok {
			return &ValidationError{
				Field:      key,
				ReasonCode: "invalid_limit",
			}
		}
	}
	return nil
}

func normalizeSort(raw json.RawMessage, schema viewschema.Schema) ([]viewschema.SortEntry, *ValidationError) {
	if len(raw) == 0 {
		return nil, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, &ValidationError{
			Field:      "sort",
			ReasonCode: "invalid_sort_entry",
		}
	}
	if len(items) > config.PublicSortLimit {
		requested := len(items)
		maxCount := config.PublicSortLimit
		return nil, &ValidationError{
			Field:          "sort",
			ReasonCode:     "sort_count_exceeded",
			RequestedCount: &requested,
			MaxCount:       &maxCount,
		}
	}

	allowed := make(map[string]struct{}, len(schema.SortFields()))
	for _, fieldKey := range schema.SortFields() {
		allowed[fieldKey] = struct{}{}
	}
	allowed["record_id"] = struct{}{}

	normalized := make([]viewschema.SortEntry, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		entry, err := normalizeSortEntry(item, allowed)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[entry.FieldKey]; ok {
			return nil, &ValidationError{
				Field:      "sort[" + itoa(index) + "].field_key",
				FieldKey:   entry.FieldKey,
				ReasonCode: "duplicate_sort_field",
			}
		}
		seen[entry.FieldKey] = struct{}{}
		normalized = append(normalized, entry)
	}
	return normalized, nil
}

func normalizeSortEntry(raw json.RawMessage, allowed map[string]struct{}) (viewschema.SortEntry, *ValidationError) {
	var entry map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entry); err != nil {
		return viewschema.SortEntry{}, &ValidationError{
			Field:      "sort",
			ReasonCode: "invalid_sort_entry",
		}
	}
	if len(entry) != 2 {
		return viewschema.SortEntry{}, &ValidationError{
			Field:      "sort",
			ReasonCode: "invalid_sort_entry",
		}
	}
	for key := range entry {
		if key != "field_key" && key != "direction" {
			return viewschema.SortEntry{}, &ValidationError{
				Field:      "sort." + key,
				ReasonCode: "invalid_sort_entry",
			}
		}
	}

	var result viewschema.SortEntry
	if err := json.Unmarshal(entry["field_key"], &result.FieldKey); err != nil || strings.TrimSpace(result.FieldKey) == "" {
		return viewschema.SortEntry{}, &ValidationError{
			Field:      "sort.field_key",
			ReasonCode: "invalid_sort_entry",
		}
	}
	if _, ok := allowed[result.FieldKey]; !ok {
		return viewschema.SortEntry{}, &ValidationError{
			Field:      "sort.field_key",
			FieldKey:   result.FieldKey,
			ReasonCode: "sort_field_not_allowed",
		}
	}
	if err := json.Unmarshal(entry["direction"], &result.Direction); err != nil || (result.Direction != "asc" && result.Direction != "desc") {
		return viewschema.SortEntry{}, &ValidationError{
			Field:      "sort.direction",
			FieldKey:   result.FieldKey,
			ReasonCode: "invalid_sort_entry",
		}
	}
	return result, nil
}

func normalizeFilters(raw json.RawMessage, spec map[string]FieldSpec) ([]viewschema.Filter, *ValidationError) {
	if len(raw) == 0 {
		return []viewschema.Filter{}, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, &ValidationError{
			Field:      "filters",
			ReasonCode: "invalid_filter_operand",
		}
	}
	if len(items) > config.PublicFilterLimit {
		requested := len(items)
		maxCount := config.PublicFilterLimit
		return nil, &ValidationError{
			Field:          "filters",
			ReasonCode:     "filter_count_exceeded",
			RequestedCount: &requested,
			MaxCount:       &maxCount,
		}
	}

	normalized := make([]viewschema.Filter, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		filter, err := normalizeFilter(index, item, spec)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[filter.FieldKey]; ok {
			return nil, filterError(index, filter.FieldKey, "duplicate_filter_field")
		}
		seen[filter.FieldKey] = struct{}{}
		normalized = append(normalized, filter)
	}

	slices.SortFunc(normalized, func(left viewschema.Filter, right viewschema.Filter) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return normalized, nil
}

func normalizeFilter(index int, raw json.RawMessage, spec map[string]FieldSpec) (viewschema.Filter, *ValidationError) {
	var item map[string]json.RawMessage
	if err := json.Unmarshal(raw, &item); err != nil {
		return viewschema.Filter{}, filterError(index, "", "invalid_filter_operand")
	}
	if len(item) != 3 {
		return viewschema.Filter{}, filterError(index, "", "invalid_filter_operand")
	}
	for key := range item {
		if key != "field_key" && key != "op" && key != "arg" {
			return viewschema.Filter{}, filterError(index, "", "invalid_filter_operand")
		}
	}

	var fieldKey string
	if err := json.Unmarshal(item["field_key"], &fieldKey); err != nil || strings.TrimSpace(fieldKey) == "" {
		return viewschema.Filter{}, filterError(index, "", "invalid_filter_operand")
	}
	fieldSpec, ok := spec[fieldKey]
	if !ok {
		return viewschema.Filter{}, filterError(index, fieldKey, "unknown_filter_field")
	}

	var op string
	if err := json.Unmarshal(item["op"], &op); err != nil || strings.TrimSpace(op) == "" {
		return viewschema.Filter{}, filterError(index, fieldKey, "invalid_filter_operand")
	}
	if _, ok := fieldSpec.AllowedOps[op]; !ok {
		return viewschema.Filter{}, filterError(index, fieldKey, "operator_not_allowed")
	}

	arg, err := normalizeFilterArg(index, fieldKey, op, fieldSpec.Kind, item["arg"])
	if err != nil {
		return viewschema.Filter{}, err
	}
	return viewschema.Filter{
		FieldKey: fieldKey,
		Op:       op,
		Arg:      arg,
	}, nil
}

func normalizeFilterArg(index int, fieldKey string, op string, kind FieldKind, raw json.RawMessage) (map[string]any, *ValidationError) {
	var arg map[string]json.RawMessage
	if err := json.Unmarshal(raw, &arg); err != nil || len(arg) == 0 {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}

	switch op {
	case "eq":
		return normalizeEqualityArg(index, fieldKey, kind, arg)
	case "range":
		return normalizeRangeArg(index, fieldKey, kind, arg)
	case "contains_any", "contains_all":
		return normalizeSetArg(index, fieldKey, kind, arg)
	case "prefix":
		return normalizePrefixArg(index, fieldKey, kind, arg)
	case "full_text":
		return normalizeFullTextArg(index, fieldKey, kind, arg)
	default:
		return nil, filterError(index, fieldKey, "operator_not_allowed")
	}
}

func normalizeEqualityArg(index int, fieldKey string, kind FieldKind, arg map[string]json.RawMessage) (map[string]any, *ValidationError) {
	valueRaw, hasValue := arg["value"]
	valuesRaw, hasValues := arg["values"]
	if hasValue == hasValues {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	if hasValue {
		if len(arg) != 1 {
			return nil, filterError(index, fieldKey, "invalid_filter_operand")
		}
		value, err := normalizeScalar(kind, valueRaw, true)
		if err != nil {
			return nil, filterError(index, fieldKey, "invalid_filter_operand")
		}
		return map[string]any{"value": value}, nil
	}
	return normalizeValuesArg(index, fieldKey, kind, valuesRaw)
}

func normalizeRangeArg(index int, fieldKey string, kind FieldKind, arg map[string]json.RawMessage) (map[string]any, *ValidationError) {
	if kind != FieldKindDate && kind != FieldKindTimestamp {
		return nil, filterError(index, fieldKey, "operator_not_allowed")
	}
	for key := range arg {
		switch key {
		case "gt", "gte", "lt", "lte":
		default:
			return nil, filterError(index, fieldKey, "invalid_filter_operand")
		}
	}
	if len(arg) == 0 {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	if _, ok := arg["gt"]; ok {
		if _, conflict := arg["gte"]; conflict {
			return nil, filterError(index, fieldKey, "invalid_filter_operand")
		}
	}
	if _, ok := arg["lt"]; ok {
		if _, conflict := arg["lte"]; conflict {
			return nil, filterError(index, fieldKey, "invalid_filter_operand")
		}
	}

	canonical := make(map[string]any, len(arg))
	for _, key := range []string{"gt", "gte", "lt", "lte"} {
		rawValue, ok := arg[key]
		if !ok {
			continue
		}
		value, err := normalizeScalar(kind, rawValue, false)
		if err != nil {
			return nil, filterError(index, fieldKey, "invalid_filter_operand")
		}
		canonical[key] = value
	}
	if len(canonical) == 0 {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}

	if contradictionDetected(canonical) {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	return canonical, nil
}

func contradictionDetected(bounds map[string]any) bool {
	var (
		lowerValue string
		lowerOpen  bool
		upperValue string
		upperOpen  bool
		ok         bool
	)
	if value, exists := bounds["gt"]; exists {
		lowerValue, ok = value.(string)
		if !ok {
			return true
		}
		lowerOpen = true
	}
	if value, exists := bounds["gte"]; exists {
		lowerValue, ok = value.(string)
		if !ok {
			return true
		}
		lowerOpen = false
	}
	if value, exists := bounds["lt"]; exists {
		upperValue, ok = value.(string)
		if !ok {
			return true
		}
		upperOpen = true
	}
	if value, exists := bounds["lte"]; exists {
		upperValue, ok = value.(string)
		if !ok {
			return true
		}
		upperOpen = false
	}
	if lowerValue == "" || upperValue == "" {
		return false
	}
	cmp := strings.Compare(lowerValue, upperValue)
	if cmp > 0 {
		return true
	}
	if cmp < 0 {
		return false
	}
	return lowerOpen || upperOpen
}

func normalizeSetArg(index int, fieldKey string, kind FieldKind, arg map[string]json.RawMessage) (map[string]any, *ValidationError) {
	if len(arg) != 1 {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	valuesRaw, ok := arg["values"]
	if !ok {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	return normalizeValuesArg(index, fieldKey, kind, valuesRaw)
}

func normalizeValuesArg(index int, fieldKey string, kind FieldKind, raw json.RawMessage) (map[string]any, *ValidationError) {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || len(values) == 0 {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	items, err := normalizeScalarSet(kind, values)
	if err != nil {
		return nil, filterError(index, fieldKey, err.ReasonCode)
	}
	return map[string]any{"values": items}, nil
}

func normalizePrefixArg(index int, fieldKey string, kind FieldKind, arg map[string]json.RawMessage) (map[string]any, *ValidationError) {
	if kind != FieldKindCaseFoldedString {
		return nil, filterError(index, fieldKey, "operator_not_allowed")
	}
	if len(arg) != 1 {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	valueRaw, ok := arg["value"]
	if !ok {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	var rawValue string
	if err := json.Unmarshal(valueRaw, &rawValue); err != nil {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	normalized, ok := fieldnorm.NormalizeLine(rawValue)
	if !ok {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	return map[string]any{"value": queryCaseFolder.String(normalized)}, nil
}

func normalizeFullTextArg(index int, fieldKey string, kind FieldKind, arg map[string]json.RawMessage) (map[string]any, *ValidationError) {
	if kind != FieldKindFullText {
		return nil, filterError(index, fieldKey, "operator_not_allowed")
	}
	if len(arg) != 1 {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	queryRaw, ok := arg["query"]
	if !ok {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	var rawQuery string
	if err := json.Unmarshal(queryRaw, &rawQuery); err != nil {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	trimmed := strings.TrimFunc(rawQuery, unicode.IsSpace)
	if trimmed == "" {
		return nil, filterError(index, fieldKey, "invalid_filter_operand")
	}
	tokens := canonicalTokens(trimmed)
	if len(tokens) == 0 {
		return nil, filterError(index, fieldKey, "empty_full_text_after_tokenization")
	}
	return map[string]any{"query": strings.Join(tokens, " ")}, nil
}

func canonicalTokens(raw string) []string {
	var (
		builder strings.Builder
		tokens  []string
		seen    = map[string]struct{}{}
	)
	flush := func() {
		if builder.Len() == 0 {
			return
		}
		token := queryCaseFolder.String(builder.String())
		builder.Reset()
		if token == "" {
			return
		}
		if _, ok := seen[token]; ok {
			return
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	slices.Sort(tokens)
	return tokens
}

func normalizeScalarSet(kind FieldKind, values []json.RawMessage) ([]any, *ValidationError) {
	items := make([]scalarValue, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		if string(raw) == "null" {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		value, err := normalizeScalar(kind, raw, false)
		if err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		item := makeScalarValue(value)
		if _, ok := seen[item.DedupeKey]; ok {
			continue
		}
		seen[item.DedupeKey] = struct{}{}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, &ValidationError{ReasonCode: "empty_values_after_normalization"}
	}
	slices.SortFunc(items, func(left scalarValue, right scalarValue) int {
		return strings.Compare(left.SortKey, right.SortKey)
	})
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, item.Value)
	}
	return result, nil
}

type scalarValue struct {
	Value     any
	SortKey   string
	DedupeKey string
}

func makeScalarValue(value any) scalarValue {
	switch typed := value.(type) {
	case string:
		return scalarValue{
			Value:     typed,
			SortKey:   "s:" + typed,
			DedupeKey: "s:" + typed,
		}
	case bool:
		if typed {
			return scalarValue{Value: true, SortKey: "b:1", DedupeKey: "b:1"}
		}
		return scalarValue{Value: false, SortKey: "b:0", DedupeKey: "b:0"}
	default:
		return scalarValue{
			Value:     value,
			SortKey:   "u:",
			DedupeKey: "u:",
		}
	}
}

func normalizeScalar(kind FieldKind, raw json.RawMessage, allowNull bool) (any, *ValidationError) {
	if allowNull && string(raw) == "null" {
		return nil, nil
	}

	switch kind {
	case FieldKindString, FieldKindStringCollection:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		normalized, ok := fieldnorm.NormalizeLine(value)
		if !ok {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		return normalized, nil
	case FieldKindCaseFoldedString:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		normalized, ok := fieldnorm.NormalizeLine(value)
		if !ok {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		return queryCaseFolder.String(normalized), nil
	case FieldKindBool:
		var value bool
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		return value, nil
	case FieldKindDate:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		return parsed.UTC().Format("2006-01-02"), nil
	case FieldKindTimestamp:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
		}
		return parsed.UTC().Format(time.RFC3339Nano), nil
	case FieldKindFullText:
		return nil, &ValidationError{ReasonCode: "operator_not_allowed"}
	default:
		return nil, &ValidationError{ReasonCode: "invalid_filter_operand"}
	}
}

func normalizeGroupBy(raw json.RawMessage, schema viewschema.Schema) (*string, *ValidationError) {
	if len(raw) == 0 {
		return nil, nil
	}
	if string(raw) == "null" {
		return nil, &ValidationError{
			Field:      "group_by",
			ReasonCode: "invalid_group_by",
		}
	}
	var groupBy string
	if err := json.Unmarshal(raw, &groupBy); err != nil || strings.TrimSpace(groupBy) == "" {
		return nil, &ValidationError{
			Field:      "group_by",
			ReasonCode: "invalid_group_by",
		}
	}
	for _, candidate := range schema.GroupingFields() {
		if candidate == groupBy {
			value := groupBy
			return &value, nil
		}
	}
	return nil, &ValidationError{
		Field:      "group_by",
		FieldKey:   groupBy,
		ReasonCode: "group_by_not_allowed",
	}
}

func effectiveSort(userSort []viewschema.SortEntry, defaultSort []viewschema.SortEntry) []viewschema.SortEntry {
	effective := make([]viewschema.SortEntry, 0, len(userSort)+len(defaultSort)+1)
	seen := make(map[string]struct{}, len(userSort)+len(defaultSort)+1)
	for _, entry := range userSort {
		effective = append(effective, entry)
		seen[entry.FieldKey] = struct{}{}
	}
	for _, entry := range defaultSort {
		if _, ok := seen[entry.FieldKey]; ok {
			continue
		}
		effective = append(effective, entry)
		seen[entry.FieldKey] = struct{}{}
	}
	if _, ok := seen["record_id"]; !ok {
		effective = append(effective, viewschema.SortEntry{
			FieldKey:  "record_id",
			Direction: "asc",
		})
	}
	return effective
}

func filterError(index int, fieldKey string, reasonCode string) *ValidationError {
	return &ValidationError{
		FieldKey:    fieldKey,
		FilterIndex: intPointer(index),
		ReasonCode:  reasonCode,
	}
}

func intPointer(value int) *int {
	return &value
}

func itoa(value int) string {
	return strconv.Itoa(value)
}

func allowedOps(values ...string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(values))
	for _, value := range values {
		allowed[value] = struct{}{}
	}
	return allowed
}

func lookupSpec(viewSchemaID string) (map[string]FieldSpec, bool) {
	querySpecsOnce.Do(loadQuerySpecs)
	spec, ok := querySpecs[viewSchemaID]
	return spec, ok
}

var (
	querySpecsOnce sync.Once
	querySpecs     map[string]map[string]FieldSpec
)

func loadQuerySpecs() {
	querySpecs = make(map[string]map[string]FieldSpec)
	for _, resource := range viewschema.ListPublicResources() {
		fieldByKey := make(map[string]viewschema.ViewFieldEntry, len(resource.Fields))
		for _, field := range resource.Fields {
			fieldByKey[field.FieldKey] = field
		}
		spec := make(map[string]FieldSpec)
		for _, fieldKey := range resource.FilterFields {
			field, ok := fieldByKey[fieldKey]
			if !ok || len(field.FilterOps) == 0 {
				continue
			}
			spec[fieldKey] = FieldSpec{
				Kind:       fieldKindForContract(field.ReadKind, field.FilterOps),
				AllowedOps: allowedOps(field.FilterOps...),
			}
		}
		for _, predicate := range resource.SyntheticFilterPredicates {
			if len(predicate.FilterOps) == 0 {
				continue
			}
			spec[predicate.FieldKey] = FieldSpec{
				Kind:       syntheticFieldKind(predicate.FilterOps),
				AllowedOps: allowedOps(predicate.FilterOps...),
			}
		}
		querySpecs[resource.ViewSchemaID] = spec
	}
}

func fieldKindForContract(readKind string, ops []string) FieldKind {
	switch readKind {
	case "boolean":
		return FieldKindBool
	case "date":
		return FieldKindDate
	case "timestamp":
		return FieldKindTimestamp
	case "collection":
		return FieldKindStringCollection
	default:
		if slices.Contains(ops, "prefix") {
			return FieldKindCaseFoldedString
		}
		return FieldKindString
	}
}

func syntheticFieldKind(ops []string) FieldKind {
	if slices.Contains(ops, "full_text") {
		return FieldKindFullText
	}
	return FieldKindString
}
