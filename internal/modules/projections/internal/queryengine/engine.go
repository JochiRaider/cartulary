package queryengine

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type FieldKind string

const (
	FieldKindText       FieldKind = "text"
	FieldKindTimestamp  FieldKind = "timestamp"
	FieldKindDate       FieldKind = "date"
	FieldKindBool       FieldKind = "bool"
	FieldKindNumber     FieldKind = "number"
	FieldKindCollection FieldKind = "collection"
)

type Field struct {
	Key      string
	Expr     string
	SortExpr string
	Kind     FieldKind
	Ordered  bool
}

type Surface struct {
	ViewSchemaID   string
	FromSQL        string
	RecordExpr     string
	IncidentExpr   string
	WhereSQL       string
	Fields         []Field
	GroupingFields []string
}

func (d Surface) Field(key string) (Field, bool) {
	if key == "record_id" {
		return Field{Key: "record_id", Expr: d.RecordExpr, Kind: FieldKindText}, true
	}
	for _, field := range d.Fields {
		if field.Key == key {
			return field, true
		}
	}
	return Field{}, false
}

func (f Field) OrderExpr() string {
	if f.SortExpr != "" {
		return f.SortExpr
	}
	return f.Expr
}

func ScanRows(rows pgx.Rows, definition Surface) ([]map[string]any, error) {
	result := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("read workbook row values: %w", err)
		}
		if len(values) != len(definition.Fields)+2 {
			return nil, fmt.Errorf("query workbook rows: unexpected value count %d", len(values))
		}
		row, err := BuildRow(definition, values)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workbook rows: %w", err)
	}
	return result, nil
}

func BuildQueryPageSQL(incidentID uuid.UUID, definition Surface, query viewschema.QueryMeta, window querypage.Window) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(definition.RecordExpr)
	builder.WriteString(", r.row_version")
	for _, field := range definition.Fields {
		builder.WriteString(", ")
		builder.WriteString(field.Expr)
	}
	builder.WriteString(" ")
	builder.WriteString(definition.FromSQL)
	builder.WriteString(" WHERE ")
	builder.WriteString(definition.IncidentExpr)
	builder.WriteString(" = $1 AND r.deleted_at IS NULL")
	if definition.WhereSQL != "" {
		builder.WriteString(" AND ")
		builder.WriteString(definition.WhereSQL)
	}
	args := []any{incidentID}

	for _, filter := range query.Filters {
		if err := appendGenericFilter(&builder, &args, definition, filter); err != nil {
			return "", nil, err
		}
	}
	pageFields := make(map[string]querypage.Field, len(definition.Fields)+1)
	pageFields["record_id"] = querypage.Field{Expression: definition.RecordExpr, Cast: "uuid"}
	for _, field := range definition.Fields {
		cast := ""
		switch field.Kind {
		case FieldKindTimestamp:
			cast = "timestamptz"
		case FieldKindDate:
			cast = "date"
		case FieldKindNumber:
			cast = "numeric"
		case FieldKindBool:
			cast = "boolean"
		}
		pageFields[field.Key] = querypage.Field{Expression: field.OrderExpr(), Cast: cast}
	}
	if err := querypage.AppendKeyset(&builder, &args, query.Sort, pageFields, window.Position); err != nil {
		return "", nil, err
	}

	builder.WriteString(" ORDER BY ")
	for index, entry := range query.Sort {
		if index > 0 {
			builder.WriteString(", ")
		}
		field, ok := definition.Field(entry.FieldKey)
		if !ok {
			return "", nil, fmt.Errorf("sort field %q not mapped for %s", entry.FieldKey, definition.ViewSchemaID)
		}
		builder.WriteString(field.OrderExpr())
		if entry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
		builder.WriteString(" NULLS LAST")
	}
	if err := querypage.AppendLimit(&builder, &args, window.Limit); err != nil {
		return "", nil, err
	}
	return builder.String(), args, nil
}

