package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func taskDecisionHistoryProjector(catalog *sourcecatalog.Catalog, viewSchemaID string) historycontract.Projector {
	fields := []historycontract.Field{}
	for _, field := range catalog.Fields() {
		if field.ViewSchemaID != viewSchemaID || field.Kind != sourcecatalog.FieldKindDirect {
			continue
		}
		valueType := field.View.ReadKind
		if valueType == "enum" || valueType == "timestamp" {
			valueType = "text"
		}
		fields = append(fields, historycontract.Field{Key: field.FieldKey, Member: field.Storage.Column, Type: valueType, Kind: "field"})
	}
	return func(facts historycontract.Facts) ([]historycontract.Unit, error) {
		return historycontract.Row(facts, fields)
	}
}
