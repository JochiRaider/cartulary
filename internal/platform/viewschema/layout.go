package viewschema

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
)

const LayoutSchemaID = "cartulary.layout.v1"

type LayoutError struct {
	Field      string
	ReasonCode string
}

type Layout struct {
	LayoutSchemaID  string              `json:"layout_schema_id"`
	ColumnOrder     []string            `json:"column_order"`
	HiddenFieldKeys []string            `json:"hidden_field_keys"`
	ColumnWidths    []LayoutColumnWidth `json:"column_widths"`
}

type LayoutColumnWidth struct {
	FieldKey string `json:"field_key"`
	WidthPX  int    `json:"width_px"`
}

func DefaultLayout(viewSchemaID string) (json.RawMessage, *LayoutError) {
	resource, ok := LookupPublicResource(viewSchemaID)
	if !ok {
		return nil, &LayoutError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	layout := defaultLayoutResource(resource)
	payload, err := json.Marshal(layout)
	if err != nil {
		return nil, &LayoutError{Field: "layout_json", ReasonCode: "invalid_value"}
	}
	return json.RawMessage(payload), nil
}

func NormalizeLayout(raw json.RawMessage, viewSchemaID string) (json.RawMessage, *LayoutError) {
	resource, ok := LookupPublicResource(viewSchemaID)
	if !ok {
		return nil, &LayoutError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("{}")) {
		payload, err := json.Marshal(defaultLayoutResource(resource))
		if err != nil {
			return nil, &LayoutError{Field: "layout_json", ReasonCode: "invalid_value"}
		}
		return json.RawMessage(payload), nil
	}
	if bytes.Equal(trimmed, []byte("null")) {
		return nil, &LayoutError{Field: "layout_json", ReasonCode: "field_not_nullable"}
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &top); err != nil || top == nil {
		return nil, &LayoutError{Field: "layout_json", ReasonCode: "invalid_value"}
	}
	for key := range top {
		switch key {
		case "layout_schema_id", "column_order", "hidden_field_keys", "column_widths":
		default:
			return nil, &LayoutError{Field: "layout_json." + key, ReasonCode: "unknown_field"}
		}
	}

	layoutSchemaID, err := requiredString(top, "layout_schema_id")
	if err != nil {
		return nil, err
	}
	if layoutSchemaID != LayoutSchemaID {
		return nil, &LayoutError{Field: "layout_json.layout_schema_id", ReasonCode: "invalid_layout_schema"}
	}

	columnOrder, layoutErr := requiredStringArray(top, "column_order")
	if layoutErr != nil {
		return nil, layoutErr
	}
	hiddenFieldKeys, layoutErr := requiredStringArray(top, "hidden_field_keys")
	if layoutErr != nil {
		return nil, layoutErr
	}
	widths, layoutErr := requiredColumnWidths(top)
	if layoutErr != nil {
		return nil, layoutErr
	}

	if hasDuplicate(columnOrder) {
		return nil, &LayoutError{Field: "layout_json.column_order", ReasonCode: "duplicate_field_key"}
	}
	if hasDuplicate(hiddenFieldKeys) {
		return nil, &LayoutError{Field: "layout_json.hidden_field_keys", ReasonCode: "duplicate_field_key"}
	}
	if !slices.IsSorted(hiddenFieldKeys) {
		return nil, &LayoutError{Field: "layout_json.hidden_field_keys", ReasonCode: "invalid_field_order"}
	}
	var evolveErr *LayoutError
	columnOrder, hiddenFieldKeys, evolveErr = normalizeAdditiveHiddenFields(resource, columnOrder, hiddenFieldKeys)
	if evolveErr != nil {
		return nil, evolveErr
	}
	allowedOrder := layoutFieldKeys(resource)
	if !sameStringSet(columnOrder, allowedOrder) {
		return nil, &LayoutError{Field: "layout_json.column_order", ReasonCode: "invalid_column_order"}
	}
	orderSet := stringSet(columnOrder)
	for _, fieldKey := range hiddenFieldKeys {
		if _, ok := orderSet[fieldKey]; !ok {
			return nil, &LayoutError{Field: "layout_json.hidden_field_keys", ReasonCode: "unknown_field_key"}
		}
	}
	seenWidths := make(map[string]struct{}, len(widths))
	previousWidthKey := ""
	for index, width := range widths {
		fieldPath := "layout_json.column_widths[" + itoa(index) + "]"
		if width.FieldKey == "record_id" || width.FieldKey == "row_version" {
			return nil, &LayoutError{Field: fieldPath + ".field_key", ReasonCode: "forbidden_field"}
		}
		if _, ok := orderSet[width.FieldKey]; !ok {
			return nil, &LayoutError{Field: fieldPath + ".field_key", ReasonCode: "unknown_field_key"}
		}
		if _, ok := seenWidths[width.FieldKey]; ok {
			return nil, &LayoutError{Field: fieldPath + ".field_key", ReasonCode: "duplicate_field_key"}
		}
		if previousWidthKey != "" && strings.Compare(previousWidthKey, width.FieldKey) > 0 {
			return nil, &LayoutError{Field: "layout_json.column_widths", ReasonCode: "invalid_field_order"}
		}
		if width.WidthPX < 40 || width.WidthPX > 4096 {
			return nil, &LayoutError{Field: fieldPath + ".width_px", ReasonCode: "invalid_width_px"}
		}
		seenWidths[width.FieldKey] = struct{}{}
		previousWidthKey = width.FieldKey
	}

	payload, marshalErr := json.Marshal(Layout{
		LayoutSchemaID:  LayoutSchemaID,
		ColumnOrder:     columnOrder,
		HiddenFieldKeys: hiddenFieldKeys,
		ColumnWidths:    widths,
	})
	if marshalErr != nil {
		return nil, &LayoutError{Field: "layout_json", ReasonCode: "invalid_value"}
	}
	return json.RawMessage(payload), nil
}

func defaultLayoutResource(resource ViewSchemaResource) Layout {
	columnOrder := layoutFieldKeys(resource)
	hidden := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		if field.DefaultHidden {
			hidden = append(hidden, field.FieldKey)
		}
	}
	slices.Sort(hidden)
	return Layout{
		LayoutSchemaID:  LayoutSchemaID,
		ColumnOrder:     columnOrder,
		HiddenFieldKeys: hidden,
		ColumnWidths:    []LayoutColumnWidth{},
	}
}