func appendGenericFilter(builder *strings.Builder, args *[]any, definition Surface, filter viewschema.Filter) error {
	if filter.FieldKey == "note.full_text" {
		query, _ := filter.Arg["query"].(string)
		for _, token := range strings.Fields(query) {
			builder.WriteString("\n   AND ")
			builder.WriteString(bind(args, token))
			builder.WriteString(" = ANY(regexp_split_to_array(lower(coalesce(p.title, '') || ' ' || coalesce(p.body, '')), '[^[:alnum:]]+'))")
		}
		return nil
	}
	if filter.FieldKey == "note.tags" {
		return appendTagFilter(builder, args, definition.RecordExpr, filter)
	}
	field, ok := definition.Field(filter.FieldKey)
	if !ok {
		return fmt.Errorf("filter field %q not mapped for %s", filter.FieldKey, definition.ViewSchemaID)
	}
	if field.Kind == FieldKindCollection && strings.HasSuffix(field.Key, ".tags") {
		return appendTagFilter(builder, args, definition.RecordExpr, filter)
	}
	switch filter.Op {
	case "eq":
		return appendEqualityFilter(builder, args, field, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		value = strings.ToLower(value)
		builder.WriteString("\n   AND left(lower(coalesce((")
		builder.WriteString(field.Expr)
		builder.WriteString(")::text, '')), char_length(")
		builder.WriteString(bind(args, value))
		builder.WriteString(")) = ")
		builder.WriteString(bind(args, value))
		return nil
	case "range":
		return appendRangeFilter(builder, args, field, filter.Arg)
	default:
		return fmt.Errorf("filter operator %q not mapped for %s", filter.Op, filter.FieldKey)
	}
}

func appendTagFilter(builder *strings.Builder, args *[]any, recordExpr string, filter viewschema.Filter) error {
	values, _ := filter.Arg["values"].([]any)
	textValues := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			textValues = append(textValues, text)
		}
	}
	switch filter.Op {
	case "contains_any":
		builder.WriteString("\n   AND EXISTS (SELECT 1 FROM ")
		builder.WriteString("active_record_tags_v1 rt")
		builder.WriteString(" WHERE rt.record_id = ")
		builder.WriteString(recordExpr)
		builder.WriteString(" AND rt.normalized_tag_name = ANY(")
		builder.WriteString(bind(args, textValues))
		builder.WriteString("::text[]))")
		return nil
	case "contains_all":
		for _, value := range values {
			builder.WriteString("\n   AND EXISTS (SELECT 1 FROM ")
			builder.WriteString("active_record_tags_v1 rt")
			builder.WriteString(" WHERE rt.record_id = ")
			builder.WriteString(recordExpr)
			builder.WriteString(" AND rt.normalized_tag_name = ")
			builder.WriteString(bind(args, value))
			builder.WriteString(")")
		}
		return nil
	default:
		return fmt.Errorf("tag filter operator %q not mapped", filter.Op)
	}
}

func appendEqualityFilter(builder *strings.Builder, args *[]any, field Field, arg map[string]any) error {
	if value, ok := arg["value"]; ok {
		builder.WriteString("\n   AND ")
		if value == nil {
			builder.WriteString(field.Expr)
			builder.WriteString(" IS NULL")
			return nil
		}
		appendComparableExpr(builder, field)
		builder.WriteString(" = ")
		builder.WriteString(bindWithFieldCast(args, comparableFilterValue(field, value), field))
		return nil
	}
	values, _ := arg["values"].([]any)
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		appendComparableExpr(builder, field)
		builder.WriteString(" = ")
		builder.WriteString(bindWithFieldCast(args, comparableFilterValue(field, value), field))
	}
	builder.WriteString(")")
	return nil
}

func comparableFilterValue(field Field, value any) any {
	if field.Kind != FieldKindText {
		return value
	}
	if text, ok := value.(string); ok {
		return strings.ToLower(text)
	}
	return value
}

