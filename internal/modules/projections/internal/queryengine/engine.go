package queryengine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
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
	Expr     sqlshape.Fragment
	SortExpr sqlshape.Fragment
	Kind     FieldKind
	Ordered  bool
}

type Surface struct {
	ViewSchemaID   string
	FromSQL        sqlshape.Fragment
	RecordExpr     sqlshape.Fragment
	IncidentExpr   sqlshape.Fragment
	WhereSQL       sqlshape.Fragment
	Fields         []Field
	GroupingFields []string
}

// CompileSurface validates one private executable plan and returns a defensive
// copy whose grouping fields are derived from the canonical view schema.
func CompileSurface(surface Surface) (Surface, error) {
	if strings.TrimSpace(surface.ViewSchemaID) == "" {
		return Surface{}, fmt.Errorf("empty view_schema_id")
	}
	for member, fragment := range map[string]sqlshape.Fragment{
		"from_sql":      surface.FromSQL,
		"record_expr":   surface.RecordExpr,
		"incident_expr": surface.IncidentExpr,
		"where_sql":     surface.WhereSQL,
	} {
		if err := validateCompiledSQLFragment(surface.ViewSchemaID, member, fragment, member == "where_sql"); err != nil {
			return Surface{}, err
		}
	}
	fields := make([]Field, 0, len(surface.Fields))
	seenFieldKeys := make(map[string]struct{}, len(surface.Fields))
	for _, field := range surface.Fields {
		if strings.TrimSpace(field.Key) == "" {
			return Surface{}, fmt.Errorf("%s declares query field with empty key", surface.ViewSchemaID)
		}
		if err := validateCompiledSQLFragment(surface.ViewSchemaID, "field "+field.Key+" expr", field.Expr, false); err != nil {
			return Surface{}, err
		}
		if err := validateCompiledSQLFragment(surface.ViewSchemaID, "field "+field.Key+" sort_expr", field.SortExpr, true); err != nil {
			return Surface{}, err
		}
		if err := validateFieldKind(field.Kind); err != nil {
			return Surface{}, fmt.Errorf("%s query field %s: %w", surface.ViewSchemaID, field.Key, err)
		}
		if _, exists := seenFieldKeys[field.Key]; exists {
			return Surface{}, fmt.Errorf("%s declares duplicate query field %s", surface.ViewSchemaID, field.Key)
		}
		seenFieldKeys[field.Key] = struct{}{}
		fields = append(fields, field)
	}
	if len(fields) == 0 {
		return Surface{}, fmt.Errorf("%s has no query fields", surface.ViewSchemaID)
	}
	schema, ok := viewschema.Lookup(surface.ViewSchemaID)
	if !ok {
		return Surface{}, fmt.Errorf("%s has no registered view schema", surface.ViewSchemaID)
	}
	schemaFields := schema.Fields()
	for fieldKey := range seenFieldKeys {
		if _, exists := schemaFields[fieldKey]; !exists {
			return Surface{}, fmt.Errorf("%s maps unknown schema field %s", surface.ViewSchemaID, fieldKey)
		}
	}
	for fieldKey := range schemaFields {
		if _, exists := seenFieldKeys[fieldKey]; !exists {
			return Surface{}, fmt.Errorf("%s does not map schema field %s", surface.ViewSchemaID, fieldKey)
		}
	}
	groupingFields := schema.GroupingFields()
	for _, fieldKey := range groupingFields {
		if _, exists := seenFieldKeys[fieldKey]; !exists {
			return Surface{}, fmt.Errorf("%s grouping field %s is not mapped", surface.ViewSchemaID, fieldKey)
		}
	}
	return Surface{
		ViewSchemaID:   surface.ViewSchemaID,
		FromSQL:        surface.FromSQL,
		RecordExpr:     surface.RecordExpr,
		IncidentExpr:   surface.IncidentExpr,
		WhereSQL:       surface.WhereSQL,
		Fields:         fields,
		GroupingFields: append([]string(nil), groupingFields...),
	}, nil
}

func validateCompiledSQLFragment(viewSchemaID string, member string, value sqlshape.Fragment, allowEmpty bool) error {
	trimmed := strings.TrimSpace(sqlshape.Text(value))
	if trimmed == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("%s has empty %s", viewSchemaID, member)
	}
	for _, forbidden := range []string{"\x00", ";", "--", "/*", "*/", "$"} {
		if strings.Contains(trimmed, forbidden) {
			return fmt.Errorf("%s %s contains forbidden SQL token %q", viewSchemaID, member, forbidden)
		}
	}
	return nil
}

