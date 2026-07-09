package projections

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func buildGenericQuerySQL(incidentID uuid.UUID, definition genericSurface, query viewschema.QueryMeta) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(", r.row_version")
	for _, field := range definition.fields {
		builder.WriteString(", ")
		builder.WriteString(field.expr)
	}
	builder.WriteString(" ")
	builder.WriteString(definition.fromSQL)
	builder.WriteString(" WHERE ")
	builder.WriteString(definition.incidentExpr)
	builder.WriteString(" = $1 AND r.deleted_at IS NULL")
	if definition.whereSQL != "" {
		builder.WriteString(" AND ")
		builder.WriteString(definition.whereSQL)
	}
	args := []any{incidentID}

	for _, filter := range query.Filters {
		if err := appendGenericFilter(&builder, &args, definition, filter); err != nil {
			return "", nil, err
		}
	}

	builder.WriteString(" ORDER BY ")
	for index, entry := range query.Sort {
		if index > 0 {
			builder.WriteString(", ")
		}
		field, ok := definition.field(entry.FieldKey)
		if !ok {
			return "", nil, fmt.Errorf("sort field %q not mapped for %s", entry.FieldKey, definition.viewSchemaID)
		}
		builder.WriteString(field.orderExpr())
		if entry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
		builder.WriteString(" NULLS LAST")
	}
	return builder.String(), args, nil
}

func appendGenericFilter(builder *strings.Builder, args *[]any, definition genericSurface, filter viewschema.Filter) error {
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
		return appendTagFilter(builder, args, definition.recordExpr, filter)
	}
	field, ok := definition.field(filter.FieldKey)
	if !ok {
		return fmt.Errorf("filter field %q not mapped for %s", filter.FieldKey, definition.viewSchemaID)
	}
	switch filter.Op {
	case "eq":
		return appendEqualityFilter(builder, args, field, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		value = strings.ToLower(value)
		builder.WriteString("\n   AND left(lower(coalesce((")
		builder.WriteString(field.expr)
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
		builder.WriteString("\n   AND EXISTS (SELECT 1 FROM active_record_tags_v1 rt WHERE rt.record_id = ")
		builder.WriteString(recordExpr)
		builder.WriteString(" AND rt.deleted_at IS NULL AND rt.normalized_tag_name = ANY(")
		builder.WriteString(bind(args, textValues))
		builder.WriteString("::text[]))")
		return nil
	case "contains_all":
		for _, value := range values {
			builder.WriteString("\n   AND EXISTS (SELECT 1 FROM active_record_tags_v1 rt WHERE rt.record_id = ")
			builder.WriteString(recordExpr)
			builder.WriteString(" AND rt.deleted_at IS NULL AND rt.normalized_tag_name = ")
			builder.WriteString(bind(args, value))
			builder.WriteString(")")
		}
		return nil
	default:
		return fmt.Errorf("tag filter operator %q not mapped", filter.Op)
	}
}

func appendEqualityFilter(builder *strings.Builder, args *[]any, field genericField, arg map[string]any) error {
	if value, ok := arg["value"]; ok {
		builder.WriteString("\n   AND ")
		if value == nil {
			builder.WriteString(field.expr)
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

func comparableFilterValue(field genericField, value any) any {
	if field.kind != fieldKindText {
		return value
	}
	if text, ok := value.(string); ok {
		return strings.ToLower(text)
	}
	return value
}

func appendRangeFilter(builder *strings.Builder, args *[]any, field genericField, arg map[string]any) error {
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
		builder.WriteString(field.expr)
		builder.WriteByte(' ')
		builder.WriteString(bound.op)
		builder.WriteByte(' ')
		builder.WriteString(bindWithFieldCast(args, value, field))
	}
	return nil
}

func appendComparableExpr(builder *strings.Builder, field genericField) {
	switch field.kind {
	case fieldKindText:
		builder.WriteString("lower(coalesce((")
		builder.WriteString(field.expr)
		builder.WriteString(")::text, ''))")
	default:
		builder.WriteString(field.expr)
	}
}

func bind(args *[]any, value any) string {
	*args = append(*args, value)
	return fmt.Sprintf("$%d", len(*args))
}

func bindWithFieldCast(args *[]any, value any, field genericField) string {
	placeholder := bind(args, value)
	switch field.kind {
	case fieldKindTimestamp:
		return placeholder + "::timestamptz"
	case fieldKindDate:
		return placeholder + "::date"
	default:
		return placeholder
	}
}
