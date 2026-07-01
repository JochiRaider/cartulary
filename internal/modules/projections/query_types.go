package projections

import (
	"strconv"
	"strings"
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

func enumSortExpr(expr string, values ...string) string {
	var builder strings.Builder
	builder.WriteString("CASE ")
	builder.WriteString(expr)
	for index, value := range values {
		builder.WriteString(" WHEN '")
		builder.WriteString(value)
		builder.WriteString("' THEN ")
		builder.WriteString(strconv.Itoa(index))
	}
	builder.WriteString(" ELSE ")
	builder.WriteString(strconv.Itoa(len(values)))
	builder.WriteString(" END")
	return builder.String()
}
