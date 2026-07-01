package projections

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
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
	viewSchemaID string
	fromSQL      string
	recordExpr   string
	incidentExpr string
	whereSQL     string
	fields       []genericField
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

func (f genericField) orderExpr() string {
	if f.sortExpr != "" {
		return f.sortExpr
	}
	return f.expr
}

func genericSurfaceFromContract(surface providercontract.QuerySurface) (genericSurface, error) {
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
	fields := make([]genericField, 0, len(surface.Fields))
	for _, field := range surface.Fields {
		converted, err := genericFieldFromContract(surface.ViewSchemaID, field)
		if err != nil {
			return genericSurface{}, err
		}
		fields = append(fields, converted)
	}
	if len(fields) == 0 {
		return genericSurface{}, fmt.Errorf("%s has no query fields", surface.ViewSchemaID)
	}
	return genericSurface{
		viewSchemaID: surface.ViewSchemaID,
		fromSQL:      surface.FromSQL,
		recordExpr:   surface.RecordExpr,
		incidentExpr: surface.IncidentExpr,
		whereSQL:     surface.WhereSQL,
		fields:       fields,
	}, nil
}

func genericFieldFromContract(viewSchemaID string, field providercontract.QueryField) (genericField, error) {
	if field.Key == "" {
		return genericField{}, fmt.Errorf("%s declares query field with empty key", viewSchemaID)
	}
	if field.Expr == "" {
		return genericField{}, fmt.Errorf("%s query field %s has empty expr", viewSchemaID, field.Key)
	}
	kind, err := fieldKindFromContract(field.Kind)
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

func fieldKindFromContract(kind providercontract.FieldKind) (fieldKind, error) {
	switch kind {
	case providercontract.FieldKindText:
		return fieldKindText, nil
	case providercontract.FieldKindTimestamp:
		return fieldKindTimestamp, nil
	case providercontract.FieldKindDate:
		return fieldKindDate, nil
	case providercontract.FieldKindBool:
		return fieldKindBool, nil
	case providercontract.FieldKindNumber:
		return fieldKindNumber, nil
	case providercontract.FieldKindCollection:
		return fieldKindCollection, nil
	default:
		return "", fmt.Errorf("unsupported field kind %q", kind)
	}
}