func layoutFieldKeys(resource ViewSchemaResource) []string {
	keys := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		if field.FieldKey == "record_id" || field.FieldKey == "row_version" {
			continue
		}
		keys = append(keys, field.FieldKey)
	}
	return keys
}

func normalizeAdditiveHiddenFields(resource ViewSchemaResource, columnOrder []string, hiddenFieldKeys []string) ([]string, []string, *LayoutError) {
	allowedOrder := layoutFieldKeys(resource)
	allowedSet := stringSet(allowedOrder)
	orderSet := stringSet(columnOrder)
	for _, fieldKey := range columnOrder {
		if _, ok := allowedSet[fieldKey]; !ok {
			return nil, nil, &LayoutError{Field: "layout_json.column_order", ReasonCode: "invalid_column_order"}
		}
	}

	fieldsByKey := make(map[string]ViewFieldEntry, len(resource.Fields))
	for _, field := range resource.Fields {
		if field.FieldKey == "record_id" || field.FieldKey == "row_version" {
			continue
		}
		fieldsByKey[field.FieldKey] = field
	}

	hiddenSet := stringSet(hiddenFieldKeys)
	addedHidden := false
	for _, fieldKey := range allowedOrder {
		if _, ok := orderSet[fieldKey]; ok {
			continue
		}
		field, ok := fieldsByKey[fieldKey]
		if !ok || !field.DefaultHidden || field.WriteKind != "read_only" {
			return nil, nil, &LayoutError{Field: "layout_json.column_order", ReasonCode: "invalid_column_order"}
		}
		columnOrder = append(columnOrder, fieldKey)
		orderSet[fieldKey] = struct{}{}
		if _, ok := hiddenSet[fieldKey]; !ok {
			hiddenFieldKeys = append(hiddenFieldKeys, fieldKey)
			hiddenSet[fieldKey] = struct{}{}
			addedHidden = true
		}
	}
	if addedHidden {
		slices.Sort(hiddenFieldKeys)
	}
	return columnOrder, hiddenFieldKeys, nil
}

