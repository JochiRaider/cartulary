package queryengine

import "github.com/JochiRaider/cartulary/internal/platform/sqlshape"

const indicatorViewSchemaID = "cartulary.view.indicators.v1"

func IndicatorPlans() []Surface {
	return []Surface{{
		ViewSchemaID: indicatorViewSchemaID,
		FromSQL:      sqlshape.Literal("FROM indicator_grid_projection i JOIN records r ON r.record_id = i.record_id"),
		RecordExpr:   sqlshape.Literal("i.record_id"),
		IncidentExpr: sqlshape.Literal("i.incident_id"),
		Fields: []Field{
			{Key: "indicator.indicator_type", Expr: sqlshape.Literal("i.indicator_type"), Kind: FieldKindText},
			{Key: "indicator.value_kind", Expr: sqlshape.Literal("i.value_kind"), Kind: FieldKindText},
			{Key: "indicator.display_value", Expr: sqlshape.Literal("i.display_value"), Kind: FieldKindText},
			{Key: "indicator.normalized_value", Expr: sqlshape.Literal("i.normalized_value"), Kind: FieldKindText},
			{Key: "indicator.defanged_value", Expr: sqlshape.Literal("i.defanged_value"), Kind: FieldKindText},
			{Key: "indicator.hash_algorithm", Expr: sqlshape.Literal("i.hash_algorithm"), Kind: FieldKindText},
			{Key: "indicator.hash_value", Expr: sqlshape.Literal("i.hash_value"), Kind: FieldKindText},
			{Key: "indicator.stix_pattern", Expr: sqlshape.Literal("i.stix_pattern"), Kind: FieldKindText},
			{Key: "indicator.first_observed_at", Expr: sqlshape.Literal("i.first_observed_at"), Kind: FieldKindTimestamp},
			{Key: "indicator.last_observed_at", Expr: sqlshape.Literal("i.last_observed_at"), Kind: FieldKindTimestamp},
			{Key: "indicator.observation_count", Expr: sqlshape.Literal("i.observation_count"), Kind: FieldKindNumber},
			{Key: "indicator.lifecycle_summary", Expr: sqlshape.Literal("i.lifecycle_summary"), Kind: FieldKindText},
			{Key: "indicator.supporting_link_count", Expr: sqlshape.Literal("i.supporting_link_count"), Kind: FieldKindNumber},
		},
	}}
}