func validateFieldKind(kind FieldKind) error {
	switch kind {
	case FieldKindText, FieldKindTimestamp, FieldKindDate, FieldKindBool, FieldKindNumber, FieldKindCollection:
		return nil
	default:
		return fmt.Errorf("unsupported field kind %q", kind)
	}
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

func (f Field) OrderExpr() sqlshape.Fragment {
	if !f.SortExpr.Empty() {
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

func LoadRowTx(
	ctx context.Context,
	tx pgx.Tx,
	definition Surface,
	recordID uuid.UUID,
) (map[string]any, error) {
	var builder sqlshape.Builder
	builder.Write(sqlshape.Literal("SELECT "))
	builder.Write(definition.RecordExpr)
	builder.Write(sqlshape.Literal(", r.row_version"))
	for _, field := range definition.Fields {
		builder.Write(sqlshape.Literal(", "))
		builder.Write(field.Expr)
	}
	builder.Write(sqlshape.Literal(" "))
	builder.Write(definition.FromSQL)
	builder.Write(sqlshape.Literal(" WHERE "))
	builder.Write(definition.RecordExpr)
	builder.Write(sqlshape.Literal(" = $1 AND r.deleted_at IS NULL"))
	if !definition.WhereSQL.Empty() {
		builder.Write(sqlshape.Literal(" AND "))
		builder.Write(definition.WhereSQL)
	}
	row := tx.QueryRow(ctx, sqlshape.Text(builder.Fragment()), recordID)
	values := make([]any, len(definition.Fields)+2)
	scanTargets := make([]any, len(values))
	for index := range values {
		scanTargets[index] = &values[index]
	}
	if err := row.Scan(scanTargets...); err != nil {
		return nil, err
	}
	return BuildRow(definition, values)
}

func BuildQueryPageSQL(incidentID uuid.UUID, definition Surface, query viewschema.QueryMeta, window querypage.Window) (sqlshape.Fragment, []any, error) {
	var builder sqlshape.Builder
	builder.Write(sqlshape.Literal("SELECT "))
	builder.Write(definition.RecordExpr)
	builder.Write(sqlshape.Literal(", r.row_version"))
	for _, field := range definition.Fields {
		builder.Write(sqlshape.Literal(", "))
		builder.Write(field.Expr)
	}
	builder.Write(sqlshape.Literal(" "))
	builder.Write(definition.FromSQL)
	builder.Write(sqlshape.Literal(" WHERE "))
	builder.Write(definition.IncidentExpr)
	builder.Write(sqlshape.Literal(" = $1 AND r.deleted_at IS NULL"))
	if !definition.WhereSQL.Empty() {
		builder.Write(sqlshape.Literal(" AND "))
		builder.Write(definition.WhereSQL)
	}
	args := []any{incidentID}
	if window.TargetRecordID != "" {
		args = append(args, window.TargetRecordID)
		builder.Write(sqlshape.Format(" AND %s = $2::uuid", definition.RecordExpr))
	}

	for _, filter := range query.Filters {
		if err := appendGenericFilter(&builder, &args, definition, filter); err != nil {
			return sqlshape.Fragment{}, nil, err
		}
	}
	pageFields := make(map[string]querypage.Field, len(definition.Fields)+1)
	pageFields["record_id"] = querypage.Field{Expression: definition.RecordExpr, Cast: sqlshape.Literal("uuid")}
	for _, field := range definition.Fields {
		var cast sqlshape.Fragment
		switch field.Kind {
		case FieldKindTimestamp:
			cast = sqlshape.Literal("timestamptz")
		case FieldKindDate:
			cast = sqlshape.Literal("date")
		case FieldKindNumber:
			cast = sqlshape.Literal("numeric")
		case FieldKindBool:
			cast = sqlshape.Literal("boolean")
		}
		pageFields[field.Key] = querypage.Field{Expression: field.OrderExpr(), Cast: cast}
	}
	appendPosition := querypage.AppendKeyset
	if window.Before {
		appendPosition = querypage.AppendBefore
	}
	if err := appendPosition(&builder, &args, query.Sort, pageFields, window.Position); err != nil {
		return sqlshape.Fragment{}, nil, err
	}

	builder.Write(sqlshape.Literal(" ORDER BY "))
	for index, entry := range query.Sort {
		if index > 0 {
			builder.Write(sqlshape.Literal(", "))
		}
		field, ok := definition.Field(entry.FieldKey)
		if !ok {
			return sqlshape.Fragment{}, nil, fmt.Errorf("sort field %q not mapped for %s", entry.FieldKey, definition.ViewSchemaID)
		}
		builder.Write(field.OrderExpr())
		if (entry.Direction == "desc") != window.Before {
			builder.Write(sqlshape.Literal(" DESC"))
		} else {
			builder.Write(sqlshape.Literal(" ASC"))
		}
		if window.Before {
			builder.Write(sqlshape.Literal(" NULLS FIRST"))
		} else {
			builder.Write(sqlshape.Literal(" NULLS LAST"))
		}
	}
	if err := querypage.AppendLimit(&builder, &args, window.Limit); err != nil {
		return sqlshape.Fragment{}, nil, err
	}
	return builder.Fragment(), args, nil
}

func appendGenericFilter(builder *sqlshape.Builder, args *[]any, definition Surface, filter viewschema.Filter) error {
	if filter.FieldKey == "note.full_text" {
		query, _ := filter.Arg["query"].(string)
		for _, token := range strings.Fields(query) {
			builder.Write(sqlshape.Literal("\n   AND "))
			builder.Write(bind(args, token))
			builder.Write(sqlshape.Literal(" = ANY(regexp_split_to_array(lower(coalesce(p.title, '') || ' ' || coalesce(p.body, '')), '[^[:alnum:]]+'))"))
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
		builder.Write(sqlshape.Literal("\n   AND left(lower(coalesce(("))
		builder.Write(field.Expr)
		builder.Write(sqlshape.Literal(")::text, '')), char_length("))
		builder.Write(bind(args, value))
		builder.Write(sqlshape.Literal(")) = "))
		builder.Write(bind(args, value))
		return nil
	case "range":
		return appendRangeFilter(builder, args, field, filter.Arg)
	default:
		return fmt.Errorf("filter operator %q not mapped for %s", filter.Op, filter.FieldKey)
	}
}

func appendTagFilter(builder *sqlshape.Builder, args *[]any, recordExpr sqlshape.Fragment, filter viewschema.Filter) error {
	values, _ := filter.Arg["values"].([]any)
	textValues := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			textValues = append(textValues, text)
		}
	}
	switch filter.Op {
	case "contains_any":
		builder.Write(sqlshape.Format("\n   AND EXISTS (SELECT 1 FROM active_record_tags_v1 rt WHERE rt.record_id = %s AND rt.normalized_tag_name = ANY(%s::text[]))", recordExpr, bind(args, textValues)))
		return nil
	case "contains_all":
		for _, value := range values {
			builder.Write(sqlshape.Format("\n   AND EXISTS (SELECT 1 FROM active_record_tags_v1 rt WHERE rt.record_id = %s AND rt.normalized_tag_name = %s)", recordExpr, bind(args, value)))
		}
		return nil
	default:
		return fmt.Errorf("tag filter operator %q not mapped", filter.Op)
	}
}

func appendEqualityFilter(builder *sqlshape.Builder, args *[]any, field Field, arg map[string]any) error {
	if value, ok := arg["value"]; ok {
		builder.Write(sqlshape.Literal("\n   AND "))
		if value == nil {
			builder.Write(field.Expr)
			builder.Write(sqlshape.Literal(" IS NULL"))
			return nil
		}
		appendComparableExpr(builder, field)
		builder.Write(sqlshape.Literal(" = "))
		builder.Write(bindWithFieldCast(args, comparableFilterValue(field, value), field))
		return nil
	}
	values, _ := arg["values"].([]any)
	builder.Write(sqlshape.Literal("\n   AND ("))
	for index, value := range values {
		if index > 0 {
			builder.Write(sqlshape.Literal(" OR "))
		}
		appendComparableExpr(builder, field)
		builder.Write(sqlshape.Literal(" = "))
		builder.Write(bindWithFieldCast(args, comparableFilterValue(field, value), field))
	}
	builder.Write(sqlshape.Literal(")"))
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

func appendRangeFilter(builder *sqlshape.Builder, args *[]any, field Field, arg map[string]any) error {
	for _, bound := range []struct {
		key string
		op  sqlshape.Fragment
	}{
		{key: "gt", op: sqlshape.Literal(">")},
		{key: "gte", op: sqlshape.Literal(">=")},
		{key: "lt", op: sqlshape.Literal("<")},
		{key: "lte", op: sqlshape.Literal("<=")},
	} {
		value, ok := arg[bound.key]
		if !ok {
			continue
		}
		builder.Write(sqlshape.Literal("\n   AND "))
		builder.Write(field.Expr)
		builder.Write(sqlshape.Literal(" "))
		builder.Write(bound.op)
		builder.Write(sqlshape.Literal(" "))
		builder.Write(bindWithFieldCast(args, value, field))
	}
	return nil
}

func appendComparableExpr(builder *sqlshape.Builder, field Field) {
	switch field.Kind {
	case FieldKindText:
		builder.Write(sqlshape.Literal("lower(coalesce(("))
		builder.Write(field.Expr)
		builder.Write(sqlshape.Literal(")::text, ''))"))
	default:
		builder.Write(field.Expr)
	}
}

func bind(args *[]any, value any) sqlshape.Fragment {
	*args = append(*args, value)
	return sqlshape.Parameter(len(*args))
}

func bindWithFieldCast(args *[]any, value any, field Field) sqlshape.Fragment {
	placeholder := bind(args, value)
	switch field.Kind {
	case FieldKindTimestamp:
		return sqlshape.Concat(placeholder, sqlshape.Literal("::timestamptz"))
	case FieldKindDate:
		return sqlshape.Concat(placeholder, sqlshape.Literal("::date"))
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