func requiredString(raw map[string]json.RawMessage, key string) (string, *LayoutError) {
	value, ok := raw[key]
	if !ok {
		return "", &LayoutError{Field: "layout_json." + key, ReasonCode: "missing_required_field"}
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
		return "", &LayoutError{Field: "layout_json." + key, ReasonCode: "invalid_value"}
	}
	if parsed == "record_id" || parsed == "row_version" {
		return "", &LayoutError{Field: "layout_json." + key, ReasonCode: "forbidden_field"}
	}
	return parsed, nil
}

func requiredStringArray(raw map[string]json.RawMessage, key string) ([]string, *LayoutError) {
	value, ok := raw[key]
	if !ok {
		return nil, &LayoutError{Field: "layout_json." + key, ReasonCode: "missing_required_field"}
	}
	var parsed []string
	if err := json.Unmarshal(value, &parsed); err != nil || parsed == nil {
		return nil, &LayoutError{Field: "layout_json." + key, ReasonCode: "invalid_value"}
	}
	for index, fieldKey := range parsed {
		if fieldKey == "" {
			return nil, &LayoutError{Field: "layout_json." + key + "[" + itoa(index) + "]", ReasonCode: "invalid_value"}
		}
		if fieldKey == "record_id" || fieldKey == "row_version" {
			return nil, &LayoutError{Field: "layout_json." + key + "[" + itoa(index) + "]", ReasonCode: "forbidden_field"}
		}
	}
	return parsed, nil
}

func requiredColumnWidths(raw map[string]json.RawMessage) ([]LayoutColumnWidth, *LayoutError) {
	value, ok := raw["column_widths"]
	if !ok {
		return nil, &LayoutError{Field: "layout_json.column_widths", ReasonCode: "missing_required_field"}
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(value, &items); err != nil || items == nil {
		return nil, &LayoutError{Field: "layout_json.column_widths", ReasonCode: "invalid_value"}
	}
	result := make([]LayoutColumnWidth, 0, len(items))
	for index, item := range items {
		fieldPath := "layout_json.column_widths[" + itoa(index) + "]"
		if len(item) != 2 {
			return nil, &LayoutError{Field: fieldPath, ReasonCode: "invalid_value"}
		}
		for key := range item {
			if key != "field_key" && key != "width_px" {
				return nil, &LayoutError{Field: fieldPath + "." + key, ReasonCode: "unknown_field"}
			}
		}
		var width LayoutColumnWidth
		if err := json.Unmarshal(item["field_key"], &width.FieldKey); err != nil || width.FieldKey == "" {
			return nil, &LayoutError{Field: fieldPath + ".field_key", ReasonCode: "invalid_value"}
		}
		if err := json.Unmarshal(item["width_px"], &width.WidthPX); err != nil {
			return nil, &LayoutError{Field: fieldPath + ".width_px", ReasonCode: "invalid_width_px"}
		}
		result = append(result, width)
	}
	return result, nil
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func sameStringSet(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	rightSet := stringSet(right)
	for _, value := range left {
		if _, ok := rightSet[value]; !ok {
			return false
		}
	}
	return true
}

func hasDuplicate(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buffer [20]byte
	i := len(buffer)
	for value > 0 {
		i--
		buffer[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buffer[i:])
}
