package projection

import "github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"

const indicatorViewSchemaID = "cartulary.view.indicators.v1"

func QuerySurfaces() []providercontract.QuerySurface {
	return []providercontract.QuerySurface{{
		ViewSchemaID: indicatorViewSchemaID,
		FromSQL:      "FROM indicator_grid_projection i JOIN records r ON r.record_id = i.record_id",
		RecordExpr:   "i.record_id",
		IncidentExpr: "i.incident_id",
		Fields: []providercontract.QueryField{
			{Key: "indicator.indicator_type", Expr: "i.indicator_type", Kind: providercontract.FieldKindText},
			{Key: "indicator.value_kind", Expr: "i.value_kind", Kind: providercontract.FieldKindText},
			{Key: "indicator.display_value", Expr: "i.display_value", Kind: providercontract.FieldKindText},
			{Key: "indicator.normalized_value", Expr: "i.normalized_value", Kind: providercontract.FieldKindText},
			{Key: "indicator.defanged_value", Expr: "i.defanged_value", Kind: providercontract.FieldKindText},
			{Key: "indicator.hash_algorithm", Expr: "i.hash_algorithm", Kind: providercontract.FieldKindText},
			{Key: "indicator.hash_value", Expr: "i.hash_value", Kind: providercontract.FieldKindText},
			{Key: "indicator.stix_pattern", Expr: "i.stix_pattern", Kind: providercontract.FieldKindText},
			{Key: "indicator.first_observed_at", Expr: "i.first_observed_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "indicator.last_observed_at", Expr: "i.last_observed_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "indicator.observation_count", Expr: "i.observation_count", Kind: providercontract.FieldKindNumber},
			{Key: "indicator.lifecycle_summary", Expr: "i.lifecycle_summary", Kind: providercontract.FieldKindText},
			{Key: "indicator.supporting_link_count", Expr: "i.supporting_link_count", Kind: providercontract.FieldKindNumber},
		},
	}}
}
