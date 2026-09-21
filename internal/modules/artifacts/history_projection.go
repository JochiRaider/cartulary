package artifacts

import (
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
)

func artifactHistoryProjector(catalog *sourcecatalog.Catalog) historycontract.Projector {
	return func(facts historycontract.Facts) ([]historycontract.Unit, error) {
		source := historycontract.Source(facts.After)
		if source == nil {
			source = historycontract.Source(facts.Before)
		}
		artifactType, _ := source["artifact_type"].(string)
		surface, ok := catalog.SurfaceByArtifactType(artifactType)
		if !ok {
			return nil, historycontract.ErrInvalidFacts
		}
		for _, snapshot := range []map[string]any{facts.Before, facts.After} {
			if snapshot != nil && historycontract.Source(snapshot)["artifact_type"] != artifactType {
				return nil, historycontract.ErrInvalidFacts
			}
		}
		fields := []historycontract.Field{{Key: "artifact.artifact_type", Member: "artifact_type", Type: "text", Required: true}}
		prefix := ""
		for _, field := range catalog.Fields() {
			if field.ViewSchemaID != surface.ViewSchemaID || field.Kind != sourcecatalog.FieldKindDirect {
				continue
			}
			valueType := field.View.ReadKind
			if valueType == "enum" || valueType == "timestamp" {
				valueType = "text"
			}
			fields = append(fields, historycontract.Field{Key: field.FieldKey, Member: field.Storage.Column, Type: valueType, Kind: "field"})
			prefix = strings.SplitN(field.FieldKey, ".", 2)[0]
		}
		// Lifecycle/assigned identifiers are source facts even when read-only in
		// the workbook. They are not inferred from the editable field catalog.
		for _, member := range map[string][]string{
			"finding": {"closed_at"}, "forensic_keyword": {"keyword_id"},
			"investigative_query": {"query_id"}, "comm_log": {"comm_id"},
			"handoff":       {"handoff_id", "acknowledged_at"},
			"status_review": {"status_review_id"}, "lesson": {"lesson_id"},
		}[prefix] {
			duplicate := false
			for _, field := range fields {
				if field.Member == member {
					duplicate = true
				}
			}
			if !duplicate {
				fields = append(fields, historycontract.Field{Key: prefix + "." + member, Member: member, Type: "text", Kind: "field"})
			}
		}
		return historycontract.Row(facts, fields)
	}
}