func appendRangeFilter(builder *strings.Builder, args *[]any, field Field, arg map[string]any) error {
	for _, bound := range []struct {
		key string
		op  string
	}{
		{key: "gt", op: ">"},
		{key: "gte", op: ">="},
		{key: "lt", op: "<"},
		{key: "lte", op: "<="},
	} {
		value, ok := arg[bound.key]
		if !ok {
			continue
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(field.Expr)
		builder.WriteByte(' ')
		builder.WriteString(bound.op)
		builder.WriteByte(' ')
		builder.WriteString(bindWithFieldCast(args, value, field))
	}
	return nil
}

func appendComparableExpr(builder *strings.Builder, field Field) {
	switch field.Kind {
	case FieldKindText:
		builder.WriteString("lower(coalesce((")
		builder.WriteString(field.Expr)
		builder.WriteString(")::text, ''))")
	default:
		builder.WriteString(field.Expr)
	}
}

func bind(args *[]any, value any) string {
	*args = append(*args, value)
	return fmt.Sprintf("$%d", len(*args))
}

func bindWithFieldCast(args *[]any, value any, field Field) string {
	placeholder := bind(args, value)
	switch field.Kind {
	case FieldKindTimestamp:
		return placeholder + "::timestamptz"
	case FieldKindDate:
		return placeholder + "::date"
	default:
		return placeholder
	}
}

func BuildRow(definition Surface, values []any) (map[string]any, error) {
	recordID, err := uuidValue(values[0])
	if err != nil {
		return nil, err
	}
	cells := make(map[string]any, len(definition.Fields))
	fieldValues := make(map[string]any, len(definition.Fields))
	for index, field := range definition.Fields {
		value := genericCellValue(field, values[index+2])
		fieldValues[field.Key] = value
		cells[field.Key] = map[string]any{"value": value}
	}
	row := map[string]any{
		"record_id":   recordID.String(),
		"row_version": values[1],
		"cells":       cells,
	}
	if len(definition.GroupingFields) > 0 {
		groupValues := make(map[string]any, len(definition.GroupingFields))
		for _, fieldKey := range definition.GroupingFields {
			groupValues[fieldKey] = fieldValues[fieldKey]
		}
		row["group_values"] = groupValues
	}
	return row, nil
}

func uuidValue(value any) (uuid.UUID, error) {
	switch typed := value.(type) {
	case uuid.UUID:
		return typed, nil
	case string:
		parsed, err := uuid.Parse(typed)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid record_id %q", typed)
		}
		return parsed, nil
	case []byte:
		if len(typed) == 16 {
			parsed, err := uuid.FromBytes(typed)
			if err != nil {
				return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid binary record_id: %w", err)
			}
			return parsed, nil
		}
		parsed, err := uuid.Parse(string(typed))
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid record_id bytes")
		}
		return parsed, nil
	case [16]byte:
		return uuid.UUID(typed), nil
	default:
		return uuid.UUID{}, fmt.Errorf("query workbook rows: record_id was %T", value)
	}
}

func genericCellValue(field Field, value any) any {
	if field.Kind == FieldKindCollection {
		if value != nil {
			if items, ok := collectionItemsFromValue(value); ok {
				return map[string]any{
					"kind":    "collection_value_v1",
					"ordered": field.Ordered,
					"items":   items,
				}
			}
		}
		return map[string]any{
			"kind":    "collection_value_v1",
			"ordered": field.Ordered,
			"items":   []map[string]any{},
		}
	}
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case time.Time:
		if field.Kind == FieldKindDate {
			return typed.UTC().Format("2006-01-02")
		}
		return typed.UTC().Format(time.RFC3339Nano)
	case uuid.UUID:
		return typed.String()
	case []byte:
		if field.Kind == FieldKindText && len(typed) == 16 {
			if parsed, err := uuid.FromBytes(typed); err == nil {
				return parsed.String()
			}
		}
		return string(typed)
	case [16]byte:
		if field.Kind == FieldKindText {
			return uuid.UUID(typed).String()
		}
		return typed
	default:
		return typed
	}
}

func collectionItemsFromValue(value any) ([]map[string]any, bool) {
	var data []byte
	switch typed := value.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	case []map[string]any:
		if typed == nil {
			return []map[string]any{}, true
		}
		return typed, true
	case []any:
		items := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			mapped, ok := item.(map[string]any)
			if !ok {
				return nil, false
			}
			items = append(items, mapped)
		}
		return items, true
	default:
		return nil, false
	}
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, false
	}
	if items == nil {
		items = []map[string]any{}
	}
	return items, true
}
