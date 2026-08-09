package runtime

import (
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type fieldKind string

const (
	fieldKindText       fieldKind = "text"
	fieldKindTimestamp  fieldKind = "timestamp"
	fieldKindDate       fieldKind = "date"
	fieldKindBool       fieldKind = "bool"
	fieldKindNumber     fieldKind = "number"
	fieldKindCollection fieldKind = "collection"
)

type genericField struct {
	key      string
	expr     string
	sortExpr string
	kind     fieldKind
	ordered  bool
}

type genericSurface struct {
	viewSchemaID   string
	fromSQL        string
	recordExpr     string
	incidentExpr   string
	whereSQL       string
	fields         []genericField
	groupingFields []string
}

func (d genericSurface) field(key string) (genericField, bool) {
	if key == "record_id" {
		return genericField{key: "record_id", expr: d.recordExpr, kind: fieldKindText}, true
	}
	for _, field := range d.fields {
		if field.key == key {
			return field, true
		}
	}
	return genericField{}, false
}

func (d genericSurface) queryEngineSurface() queryengine.Surface {
	fields := make([]queryengine.Field, 0, len(d.fields))
	for _, field := range d.fields {
		fields = append(fields, queryengine.Field{
			Key:      field.key,
			Expr:     field.expr,
			SortExpr: field.sortExpr,
			Kind:     queryengine.FieldKind(field.kind),
			Ordered:  field.ordered,
		})
	}
	return queryengine.Surface{
		ViewSchemaID:   d.viewSchemaID,
		FromSQL:        d.fromSQL,
		RecordExpr:     d.recordExpr,
		IncidentExpr:   d.incidentExpr,
		WhereSQL:       d.whereSQL,
		Fields:         fields,
		GroupingFields: append([]string(nil), d.groupingFields...),
	}
}

func genericSurfaceFromPlan(surface queryengine.Surface) (genericSurface, error) {
	if surface.ViewSchemaID == "" {
		return genericSurface{}, fmt.Errorf("empty view_schema_id")
	}
	if surface.FromSQL == "" {
		return genericSurface{}, fmt.Errorf("%s has empty from_sql", surface.ViewSchemaID)
	}
	if surface.RecordExpr == "" {
		return genericSurface{}, fmt.Errorf("%s has empty record_expr", surface.ViewSchemaID)
	}
	if surface.IncidentExpr == "" {
		return genericSurface{}, fmt.Errorf("%s has empty incident_expr", surface.ViewSchemaID)
	}
	for member, fragment := range map[string]string{
		"from_sql":      surface.FromSQL,
		"record_expr":   surface.RecordExpr,
		"incident_expr": surface.IncidentExpr,
		"where_sql":     surface.WhereSQL,
	} {
		if err := validateCompiledSQLFragment(surface.ViewSchemaID, member, fragment, member == "where_sql"); err != nil {
			return genericSurface{}, err
		}
	}
	fields := make([]genericField, 0, len(surface.Fields))
	seenFieldKeys := make(map[string]struct{}, len(surface.Fields))
	for _, field := range surface.Fields {
		converted, err := genericFieldFromPlan(surface.ViewSchemaID, field)
		if err != nil {
			return genericSurface{}, err
		}
		if _, exists := seenFieldKeys[converted.key]; exists {
			return genericSurface{}, fmt.Errorf("%s declares duplicate query field %s", surface.ViewSchemaID, converted.key)
		}
		seenFieldKeys[converted.key] = struct{}{}
		fields = append(fields, converted)
	}
	if len(fields) == 0 {
		return genericSurface{}, fmt.Errorf("%s has no query fields", surface.ViewSchemaID)
	}
	schema, ok := viewschema.Lookup(surface.ViewSchemaID)
	if !ok {
		return genericSurface{}, fmt.Errorf("%s has no registered view schema", surface.ViewSchemaID)
	}
	schemaFields := schema.Fields()
	for fieldKey := range seenFieldKeys {
		if _, exists := schemaFields[fieldKey]; !exists {
			return genericSurface{}, fmt.Errorf("%s maps unknown schema field %s", surface.ViewSchemaID, fieldKey)
		}
	}
	for fieldKey := range schemaFields {
		if _, exists := seenFieldKeys[fieldKey]; !exists {
			return genericSurface{}, fmt.Errorf("%s does not map schema field %s", surface.ViewSchemaID, fieldKey)
		}
	}
	groupingFields := schema.GroupingFields()
	for _, fieldKey := range groupingFields {
		if _, ok := surfaceField(fields, fieldKey); !ok {
			return genericSurface{}, fmt.Errorf("%s grouping field %s is not mapped", surface.ViewSchemaID, fieldKey)
		}
	}
	return genericSurface{
		viewSchemaID:   surface.ViewSchemaID,
		fromSQL:        surface.FromSQL,
		recordExpr:     surface.RecordExpr,
		incidentExpr:   surface.IncidentExpr,
		whereSQL:       surface.WhereSQL,
		fields:         fields,
		groupingFields: groupingFields,
	}, nil
}

func surfaceField(fields []genericField, key string) (genericField, bool) {
	for _, field := range fields {
		if field.key == key {
			return field, true
		}
	}
	return genericField{}, false
}

func genericFieldFromPlan(viewSchemaID string, field queryengine.Field) (genericField, error) {
	if field.Key == "" {
		return genericField{}, fmt.Errorf("%s declares query field with empty key", viewSchemaID)
	}
	if field.Expr == "" {
		return genericField{}, fmt.Errorf("%s query field %s has empty expr", viewSchemaID, field.Key)
	}
	if err := validateCompiledSQLFragment(viewSchemaID, "field "+field.Key+" expr", field.Expr, false); err != nil {
		return genericField{}, err
	}
	if err := validateCompiledSQLFragment(viewSchemaID, "field "+field.Key+" sort_expr", field.SortExpr, true); err != nil {
		return genericField{}, err
	}
	kind, err := fieldKindFromPlan(field.Kind)
	if err != nil {
		return genericField{}, fmt.Errorf("%s query field %s: %w", viewSchemaID, field.Key, err)
	}
	return genericField{
		key:      field.Key,
		expr:     field.Expr,
		sortExpr: field.SortExpr,
		kind:     kind,
		ordered:  field.Ordered,
	}, nil
}

func validateCompiledSQLFragment(viewSchemaID string, member string, value string, allowEmpty bool) error {
	trimmed := strings.TrimSpace(value)
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

func fieldKindFromPlan(kind queryengine.FieldKind) (fieldKind, error) {
	switch kind {
	case queryengine.FieldKindText:
		return fieldKindText, nil
	case queryengine.FieldKindTimestamp:
		return fieldKindTimestamp, nil
	case queryengine.FieldKindDate:
		return fieldKindDate, nil
	case queryengine.FieldKindBool:
		return fieldKindBool, nil
	case queryengine.FieldKindNumber:
		return fieldKindNumber, nil
	case queryengine.FieldKindCollection:
		return fieldKindCollection, nil
	default:
		return "", fmt.Errorf("unsupported field kind %q", kind)
	}
}
