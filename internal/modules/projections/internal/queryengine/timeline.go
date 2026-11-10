package queryengine

import "github.com/JochiRaider/cartulary/internal/platform/sqlshape"

const timelineViewSchemaID = "cartulary.view.timeline.v2"

func TimelinePlans() []Surface {
	return []Surface{{
		ViewSchemaID: timelineViewSchemaID,
		FromSQL:      sqlshape.Literal(`FROM timeline_grid_projection t JOIN records r ON r.record_id = t.record_id`),
		RecordExpr:   sqlshape.Literal("t.record_id"),
		IncidentExpr: sqlshape.Literal("t.incident_id"),
		Fields: []Field{
			{Key: "timeline.date_entered_text", Expr: sqlshape.Literal("t.date_entered_text"), Kind: FieldKindText},
			{Key: "timeline.analyst_text", Expr: sqlshape.Literal("t.analyst_text"), Kind: FieldKindText},
			{Key: "timeline.mitre_stage_text", Expr: sqlshape.Literal("t.mitre_stage_text"), Kind: FieldKindText},
			{Key: "timeline.device_object_text", Expr: sqlshape.Literal("t.device_object_text"), Kind: FieldKindText},
			{Key: "timeline.ip_address_text", Expr: sqlshape.Literal("t.ip_address_text"), Kind: FieldKindText},
			{Key: "timeline.activity_utc_text", Expr: sqlshape.Literal("t.activity_utc_text"), Kind: FieldKindText},
			{Key: "timeline.activity_local_text", Expr: sqlshape.Literal("t.activity_local_text"), Kind: FieldKindText},
			{Key: "timeline.raw_activity_text", Expr: sqlshape.Literal("t.raw_activity_text"), Kind: FieldKindText},
			{Key: "timeline.activity_synopsis_text", Expr: sqlshape.Literal("t.activity_synopsis_text"), Kind: FieldKindText},
			{Key: "timeline.data_source_text", Expr: sqlshape.Literal("t.data_source_text"), Kind: FieldKindText},
			{Key: "timeline.host_refs", Expr: sqlshape.Literal("t.host_refs"), Kind: FieldKindCollection, Ordered: true},
			{Key: "timeline.identity_refs", Expr: sqlshape.Literal("t.identity_refs"), Kind: FieldKindCollection, Ordered: true},
			{Key: "timeline.tags", Expr: sqlshape.Literal("t.tags"), Kind: FieldKindCollection},
			{Key: "timeline.attached_evidence_ids", Expr: sqlshape.Literal("t.attached_evidence_refs"), Kind: FieldKindCollection},
			{Key: "timeline.evidence_count", Expr: sqlshape.Literal("t.evidence_count"), Kind: FieldKindNumber},
			{Key: "timeline.recorded_at", Expr: sqlshape.Literal("t.recorded_at"), Kind: FieldKindTimestamp},
			{Key: "timeline.edited_at", Expr: sqlshape.Literal("t.edited_at"), Kind: FieldKindTimestamp},
			{Key: "timeline.activity_sort_ts", Expr: sqlshape.Literal("t.activity_sort_ts"), Kind: FieldKindTimestamp},
			{Key: "timeline.date_entered_sort_day", Expr: sqlshape.Literal("t.date_entered_sort_day"), Kind: FieldKindDate},
			{Key: "timeline.activity_time_pair_state", Expr: sqlshape.Literal("t.activity_time_pair_state"), Kind: FieldKindText},
			{Key: "timeline.capture_state", Expr: sqlshape.Literal("t.capture_state"), Kind: FieldKindText},
			{Key: "timeline.replacement_record_id", Expr: sqlshape.Literal("t.replacement_record_id"), Kind: FieldKindText},
			{Key: "timeline.has_evidence", Expr: sqlshape.Literal("t.has_evidence"), Kind: FieldKindBool},
			{Key: "timeline.has_unresolved_mentions", Expr: sqlshape.Literal("t.has_unresolved_mentions"), Kind: FieldKindBool},
		},
	}}
}
